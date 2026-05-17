package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
)

func TestErrorLocalizesMessageFromAcceptLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	Error(c, apperror.New(apperror.CodeInvalidArgument, types.MessageInvalidRequestBody))

	var body Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Msg != "请求体无效" {
		t.Fatalf("localized msg = %q", body.Msg)
	}
}

func TestOKKeepsEnglishByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/", nil)

	OK(c, gin.H{"status": "ok"})

	var body Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Msg != types.MessageOK {
		t.Fatalf("default msg = %q, want %q", body.Msg, types.MessageOK)
	}
}
