package apidocs

import (
	"path"
	"strings"
)

func normalizeOptions(options Options) Options {
	options.Path = normalizeRoutePath(options.Path, DefaultPath)
	options.SpecPath = normalizeRoutePath(options.SpecPath, DefaultSpecPath)
	options.Title = strings.TrimSpace(options.Title)
	if options.Title == "" {
		options.Title = DefaultTitle
	}
	return options
}

func normalizeRoutePath(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == "/" {
		return fallback
	}
	return cleaned
}
