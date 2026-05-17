package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/types"
	"golang.org/x/text/language"
)

const (
	HeaderAcceptLanguage  = "Accept-Language"
	HeaderContentLanguage = "Content-Language"
	contextLanguageKey    = "keiyaku.language"
)

var (
	LanguageENUS = language.MustParse("en-US")
	LanguageZHCN = language.MustParse("zh-CN")
)

func Resolve(acceptLanguage string) language.Tag {
	return translator.Match(acceptLanguage)
}

func Set(c *gin.Context, tag language.Tag) {
	if c == nil {
		return
	}
	tag = translator.MatchTags(tag)
	c.Set(contextLanguageKey, tag)
	c.Header(HeaderContentLanguage, tag.String())
}

func FromContext(c *gin.Context) language.Tag {
	if c == nil {
		return translator.Default()
	}
	if value, ok := c.Get(contextLanguageKey); ok {
		if tag, ok := value.(language.Tag); ok {
			return translator.MatchTags(tag)
		}
	}
	return Resolve(c.GetHeader(HeaderAcceptLanguage))
}

func Message(c *gin.Context, code types.Code, msg string) string {
	if msg == "" {
		msg = types.Message(code)
	}
	return Translate(msg, FromContext(c))
}

func Translate(msg string, tag language.Tag) string {
	return translator.Localize(tag, msg, msg)
}
