package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/observability/trace"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered",
					zap.String("panic", fmt.Sprint(recovered)),
					zap.String("trace_id", trace.IDFromContext(c.Request.Context())),
					zap.ByteString("stack", debug.Stack()),
				)
				response.Error(c, apperror.New(apperror.CodeInternal, "internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
