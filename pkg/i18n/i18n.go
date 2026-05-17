package i18n

import (
	"fmt"
	"sort"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator wraps go-i18n with repository-local defaults for fallback and
// Accept-Language matching.
type Translator struct {
	bundle        *goi18n.Bundle
	matcher       language.Matcher
	defaultTag    language.Tag
	supportedTags []language.Tag
}

// NewTranslator builds a reusable translator from an in-memory catalog.
func NewTranslator(catalog Catalog) (*Translator, error) {
	defaultTag := catalog.Default
	if defaultTag == language.Und {
		defaultTag = language.English
	}

	bundle := goi18n.NewBundle(defaultTag)
	for tag, messages := range catalog.Messages {
		for _, message := range messages {
			if message.ID == "" {
				return nil, fmt.Errorf("i18n message id is required for %s", tag)
			}
			other := message.Other
			if other == "" {
				other = message.ID
			}
			if err := bundle.AddMessages(tag, &goi18n.Message{ID: message.ID, Other: other}); err != nil {
				return nil, fmt.Errorf("add i18n message %q for %s: %w", message.ID, tag, err)
			}
		}
	}

	supported := supportedTags(defaultTag, catalog)
	return &Translator{
		bundle:        bundle,
		matcher:       language.NewMatcher(supported),
		defaultTag:    defaultTag,
		supportedTags: supported,
	}, nil
}

// Default returns the translator fallback language.
func (t *Translator) Default() language.Tag {
	if t == nil {
		return language.Und
	}
	return t.defaultTag
}

// Match selects the best supported language for an Accept-Language header.
func (t *Translator) Match(acceptLanguage string) language.Tag {
	if t == nil {
		return language.Und
	}
	tags, _, err := language.ParseAcceptLanguage(strings.ReplaceAll(acceptLanguage, "_", "-"))
	if err != nil && len(tags) == 0 {
		return t.defaultTag
	}
	return t.MatchTags(tags...)
}

// MatchTags selects the best supported language from parsed language tags.
func (t *Translator) MatchTags(tags ...language.Tag) language.Tag {
	if t == nil {
		return language.Und
	}
	if len(tags) == 0 {
		return t.defaultTag
	}
	_, index, _ := t.matcher.Match(tags...)
	if index < 0 || index >= len(t.supportedTags) {
		return t.defaultTag
	}
	return t.supportedTags[index]
}

// Localize returns a localized message or the supplied fallback.
func (t *Translator) Localize(tag language.Tag, messageID string, fallback string) string {
	if t == nil || messageID == "" {
		return fallback
	}
	if fallback == "" {
		fallback = messageID
	}
	tag = t.MatchTags(tag)
	localizer := goi18n.NewLocalizer(t.bundle, tag.String(), t.defaultTag.String())
	message, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: &goi18n.Message{ID: messageID, Other: fallback},
	})
	if err != nil {
		return fallback
	}
	return message
}

func supportedTags(defaultTag language.Tag, catalog Catalog) []language.Tag {
	tags := []language.Tag{defaultTag}
	for _, tag := range catalog.Supported {
		tags = appendTag(tags, tag)
	}

	messageTags := make([]language.Tag, 0, len(catalog.Messages))
	for tag := range catalog.Messages {
		messageTags = append(messageTags, tag)
	}
	sort.Slice(messageTags, func(i, j int) bool {
		return messageTags[i].String() < messageTags[j].String()
	})
	for _, tag := range messageTags {
		tags = appendTag(tags, tag)
	}

	return tags
}

func appendTag(tags []language.Tag, tag language.Tag) []language.Tag {
	if tag == language.Und {
		return tags
	}
	for _, existing := range tags {
		if existing == tag {
			return tags
		}
	}
	return append(tags, tag)
}
