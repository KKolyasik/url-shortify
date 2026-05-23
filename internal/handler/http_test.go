package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/KKolyasik/url-shortify/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type BodyErr struct {
	msg string
}

func (b *BodyErr) Read(p []byte) (int, error) {
	return 0, errors.New(b.msg)
}

type mockShortener struct {
	shortenFn  func(raw string) (string, error)
	resolveFn  func(id string) (string, error)
	userURLsFn func(vid uuid.UUID) ([]model.URL, error)
}

type mockAuditor struct{}

func (mockAuditor) Notify(_ context.Context, _ model.AuditEvent) {}

func (m *mockShortener) Shorten(ctx context.Context, raw string, vid uuid.UUID) (string, error) {
	if m.shortenFn == nil {
		return "", nil
	}
	return m.shortenFn(raw)
}

func (m *mockShortener) Resolve(ctx context.Context, id string) (string, error) {
	if m.resolveFn == nil {
		return "", nil
	}
	return m.resolveFn(id)
}

func (m *mockShortener) UserResolve(ctx context.Context, vid uuid.UUID) ([]model.URL, error) {
	if m.userURLsFn == nil {
		return nil, nil
	}

	return m.userURLsFn(vid)
}

func (m *mockShortener) Delete(ctx context.Context, vid uuid.UUID, shortCodes ...string) error {
	return nil
}

func marshalRequestBody(url string) string {
	request := model.URLRequest{URL: url}
	r, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	return string(r)
}

func marshalResponseBody(result string) string {
	buf := bytes.NewBuffer([]byte{})
	response := model.URLResponse{Result: result}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(response); err != nil {
		panic(err)
	}
	return buf.String()
}

func marshalUserURLsBody(urls []model.UserURLsResponse) string {
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)
	if err := enc.Encode(urls); err != nil {
		panic(err)
	}
	return buf.String()
}

func withVisitorID(r *http.Request, vid uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), model.VisitorIDKey, vid))
}

func TestHandler_Shortify(t *testing.T) {
	const (
		baseURL = "http://localhost:8080"
		path    = "/"
	)
	type want struct {
		code        int
		contentType string
		body        string
	}
	tests := []struct {
		name        string
		method      string
		contentType string
		body        io.Reader
		mock        mockShortener
		want        want
	}{
		{
			name:        "success: returns 201",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader("http://example.com"),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "abc", nil
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				body:        baseURL + "/" + "abc",
			},
		},
		{
			name:   "bad method -> 405, shorten not called",
			method: http.MethodGet,
			body:   strings.NewReader("http://example.com"),
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				body:        "Invalid request method\n",
			},
		},
		{
			name:        "missing content-type -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "",
			body:        strings.NewReader("http://example.com"),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Content-Type must be text/plain\n",
			},
		},
		{
			name:        "wrong content-type -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader("http://example.com"),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Content-Type must be text/plain\n",
			},
		},
		{
			name:        "body read error -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        &BodyErr{msg: "read error"},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Failed to read body\n",
			},
		},
		{
			name:        "body too large 400, shorten not called",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader(strings.Repeat("a", 8<<10+1)),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Failed to read body\n",
			},
		},
		{
			name:        "shortener error -> 400, error message returned",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader("http://example.com"),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "", errors.New("Shorten error")
				},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Shorten error\n",
			},
		},
		{
			name:        "duplicate URL -> 409 with existing short URL",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader("http://example.com"),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "", &domainerr.URLAlreadyExistsError{ShortCode: "abc"}
				},
			},
			want: want{
				code:        http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
				body:        baseURL + "/" + "abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(baseURL, &tt.mock, mockAuditor{})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, path, tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			r = withVisitorID(r, uuid.New())

			handler.Shortify(w, r)
			response := w.Result()

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.code, response.StatusCode)
			assert.True(t, strings.EqualFold(tt.want.contentType, response.Header.Get("Content-Type")))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	const (
		baseURL = "http://localhost:8080"
		path    = "/abc"
	)
	type want struct {
		code        int
		contentType string
		body        string
		location    string
	}

	tests := []struct {
		name    string
		method  string
		id      string
		mock    mockShortener
		want    want
		wantErr bool
	}{
		{
			name:   "success: 307 with Location header",
			method: http.MethodGet,
			id:     "abc",
			mock: mockShortener{
				resolveFn: func(id string) (string, error) {
					return "http://example.com/", nil
				},
			},
			want: want{
				code:        http.StatusTemporaryRedirect,
				contentType: "text/html; charset=utf-8",
				location:    "http://example.com/",
			},
		},
		{
			name:   "bad method -> 405, resolve not called",
			method: http.MethodPost,
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				body:        "Invalid request method\n",
			},
			wantErr: true,
		},
		{
			name:   "empty id -> 404, not found",
			method: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "URL not found\n",
			},
			wantErr: true,
		},
		{
			name:   "resolve error -> 404, not found",
			method: http.MethodGet,
			mock: mockShortener{
				resolveFn: func(id string) (string, error) { return "", errors.New("Resolve error") },
			},
			id: "abc",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "URL not found\n",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(baseURL, &tt.mock, mockAuditor{})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, path, nil)

			handler.Redirect(w, r, tt.id)
			response := w.Result()
			defer response.Body.Close()

			assert.Equal(t, tt.want.code, response.StatusCode)
			assert.True(t, strings.EqualFold(tt.want.contentType, response.Header.Get("Content-Type")))

			if tt.wantErr {
				body, err := io.ReadAll(response.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.body, string(body))
			}
			assert.Equal(t, tt.want.location, response.Header.Get("Location"))
		})
	}

}

func TestHandler_ShortifyJSON(t *testing.T) {
	const (
		baseURL = "http://localhost:8080"
		path    = "/api/shorten"
	)
	type want struct {
		code        int
		contentType string
		body        string
	}
	tests := []struct {
		name        string
		method      string
		contentType string
		body        io.Reader
		mock        mockShortener
		want        want
	}{
		{
			name:        "success: returns 201",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(marshalRequestBody("http://example.com")),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "abc", nil
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				body:        marshalResponseBody(baseURL + "/" + "abc"),
			},
		},
		{
			name:   "bad method -> 405, shorten not called",
			method: http.MethodGet,
			body:   strings.NewReader(marshalRequestBody("http://example.com")),
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				body:        "Invalid request method\n",
			},
		},
		{
			name:        "missing content-type -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "",
			body:        strings.NewReader(marshalRequestBody("http://example.com")),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Content-Type must be application/json\n",
			},
		},
		{
			name:        "wrong content-type -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader(marshalRequestBody("http://example.com")),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Content-Type must be application/json\n",
			},
		},
		{
			name:        "body read error -> 400, shorten not called",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        &BodyErr{msg: "read error"},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Cannot decode request JSON body\n",
			},
		},
		{
			name:        "body too large 400, shorten not called",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(marshalRequestBody(strings.Repeat("a", 8<<10+1))),
			want: want{
				code:        http.StatusRequestEntityTooLarge,
				contentType: "text/plain; charset=utf-8",
				body:        "Request body too large\n",
			},
		},
		{
			name:        "shortener error -> 400, error message returned",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(marshalRequestBody("http://example.com")),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "", errors.New("Shorten error")
				},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Shorten error\n",
			},
		},
		{
			name:        "duplicate URL -> 409 with existing short URL in JSON format",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        strings.NewReader(marshalRequestBody("http://example.com")),
			mock: mockShortener{
				shortenFn: func(raw string) (string, error) {
					return "", &domainerr.URLAlreadyExistsError{ShortCode: "abc"}
				},
			},
			want: want{
				code:        http.StatusConflict,
				contentType: "application/json",
				body:        marshalResponseBody(baseURL + "/" + "abc"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(baseURL, &tt.mock, mockAuditor{})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, path, tt.body)
			r.Header.Set("Content-Type", tt.contentType)
			r = withVisitorID(r, uuid.New())

			handler.ShortifyJSON(w, r)
			response := w.Result()

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.code, response.StatusCode)
			assert.True(t, strings.EqualFold(tt.want.contentType, response.Header.Get("Content-Type")))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}

func TestHandler_UserURLS(t *testing.T) {
	const (
		baseURL = "http://localhost:8080"
		path    = "/api/user/urls"
	)

	type want struct {
		code        int
		contentType string
		body        string
	}

	vid := uuid.New()

	tests := []struct {
		name    string
		method  string
		withVID bool
		mock    mockShortener
		want    want
	}{
		{
			name:    "success: returns 200 with user urls",
			method:  http.MethodGet,
			withVID: true,
			mock: mockShortener{
				userURLsFn: func(vid uuid.UUID) ([]model.URL, error) {
					return []model.URL{
						{OriginalURL: "http://example.com/a", ShortCode: "abc"},
						{OriginalURL: "http://example.com/b", ShortCode: "xyz"},
					}, nil
				},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body: marshalUserURLsBody([]model.UserURLsResponse{
					{ShortURL: baseURL + "/abc", OriginalURL: "http://example.com/a"},
					{ShortURL: baseURL + "/xyz", OriginalURL: "http://example.com/b"},
				}),
			},
		},
		{
			name:    "no urls: returns 204",
			method:  http.MethodGet,
			withVID: true,
			mock: mockShortener{
				userURLsFn: func(vid uuid.UUID) ([]model.URL, error) {
					return []model.URL{}, nil
				},
			},
			want: want{
				code:        http.StatusNoContent,
				contentType: "application/json",
				body:        "",
			},
		},
		{
			name:   "bad method -> 405",
			method: http.MethodPost,
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				body:        "Invalid request method\n",
			},
		},
		{
			name:   "missing visitor id -> 500",
			method: http.MethodGet,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "invalid visitor id\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(baseURL, &tt.mock, mockAuditor{})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, path, nil)
			if tt.withVID {
				r = withVisitorID(r, vid)
			}

			handler.UserURLS(w, r)
			response := w.Result()
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.code, response.StatusCode)
			assert.True(t, strings.EqualFold(tt.want.contentType, response.Header.Get("Content-Type")))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}
