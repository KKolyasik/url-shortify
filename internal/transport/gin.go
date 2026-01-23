package transport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Shortifyer interface {
	Shortify(http.ResponseWriter, *http.Request)
}

type Redirector interface {
	Redirect(http.ResponseWriter, *http.Request, string)
}

func GinShortify(s Shortifyer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.Shortify(ctx.Writer, ctx.Request)
	}
}

func GinRedirect(r Redirector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		r.Redirect(ctx.Writer, ctx.Request, id)
	}
}

func GinRequestLogger(log *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        log.Info("http request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Int("size", c.Writer.Size()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}