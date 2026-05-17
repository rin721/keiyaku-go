package i18n

import (
	"testing"

	"github.com/rin721/keiyaku-go/types"
	"golang.org/x/text/language"
)

func TestResolve(t *testing.T) {
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
