package platform

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
)

func GatewaySignature(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := pluginsdk.VerifySignedRequest(c.Request, secret, 10<<20, time.Now().UTC(), pluginsdk.DefaultSignatureSkew)
		if err != nil {
			status := http.StatusUnauthorized
			msg := "invalid gateway signature"
			if errors.Is(err, pluginsdk.ErrBodyTooLarge) {
				status = http.StatusRequestEntityTooLarge
				msg = "request body too large"
			}
			c.AbortWithStatusJSON(status, gin.H{"code": status, "msg": msg})
			return
		}
		c.Next()
	}
}
