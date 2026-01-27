package ginmw

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gZIPCompress struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *gZIPCompress) Write(p []byte) (int, error) {
	return c.Writer.Write(p)
}

func GinCompress() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
			
		}
	}
}