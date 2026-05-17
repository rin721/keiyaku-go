package i18n

import "golang.org/x/text/language"

// Message is the minimal message shape this repository needs from go-i18n.
type Message struct {
	ID    string
	Other string
}

// Catalog describes the supported languages and message sets for a translator.
type Catalog struct {
	Default   language.Tag
	Supported []language.Tag
	Messages  map[language.Tag][]Message
}
