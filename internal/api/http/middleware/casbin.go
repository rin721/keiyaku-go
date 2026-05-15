package middleware

import (
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
)

func Casbin(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			response.Error(c, apperror.New(apperror.CodeForbidden, "permission service is not ready"))
			c.Abort()
			return
		}
		claims, ok := Claims(c)
		if !ok {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "missing auth claims"))
			c.Abort()
			return
		}
		for _, role := range claims.Roles {
			allowed, err := enforcer.Enforce(role, c.FullPath(), c.Request.Method)
			if err != nil {
				response.Error(c, apperror.Wrap(apperror.CodeDependency, "permission check failed", err))
				c.Abort()
				return
			}
			if allowed {
				c.Next()
				return
			}
		}
		response.Error(c, apperror.New(apperror.CodeForbidden, "permission denied"))
		c.Abort()
	}
}
