package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type URL struct {
	url string
	id string
}

func (u *URL) Shorten(raw string) (string, error) {
	return u.id, nil
}

func (u *URL) Resolve(id string) (string, error) {
	return u.url, nil
}

func TestHandler_Shortify(t *testing.T) {
	var path string = "/"
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		url URL
		want want
	}{
		{
			name: "success",
			url: URL{
				url: "http://example.com/",
				id: "example",
			},
			want: want{
				code: http.StatusCreated,
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hander := New("test", &tt.url)
			body := bytes.NewReader([]byte(tt.url.url))
			r := httptest.NewRequest(http.MethodPost, path, body)
			r.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			hander.Shortify(w, r)
			response := w.Result()
			assert.Equal(t, tt.want.code, response.StatusCode)
		})
	}
}
