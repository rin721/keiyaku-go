package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"
)

func TestTranslatorMatchesAcceptLanguageAndLocalizes(t *testing.T) {
	enUS := language.MustParse("en-US")
	zhCN := language.MustParse("zh-CN")
	translator, err := NewTranslator(Catalog{
		Default:   enUS,
		Supported: []language.Tag{enUS, zhCN},
		Messages: map[language.Tag][]Message{
			enUS: {{ID: "invalid request body", Other: "invalid request body"}},
			zhCN: {{ID: "invalid request body", Other: "请求体无效"}},
		},
	})
	if err != nil {
		t.Fatalf("NewTranslator() error = %v", err)
	}

	tag := translator.Match("fr-FR,zh;q=0.8,en;q=0.7")
	if tag != zhCN {
		t.Fatalf("Match() = %s, want %s", tag, zhCN)
	}
	if got := translator.Localize(tag, "invalid request body", "invalid request body"); got != "请求体无效" {
		t.Fatalf("Localize() = %q", got)
	}
	if got := translator.Localize(tag, "custom message", "custom message"); got != "custom message" {
		t.Fatalf("Localize() fallback = %q", got)
	}
}

func TestNewTranslatorFromFilesLoadsYAML(t *testing.T) {
	enUS := language.MustParse("en-US")
	zhCN := language.MustParse("zh-CN")
	dir := t.TempDir()
	enPath := filepath.Join(dir, "en-US.yaml")
	zhPath := filepath.Join(dir, "zh-CN.yaml")
	mustWriteI18NFile(t, enPath, `"invalid request body": "invalid request body"`)
	mustWriteI18NFile(t, zhPath, `"invalid request body": "请求体无效"`)

	translator, err := NewTranslatorFromFiles(FileCatalog{
		Default:   enUS,
		Supported: []language.Tag{enUS, zhCN},
		Files: map[language.Tag]string{
			enUS: enPath,
			zhCN: zhPath,
		},
	})
	if err != nil {
		t.Fatalf("NewTranslatorFromFiles() error = %v", err)
	}

	tag := translator.Match("fr-FR,zh;q=0.8,en;q=0.7")
	if tag != zhCN {
		t.Fatalf("Match() = %s, want %s", tag, zhCN)
	}
	if got := translator.Localize(tag, "invalid request body", "invalid request body"); got != "请求体无效" {
		t.Fatalf("Localize() = %q", got)
	}
	if got := translator.Localize(tag, "custom message", "custom message"); got != "custom message" {
		t.Fatalf("Localize() fallback = %q", got)
	}
}

func mustWriteI18NFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func TestNewTranslatorRejectsEmptyMessageID(t *testing.T) {
	_, err := NewTranslator(Catalog{
		Default: language.English,
		Messages: map[language.Tag][]Message{
			language.English: {{Other: "missing id"}},
		},
	})
	if err == nil {
		t.Fatal("NewTranslator() error is nil")
	}
}
