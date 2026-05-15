package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	"github.com/rin721/keiyaku-go/types"
)

const claimsKey = types.ContextAuthClaims

func Auth(tokens port.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, types.MessageMissingAuthHeader))
			c.Abort()
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, types.MessageInvalidAuthScheme))
			c.Abort()
			return
		}
		claims, err := tokens.ParseAccessToken(c.Request.Context(), strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, types.MessageInvalidAccessToken))
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func Claims(c *gin.Context) (port.TokenClaims, bool) {
	if c == nil {
		return port.TokenClaims{}, false
	}
	value, ok := c.Get(claimsKey)
	if !ok {
		return port.TokenClaims{}, false
	}
	claims, ok := value.(port.TokenClaims)
	return claims, ok
}
