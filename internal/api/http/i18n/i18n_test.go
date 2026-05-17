package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rin721/keiyaku-go/types"
	"golang.org/x/text/language"
)

func TestResolve(t *testing.T) {
	initTestTranslator(t)

	tests := map[string]language.Tag{
		"":                        LanguageENUS,
		"en-US,en;q=0.9":          LanguageENUS,
		"zh-CN,zh;q=0.9,en;q=0.8": LanguageZHCN,
		"fr-FR,zh;q=0.8,en;q=0.7": LanguageZHCN,
		"en;q=0.2, zh-CN;q=0.9":   LanguageZHCN,
		"fr-FR, de-DE;q=0.8":      LanguageENUS,
		"zh_TW;q=0":               LanguageENUS,
	}
	for header, want := range tests {
		if got := Resolve(header); got != want {
			t.Fatalf("Resolve(%q) = %s, want %s", header, got, want)
		}
	}
}

func TestTranslate(t *testing.T) {
	initTestTranslator(t)

	if got := Translate(types.MessageInvalidRequestBody, LanguageZHCN); got != "请求体无效" {
		t.Fatalf("Translate zh-CN = %q", got)
	}
	if got := Translate(types.MessageInvalidRequestBody, LanguageENUS); got != types.MessageInvalidRequestBody {
		t.Fatalf("Translate en-US = %q", got)
	}
	if got := Translate("custom message", LanguageZHCN); got != "custom message" {
		t.Fatalf("Translate fallback = %q", got)
	}
}

func TestInitLoadsRepositoryConfigFiles(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	err := Init("en-US", []string{"en-US", "zh-CN"}, map[string]string{
		"en-US": filepath.Join(root, "configs", "i18n", "en-US.yaml"),
		"zh-CN": filepath.Join(root, "configs", "i18n", "zh-CN.yaml"),
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if got := Translate(types.MessageInvalidRequestBody, LanguageZHCN); got != "请求体无效" {
		t.Fatalf("Translate zh-CN = %q", got)
	}
}

func initTestTranslator(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	enPath := filepath.Join(dir, "en-US.yaml")
	zhPath := filepath.Join(dir, "zh-CN.yaml")
	mustWriteFile(t, enPath, `"invalid request body": "invalid request body"`)
	mustWriteFile(t, zhPath, `"invalid request body": "请求体无效"`)
	err := Init("en-US", []string{"en-US", "zh-CN"}, map[string]string{
		"en-US": enPath,
		"zh-CN": zhPath,
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
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
