package openapi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func Generate(config Config) ([]byte, error) {
	config = NormalizeConfig(config)

	registry, err := LoadSchemas(config.DTODir)
	if err != nil {
		return nil, err
	}
	operations, err := ParseHandlers(config.HandlerDir, registry)
	if err != nil {
		return nil, err
	}
	doc := Document{
		Title:    config.Title,
		Version:  config.Version,
		BasePath: config.BasePath,
		Paths:    groupOperations(operations),
		Schemas:  registry.Schemas(),
	}
	return MarshalYAML(doc), nil
}

func GenerateFile(config Config) error {
	config = NormalizeConfig(config)
	content, err := Generate(config)
	if err != nil {
		return err
	}
	if config.Check {
		existing, err := os.ReadFile(config.Output)
		if err != nil {
			return fmt.Errorf("read openapi output: %w", err)
		}
		if !bytes.Equal(existing, content) {
			return fmt.Errorf("openapi spec is out of date: run go run ./cmd/openapi generate")
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(config.Output), 0o755); err != nil {
		return fmt.Errorf("create openapi output dir: %w", err)
	}
	if err := os.WriteFile(config.Output, content, 0o644); err != nil {
		return fmt.Errorf("write openapi output: %w", err)
	}
	return nil
}

func groupOperations(operations []Operation) map[string]map[string]Operation {
	paths := make(map[string]map[string]Operation)
	for _, operation := range operations {
		if _, ok := paths[operation.Path]; !ok {
			paths[operation.Path] = make(map[string]Operation)
		}
		paths[operation.Path][operation.Method] = operation
	}
	return paths
}
