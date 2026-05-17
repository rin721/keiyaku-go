# pkg/i18n

`pkg/i18n` wraps `github.com/nicksnyder/go-i18n/v2/i18n` behind the small API this repository needs.

It solves language matching and message lookup. It does not own application message catalogs, HTTP response shapes, or Gin middleware. Keep those adapters in `internal/`.

## Usage

```go
translator, err := i18n.NewTranslator(i18n.Catalog{
    Default: language.MustParse("en-US"),
    Supported: []language.Tag{
        language.MustParse("en-US"),
        language.MustParse("zh-CN"),
    },
    Messages: map[language.Tag][]i18n.Message{
        language.MustParse("en-US"): {{ID: "ok", Other: "ok"}},
        language.MustParse("zh-CN"): {{ID: "ok", Other: "success"}},
    },
})
```

Use `Match` for `Accept-Language` headers, `MatchTags` for already parsed tags, and `Localize` for fallback-safe message lookup.
