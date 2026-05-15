package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/observability/trace"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(trace.HeaderName)
		if traceID == "" {
			traceID = trace.NewID()
		}
		ctx := trace.WithID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(trace.HeaderName, traceID)
		c.Next()
	}
}
