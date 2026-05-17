// Package i18n provides a small repository-facing wrapper around go-i18n.
//
// It owns language negotiation and message lookup, while callers keep their
// domain-specific catalogs and transport adapters outside this package.
package i18n
