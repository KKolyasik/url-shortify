package ginmw

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger captures the minimal logging contract for this middleware.
// It aligns with zap.SugaredLogger and similar "structured message + key/values" loggers.
type Logger interface {
	Infow(msg string, keysAndValues ...interface{})
}

func RequestLogger(log Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			log.Infow("http request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"size", c.Writer.Size(),
				"duration", time.Since(start),
			)
		}()

		c.Next()
	}
}
