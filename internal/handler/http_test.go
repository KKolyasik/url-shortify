package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockShortener struct {
	shortenFn func(raw string) (string, error)
	resolveFn func(id string) (string, error)

	shortenCalls  int
	resolveCalls  int
	lastShortenIn string
	lastResolveIn string
}

func (m *mockShortener) Shorten(raw string) (string, error) {
	m.shortenCalls++
	m.lastShortenIn = raw
	if m.shortenFn == nil {
		return "", nil
	}
	return m.shortenFn(raw)
}

func (m *mockShortener) Resolve(id string) (string, error) {
	m.resolveCalls++
	m.lastResolveIn = id
	if m.resolveFn == nil {
		return "", nil
	}
	return m.resolveFn(id)
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) { return 0, errors.New("read error") }

func TestHandler_Shortify(t *testing.T) {
	const (
		baseURL = "test"
		path    = "/"
	)

	type tc struct {
		name string

		method      string
		contentType string
		body        io.Reader

		shortenFn func(raw string) (string, error)

		wantCode              int
		wantBodyContains      string
		wantContentTypePrefix string

		wantShortenCalls int
		wantShortenIn    string
	}

	tests := []tc{
		{
			name:        "success: returns 201, body = baseURL/id, content-type set",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        bytes.NewReader([]byte("http://example.com/")),
			shortenFn: func(raw string) (string, error) {
				return "example", nil
			},
			wantCode:              http.StatusCreated,
			wantBodyContains:      "test/example",
			wantContentTypePrefix: "text/plain",
			wantShortenCalls:      1,
			wantShortenIn:         "http://example.com/",
		},
		{
			name:                  "bad method -> 400, shorten not called",
			method:                http.MethodGet,
			contentType:           "text/plain",
			body:                  bytes.NewReader([]byte("http://example.com/")),
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "Invalid request method",
			wantShortenCalls:      0,
			wantContentTypePrefix: "text/plain", // http.Error тоже ставит text/plain; charset=utf-8
		},
		{
			name:                  "missing content-type -> 400, shorten not called",
			method:                http.MethodPost,
			contentType:           "",
			body:                  bytes.NewReader([]byte("http://example.com/")),
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "Content-Type must be text/plain",
			wantShortenCalls:      0,
			wantContentTypePrefix: "text/plain",
		},
		{
			name:                  "wrong content-type -> 400, shorten not called",
			method:                http.MethodPost,
			contentType:           "application/json",
			body:                  bytes.NewReader([]byte(`"http://example.com/"`)),
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "Content-Type must be text/plain",
			wantShortenCalls:      0,
			wantContentTypePrefix: "text/plain",
		},
		{
			name:                  "body read error -> 400, shorten not called",
			method:                http.MethodPost,
			contentType:           "text/plain",
			body:                  errReader{},
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "Failed to read body",
			wantShortenCalls:      0,
			wantContentTypePrefix: "text/plain",
		},
		{
			name:        "body too large (> 8KB) -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "text/plain",
			// 8<<10 = 8192, сделаем 8193 байта
			body:                  bytes.NewReader(bytes.Repeat([]byte("a"), (8<<10)+1)),
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "Failed to read body",
			wantShortenCalls:      0,
			wantContentTypePrefix: "text/plain",
		},
		{
			name:        "shortener error -> 400, error message returned",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        bytes.NewReader([]byte("http://example.com/")),
			shortenFn: func(raw string) (string, error) {
				return "", errors.New("boom")
			},
			wantCode:              http.StatusBadRequest,
			wantBodyContains:      "boom",
			wantContentTypePrefix: "text/plain",
			wantShortenCalls:      1,
			wantShortenIn:         "http://example.com/",
		},
		{
			name:        "content-type with charset in request is accepted",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        bytes.NewReader([]byte("http://example.com/")),
			shortenFn: func(raw string) (string, error) {
				return "ok", nil
			},
			wantCode:              http.StatusCreated,
			wantBodyContains:      "test/ok",
			wantContentTypePrefix: "text/plain",
			wantShortenCalls:      1,
			wantShortenIn:         "http://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockShortener{shortenFn: tt.shortenFn}
			h := New(baseURL, m)

			r := httptest.NewRequest(tt.method, path, tt.body)
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			h.Shortify(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			b, _ := io.ReadAll(resp.Body)
			bodyStr := string(b)

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			if tt.wantContentTypePrefix != "" {
				assert.True(t,
					strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), strings.ToLower(tt.wantContentTypePrefix)),
					"unexpected content-type: %q", resp.Header.Get("Content-Type"),
				)
			}

			if tt.wantBodyContains != "" {
				assert.Contains(t, bodyStr, tt.wantBodyContains)
			}

			assert.Equal(t, tt.wantShortenCalls, m.shortenCalls)
			if tt.wantShortenIn != "" {
				assert.Equal(t, tt.wantShortenIn, m.lastShortenIn)
			}
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	const baseURL = "test"

	type tc struct {
		name string

		method string
		path   string

		resolveFn func(id string) (string, error)

		wantCode        int
		wantLocation    string
		wantBodyContain string

		wantResolveCalls int
		wantResolveIn    string
	}

	tests := []tc{
		{
			name:   "success: 307 with Location header",
			method: http.MethodGet,
			path:   "/example",
			resolveFn: func(id string) (string, error) {
				return "http://example.com/", nil
			},
			wantCode:         http.StatusTemporaryRedirect, // 307
			wantLocation:     "http://example.com/",
			wantResolveCalls: 1,
			wantResolveIn:    "example",
		},
		{
			name:             "bad method -> 400, resolve not called",
			method:           http.MethodPost,
			path:             "/example",
			wantCode:         http.StatusBadRequest,
			wantBodyContain:  "Invalid request method",
			wantResolveCalls: 0,
		},
		{
			name:             "empty id (path=/) -> 404, resolve not called",
			method:           http.MethodGet,
			path:             "/",
			wantCode:         http.StatusNotFound,
			wantResolveCalls: 0,
		},
		{
			name:   "resolve error -> 404",
			method: http.MethodGet,
			path:   "/missing",
			resolveFn: func(id string) (string, error) {
				return "", errors.New("not found")
			},
			wantCode:         http.StatusNotFound,
			wantResolveCalls: 1,
			wantResolveIn:    "missing",
		},
		{
			name:   "path with slashes: uses TrimPrefix only (id includes rest)",
			method: http.MethodGet,
			path:   "/a/b",
			resolveFn: func(id string) (string, error) {
				return "http://example.com/a/b", nil
			},
			wantCode:         http.StatusTemporaryRedirect,
			wantLocation:     "http://example.com/a/b",
			wantResolveCalls: 1,
			wantResolveIn:    "a/b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockShortener{
				resolveFn: tt.resolveFn,
			}
			h := New(baseURL, m)

			r := httptest.NewRequest(tt.method, tt.path, nil)

			// Важно: чтобы не следовать редиректу автоматически, мы вызываем handler напрямую.
			w := httptest.NewRecorder()
			h.Redirect(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)

			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, resp.Header.Get("Location"))
			}

			if tt.wantBodyContain != "" {
				body := w.Body.String()
				assert.Contains(t, body, tt.wantBodyContain)
			}

			assert.Equal(t, tt.wantResolveCalls, m.resolveCalls)
			if tt.wantResolveIn != "" {
				assert.Equal(t, tt.wantResolveIn, m.lastResolveIn)
			}
		})
	}
}
