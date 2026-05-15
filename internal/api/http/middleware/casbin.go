package middleware

import (
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
)

func Casbin(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			response.Error(c, apperror.New(apperror.CodeForbidden, types.MessagePermissionNotReady))
			c.Abort()
			return
		}
		claims, ok := Claims(c)
		if !ok {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, types.MessageMissingAuthClaims))
			c.Abort()
			return
		}
		for _, role := range claims.Roles {
			allowed, err := enforcer.Enforce(role, c.FullPath(), c.Request.Method)
			if err != nil {
				response.Error(c, apperror.Wrap(apperror.CodeDependency, types.MessagePermissionCheckFail, err))
				c.Abort()
				return
			}
			if allowed {
				c.Next()
				return
			}
		}
		response.Error(c, apperror.New(apperror.CodeForbidden, types.MessagePermissionDenied))
		c.Abort()
	}
}
