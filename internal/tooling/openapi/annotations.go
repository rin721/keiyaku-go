package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var supportedMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodHead:    {},
	http.MethodOptions: {},
}

func ParseHandlers(dir string, registry *SchemaRegistry) ([]Operation, error) {
	files, err := goFiles(dir)
	if err != nil {
		return nil, err
	}
	fileSet := token.NewFileSet()
	var operations []Operation
	seen := make(map[string]token.Position)
	for _, file := range files {
		parsed, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse handler %s: %w", file, err)
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			operation, found, err := parseOperation(fn.Doc, fileSet.Position(fn.Pos()), registry)
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			key := operation.Method + " " + operation.Path
			if previous, ok := seen[key]; ok {
				return nil, fmt.Errorf("%s: duplicate operation %s already defined at %s", fileSet.Position(fn.Pos()), key, previous)
			}
			seen[key] = fileSet.Position(fn.Pos())
			operations = append(operations, operation)
		}
	}
	sort.Slice(operations, func(i, j int) bool {
		left := operations[i].Path + " " + operations[i].Method
		right := operations[j].Path + " " + operations[j].Method
		return left < right
	})
	return operations, nil
}

func parseOperation(doc *ast.CommentGroup, position token.Position, registry *SchemaRegistry) (Operation, bool, error) {
	if doc == nil {
		return Operation{}, false, nil
	}
	var operation Operation
	found := false
	for _, comment := range doc.List {
		line := normalizeComment(comment.Text)
		if !strings.HasPrefix(line, "@") {
			continue
		}
		name, value := splitAnnotation(line)
		if name == "" {
			continue
		}
		switch name {
		case "Summary":
			found = true
			operation.Summary = value
		case "Description":
			found = true
			if operation.Description == "" {
				operation.Description = value
			} else {
				operation.Description += "\n" + value
			}
		case "Tags":
			found = true
			operation.Tags = splitCommaValues(value)
		case "Accept":
			found = true
			operation.Accepts = normalizeMIMEs(value)
		case "Produce":
			found = true
			operation.Produces = normalizeMIMEs(value)
		case "Param":
			found = true
			parameter, err := parseParam(value, registry)
			if err != nil {
				return Operation{}, true, fmt.Errorf("%s: %w", position, err)
			}
			if parameter.In == "body" {
				operation.RequestBody = &RequestBody{
					Description: parameter.Description,
					Required:    parameter.Required,
					Schema:      parameter.Schema,
				}
				continue
			}
			operation.Parameters = append(operation.Parameters, parameter)
		case "Success", "Failure", "Response":
			found = true
			response, err := parseResponse(name, value, registry)
			if err != nil {
				return Operation{}, true, fmt.Errorf("%s: %w", position, err)
			}
			operation.Responses = append(operation.Responses, response)
		case "Security":
			found = true
			if strings.TrimSpace(value) != "" {
				operation.Security = append(operation.Security, strings.TrimSpace(value))
			}
		case "Router":
			found = true
			path, method, err := parseRouter(value)
			if err != nil {
				return Operation{}, true, fmt.Errorf("%s: %w", position, err)
			}
			operation.Path = path
			operation.Method = method
		}
	}
	if !found {
		return Operation{}, false, nil
	}
	if operation.Path == "" || operation.Method == "" {
		return Operation{}, true, fmt.Errorf("%s: openapi annotation is missing @Router", position)
	}
	if len(operation.Responses) == 0 {
		return Operation{}, true, fmt.Errorf("%s: openapi annotation is missing @Success/@Failure response", position)
	}
	if len(operation.Produces) == 0 {
		operation.Produces = []string{"application/json"}
	}
	sort.Slice(operation.Responses, func(i, j int) bool {
		return operation.Responses[i].Code < operation.Responses[j].Code
	})
	return operation, true, nil
}

func parseParam(value string, registry *SchemaRegistry) (Parameter, error) {
	fields := splitFields(value)
	if len(fields) < 5 {
		return Parameter{}, fmt.Errorf("@Param requires name, in, type, required and description")
	}
	required, err := strconv.ParseBool(fields[3])
	if err != nil {
		return Parameter{}, fmt.Errorf("@Param %s required value must be true or false", fields[0])
	}
	parameter := Parameter{
		Name:        fields[0],
		In:          fields[1],
		TypeName:    fields[2],
		Required:    required,
		Description: strings.Join(fields[4:], " "),
		Schema:      registry.SchemaForAnnotationType(fields[2]),
	}
	if parameter.In == "path" {
		parameter.Required = true
	}
	switch parameter.In {
	case "query", "path", "header", "body":
		return parameter, nil
	default:
		return Parameter{}, fmt.Errorf("@Param %s has unsupported location %q", parameter.Name, parameter.In)
	}
}

func parseResponse(name string, value string, registry *SchemaRegistry) (Response, error) {
	fields := splitFields(value)
	if len(fields) < 4 {
		return Response{}, fmt.Errorf("@%s requires code, kind, type and description", name)
	}
	code := fields[0]
	description := strings.Join(fields[3:], " ")
	response := Response{Code: code, Description: description}
	if strings.HasPrefix(code, "2") {
		response.Schema = registry.SchemaForAnnotationType(fields[2])
		if strings.EqualFold(strings.Trim(fields[1], "{}"), "array") {
			response.Schema = &Schema{Type: "array", Items: response.Schema}
		}
	}
	return response, nil
}

func parseRouter(value string) (string, string, error) {
	fields := splitFields(value)
	if len(fields) != 2 {
		return "", "", fmt.Errorf("@Router requires path and method")
	}
	path := strings.TrimSpace(fields[0])
	if !strings.HasPrefix(path, "/") {
		return "", "", fmt.Errorf("@Router path must start with /")
	}
	method := strings.ToUpper(strings.Trim(fields[1], "[]"))
	if _, ok := supportedMethods[method]; !ok {
		return "", "", fmt.Errorf("@Router has unsupported method %q", method)
	}
	return path, strings.ToLower(method), nil
}

func normalizeComment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "//")
	value = strings.TrimPrefix(value, "/*")
	value = strings.TrimSuffix(value, "*/")
	return strings.TrimSpace(value)
}

func splitAnnotation(line string) (string, string) {
	line = strings.TrimPrefix(strings.TrimSpace(line), "@")
	if line == "" {
		return "", ""
	}
	name, value, ok := strings.Cut(line, " ")
	if !ok {
		return line, ""
	}
	return strings.TrimSpace(name), strings.TrimSpace(value)
}

func splitCommaValues(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func normalizeMIMEs(value string) []string {
	values := splitCommaValues(value)
	if len(values) == 0 && strings.TrimSpace(value) != "" {
		values = strings.Fields(value)
	}
	result := make([]string, 0, len(values))
	for _, item := range values {
		switch strings.ToLower(item) {
		case "json":
			result = append(result, "application/json")
		case "xml":
			result = append(result, "application/xml")
		case "plain":
			result = append(result, "text/plain")
		default:
			result = append(result, item)
		}
	}
	return result
}

func splitFields(value string) []string {
	var fields []string
	var current strings.Builder
	inQuote := false
	escaped := false
	for _, r := range value {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case (r == ' ' || r == '\t') && !inQuote:
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields
}

func goFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read go dir %s: %w", dir, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}
