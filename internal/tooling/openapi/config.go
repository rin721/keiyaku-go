package openapi

import (
	"path/filepath"
	"strings"
)

func NormalizeConfig(config Config) Config {
	if strings.TrimSpace(config.HandlerDir) == "" {
		config.HandlerDir = DefaultHandlerDir
	}
	if strings.TrimSpace(config.DTODir) == "" {
		config.DTODir = DefaultDTODir
	}
	if strings.TrimSpace(config.Output) == "" {
		config.Output = DefaultOutputPath
	}
	if strings.TrimSpace(config.BasePath) == "" {
		config.BasePath = DefaultBasePath
	}
	if !strings.HasPrefix(config.BasePath, "/") {
		config.BasePath = "/" + config.BasePath
	}
	config.BasePath = strings.TrimRight(filepath.ToSlash(config.BasePath), "/")
	if config.BasePath == "" {
		config.BasePath = "/"
	}
	if strings.TrimSpace(config.Title) == "" {
		config.Title = DefaultTitle
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = DefaultVersion
	}
	return config
}
