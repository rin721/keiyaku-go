package platform

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
)

func GatewaySignature(pluginKey string, secret string, nonceStore pluginsdk.NonceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := pluginsdk.VerifySignedRequest(c.Request, pluginsdk.VerifyRequestOptions{
			Secret:            secret,
			MaxBodyBytes:      10 << 20,
			Now:               time.Now().UTC(),
			Skew:              pluginsdk.DefaultSignatureSkew,
			ExpectedPluginKey: pluginKey,
			NonceStore:        nonceStore,
		})
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
