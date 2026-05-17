package i18n

import (
	"fmt"

	tooli18n "github.com/rin721/keiyaku-go/pkg/i18n"
	"golang.org/x/text/language"
)

var translator = mustEmptyTranslator()

func Init(defaultLanguage string, supportedLanguages []string, files map[string]string) error {
	defaultTag, err := language.Parse(defaultLanguage)
	if err != nil {
		return fmt.Errorf("parse i18n default language %q: %w", defaultLanguage, err)
	}

	supported := make([]language.Tag, 0, len(supportedLanguages))
	for _, raw := range supportedLanguages {
		tag, err := language.Parse(raw)
		if err != nil {
			return fmt.Errorf("parse i18n supported language %q: %w", raw, err)
		}
		supported = append(supported, tag)
	}

	messageFiles := make(map[language.Tag]string, len(files))
	for raw, path := range files {
		tag, err := language.Parse(raw)
		if err != nil {
			return fmt.Errorf("parse i18n file language %q: %w", raw, err)
		}
		messageFiles[tag] = path
	}

	loaded, err := tooli18n.NewTranslatorFromFiles(tooli18n.FileCatalog{
		Default:   defaultTag,
		Supported: supported,
		Files:     messageFiles,
	})
	if err != nil {
		return err
	}
	translator = loaded
	return nil
}

func mustEmptyTranslator() *tooli18n.Translator {
	translator, err := tooli18n.NewTranslator(tooli18n.Catalog{
		Default:   LanguageENUS,
		Supported: []language.Tag{LanguageENUS, LanguageZHCN},
	})
	if err != nil {
		panic(fmt.Sprintf("build empty http i18n translator: %v", err))
	}
	return translator
}
