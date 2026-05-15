package middleware

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/sony/gobreaker"
)

func CircuitBreaker(name string, threshold uint32, timeout time.Duration) gin.HandlerFunc {
	if threshold == 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool { return counts.ConsecutiveFailures >= threshold },
	})
	return func(c *gin.Context) {
		_, err := breaker.Execute(func() (interface{}, error) {
			c.Next()
			if c.Writer.Status() >= 500 {
				return nil, errors.New("request failed")
			}
			return nil, nil
		})
		if err != nil && !c.Writer.Written() {
			response.Error(c, apperror.New(apperror.CodeDependency, "service temporarily unavailable"))
			c.Abort()
		}
	}
}
