package platform

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
)

func TestGatewaySignatureAllowsSignedRequestAndRestoresBody(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	secret := "abcdefghijklmnopqrstuvwxyz123456"
	body := []byte(`{"title":"ok"}`)
	engine := gin.New()
	engine.POST("/articles", GatewaySignature(secret), func(c *gin.Context) {
		content, err := c.GetRawData()
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(content) != string(body) {
			t.Fatalf("body = %q, want %q", content, body)
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(string(body)))
	if err := pluginsdk.SignRequest(req, "demo-plugin", secret, body, time.Now().UTC(), "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestGatewaySignatureRejectsBodyMismatch(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	secret := "abcdefghijklmnopqrstuvwxyz123456"
	engine := gin.New()
	engine.POST("/articles", GatewaySignature(secret), func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})
	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"title":"right"}`))
	if err := pluginsdk.SignRequest(req, "demo-plugin", secret, []byte(`{"title":"left"}`), time.Now().UTC(), "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
