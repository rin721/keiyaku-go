package i18n

import (
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
