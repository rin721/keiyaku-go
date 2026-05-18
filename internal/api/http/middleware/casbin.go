package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
)

func Casbin(authorizer port.Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authorizer == nil {
			response.Error(c, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionNotReady))
			c.Abort()
			return
		}
		claims, ok := Claims(c)
		if !ok {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageMissingAuthClaims))
			c.Abort()
			return
		}
		for _, role := range claims.Roles {
			allowed, err := authorizer.Allow(role, c.FullPath(), c.Request.Method)
			if err != nil {
				response.Error(c, apperror.Wrap(apperror.CodeDependency, apperror.MessagePermissionCheckFail, err))
				c.Abort()
				return
			}
			if allowed {
				c.Next()
				return
			}
		}
		response.Error(c, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionDenied))
		c.Abort()
	}
}
