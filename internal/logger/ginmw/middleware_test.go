package ginmw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	msg    string
	fields []interface{}
	calls  int
}

func (l *mockLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.msg = msg
	l.fields = keysAndValues
	l.calls++
}

func TestRequestLogger(t *testing.T) {
	type want struct {
		code   int
		calls  int
		status int
		msg    string
		method string
		path   string
		size   int
	}
	tests := []struct {
		name       string
		method     string
		sendMetgod string
		path       string
		sendPath   string
		router     func(ctx *gin.Context)
		want       want
	}{
		{
			name:       "success log",
			router:     func(ctx *gin.Context) { ctx.String(http.StatusOK, "pong") },
			method:     http.MethodGet,
			sendMetgod: http.MethodGet,
			path:       "/ping",
			sendPath:   "/ping",
			want: want{
				code:   http.StatusOK,
				calls:  1,
				status: http.StatusOK,
				msg:    "http request",
				method: http.MethodGet,
				path:   "/ping",
				size:   len("pong"),
			},
		},
		{
			name:       "invalid method",
			router:     func(ctx *gin.Context) { ctx.String(http.StatusOK, "pong") },
			method:     http.MethodGet,
			sendMetgod: http.MethodPost,
			path:       "/ping",
			sendPath:   "/ping",
			want: want{
				code:   http.StatusNotFound,
				calls:  1,
				status: http.StatusNotFound,
				msg:    "http request",
				method: http.MethodPost,
				path:   "/ping",
				size:   -1,
			},
		},
		{
			name:       "invalid path",
			router:     func(ctx *gin.Context) { ctx.String(http.StatusOK, "pong") },
			method:     http.MethodGet,
			sendMetgod: http.MethodGet,
			path:       "/ping",
			sendPath:   "/pong",
			want: want{
				code:   http.StatusNotFound,
				calls:  1,
				status: http.StatusNotFound,
				msg:    "http request",
				method: http.MethodGet,
				path:   "/pong",
				size:   -1,
			},
		},
	}

	for _, tt := range tests {

		l := &mockLogger{}

		g := gin.New()
		g.Use(RequestLogger(l))
		g.Handle(tt.method, tt.path, tt.router)
		r := httptest.NewRequest(tt.sendMetgod, tt.sendPath, nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, r)

		response := w.Result()

		assert.Equal(t, tt.want.code, response.StatusCode)
		assert.Equal(t, tt.want.msg, l.msg)
		assert.Equal(t, tt.want.calls, l.calls)

		method, ok := kvGet(l.fields, "method")
		require.True(t, ok)
		assert.Equal(t, tt.want.method, method)

		path, ok := kvGet(l.fields, "path")
		require.True(t, ok)
		assert.Equal(t, tt.want.path, path)

		status, ok := kvGet(l.fields, "status")
		require.True(t, ok)
		assert.Equal(t, tt.want.status, status)

		size, ok := kvGet(l.fields, "size")
		require.True(t, ok)
		assert.Equal(t, tt.want.size, size)

	}
}

func kvGet(kv []interface{}, key string) (interface{}, bool) {
	for i := 0; i+1 < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			continue
		}
		if k == key {
			return kv[i+1], true
		}
	}
	return nil, false
}
