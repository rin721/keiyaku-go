package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/observability/trace"
	"go.uber.org/zap"
)

func Logging(logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request",
			zap.String("trace_id", trace.IDFromContext(c.Request.Context())),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
