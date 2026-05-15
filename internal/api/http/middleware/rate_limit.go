package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
	"golang.org/x/time/rate"
)

func RateLimit(requestsPerSecond float64, burst int) gin.HandlerFunc {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 100
	}
	if burst <= 0 {
		burst = int(requestsPerSecond)
	}
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Error(c, apperror.New(apperror.CodeTooManyRequests, types.MessageTooManyRequests))
			c.Abort()
			return
		}
		c.Next()
	}
}
