package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/i18n"
)

func I18N() gin.HandlerFunc {
	return func(c *gin.Context) {
		i18n.Set(c, i18n.Resolve(c.GetHeader(i18n.HeaderAcceptLanguage)))
		c.Next()
	}
}
