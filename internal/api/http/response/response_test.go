package response

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	httpi18n "github.com/rin721/keiyaku-go/internal/api/http/i18n"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
)

func TestErrorLocalizesMessageFromAcceptLanguage(t *testing.T) {
	initTestTranslator(t)

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
	initTestTranslator(t)

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

func initTestTranslator(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	enPath := filepath.Join(dir, "en-US.yaml")
	zhPath := filepath.Join(dir, "zh-CN.yaml")
	mustWriteFile(t, enPath, `"ok": "ok"`+"\n"+`"invalid request body": "invalid request body"`)
	mustWriteFile(t, zhPath, `"ok": "成功"`+"\n"+`"invalid request body": "请求体无效"`)
	err := httpi18n.Init("en-US", []string{"en-US", "zh-CN"}, map[string]string{
		"en-US": enPath,
		"zh-CN": zhPath,
	})
	if err != nil {
		t.Fatalf("i18n Init() error = %v", err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
