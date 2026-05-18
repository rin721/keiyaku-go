package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type SchemaRegistry struct {
	structs map[string]*ast.StructType
	schemas map[string]*Schema
}

func LoadSchemas(dir string) (*SchemaRegistry, error) {
	files, err := goFiles(dir)
	if err != nil {
		return nil, err
	}
	registry := &SchemaRegistry{
		structs: make(map[string]*ast.StructType),
		schemas: make(map[string]*Schema),
	}
	fileSet := token.NewFileSet()
	for _, file := range files {
		parsed, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse dto %s: %w", file, err)
		}
		for _, decl := range parsed.Decls {
			general, ok := decl.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				registry.structs[typeSpec.Name.Name] = structType
			}
		}
	}
	names := make([]string, 0, len(registry.structs))
	for name := range registry.structs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := registry.buildStruct(name); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *SchemaRegistry) Schemas() map[string]*Schema {
	result := make(map[string]*Schema, len(r.schemas))
	for name, schema := range r.schemas {
		result[name] = schema
	}
	return result
}

func (r *SchemaRegistry) SchemaForAnnotationType(typeName string) *Schema {
	typeName = cleanTypeName(typeName)
	if typeName == "" || typeName == "nil" {
		return nil
	}
	if schema := primitiveSchema(typeName); schema != nil {
		return schema
	}
	if _, ok := r.structs[typeName]; ok {
		return refSchema(typeName)
	}
	return &Schema{Type: "object"}
}

func (r *SchemaRegistry) buildStruct(name string) (*Schema, error) {
	if schema, ok := r.schemas[name]; ok {
		return schema, nil
	}
	structType, ok := r.structs[name]
	if !ok {
		return nil, fmt.Errorf("unknown dto struct %s", name)
	}
	schema := &Schema{Type: "object"}
	r.schemas[name] = schema
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		for _, fieldName := range field.Names {
			if !fieldName.IsExported() {
				continue
			}
			jsonName, omitempty, skip := jsonFieldName(fieldName.Name, field.Tag)
			if skip {
				continue
			}
			fieldSchema := r.schemaForExpr(field.Type)
			applyBinding(fieldSchema, field.Tag)
			schema.Properties = append(schema.Properties, Property{Name: jsonName, Schema: fieldSchema})
			if bindingRequired(field.Tag) && !omitempty {
				schema.Required = append(schema.Required, jsonName)
			}
		}
	}
	sort.Slice(schema.Properties, func(i, j int) bool {
		return schema.Properties[i].Name < schema.Properties[j].Name
	})
	sort.Strings(schema.Required)
	return schema, nil
}

func (r *SchemaRegistry) schemaForExpr(expr ast.Expr) *Schema {
	switch typed := expr.(type) {
	case *ast.Ident:
		if schema := primitiveSchema(typed.Name); schema != nil {
			return schema
		}
		if _, ok := r.structs[typed.Name]; ok {
			return refSchema(typed.Name)
		}
		return &Schema{Type: "object"}
	case *ast.SelectorExpr:
		if ident, ok := typed.X.(*ast.Ident); ok && ident.Name == "time" && typed.Sel.Name == "Time" {
			return &Schema{Type: "string", Format: "date-time"}
		}
		if schema := primitiveSchema(typed.Sel.Name); schema != nil {
			return schema
		}
		return &Schema{Type: "object"}
	case *ast.StarExpr:
		schema := cloneSchema(r.schemaForExpr(typed.X))
		if schema.Ref == "" {
			schema.Nullable = true
		}
		return schema
	case *ast.ArrayType:
		return &Schema{Type: "array", Items: r.schemaForExpr(typed.Elt)}
	case *ast.MapType:
		return &Schema{Type: "object"}
	default:
		return &Schema{Type: "object"}
	}
}

func primitiveSchema(typeName string) *Schema {
	switch cleanTypeName(typeName) {
	case "string":
		return &Schema{Type: "string"}
	case "bool", "boolean":
		return &Schema{Type: "boolean"}
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "integer":
		return &Schema{Type: "integer", Format: "int32"}
	case "int64", "uint64":
		return &Schema{Type: "integer", Format: "int64"}
	case "float32":
		return &Schema{Type: "number", Format: "float"}
	case "float64", "number":
		return &Schema{Type: "number", Format: "double"}
	default:
		return nil
	}
}

func refSchema(name string) *Schema {
	return &Schema{Ref: "#/components/schemas/" + name}
}

func cleanTypeName(typeName string) string {
	typeName = strings.TrimSpace(typeName)
	typeName = strings.TrimPrefix(typeName, "[]")
	typeName = strings.TrimPrefix(typeName, "*")
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
	}
	return strings.TrimSpace(typeName)
}

func cloneSchema(schema *Schema) *Schema {
	if schema == nil {
		return nil
	}
	cloned := *schema
	return &cloned
}

func jsonFieldName(fallback string, tag *ast.BasicLit) (string, bool, bool) {
	if tag == nil {
		return lowerFirst(fallback), false, false
	}
	raw, err := strconv.Unquote(tag.Value)
	if err != nil {
		return lowerFirst(fallback), false, false
	}
	value := reflect.StructTag(raw).Get("json")
	if value == "-" {
		return "", false, true
	}
	parts := strings.Split(value, ",")
	name := strings.TrimSpace(parts[0])
	if name == "" {
		name = lowerFirst(fallback)
	}
	omitempty := false
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty, false
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func bindingRequired(tag *ast.BasicLit) bool {
	return bindingHas(tag, "required")
}

func applyBinding(schema *Schema, tag *ast.BasicLit) {
	if schema == nil || tag == nil {
		return
	}
	raw, err := strconv.Unquote(tag.Value)
	if err != nil {
		return
	}
	for _, part := range strings.Split(reflect.StructTag(raw).Get("binding"), ",") {
		part = strings.TrimSpace(part)
		if part == "email" && schema.Type == "string" {
			schema.Format = "email"
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		number, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		switch key {
		case "min":
			if schema.Type == "string" {
				schema.MinLength = &number
			} else {
				schema.Minimum = &number
			}
		case "max":
			if schema.Type == "string" {
				schema.MaxLength = &number
			} else {
				schema.Maximum = &number
			}
		}
	}
}

func bindingHas(tag *ast.BasicLit, target string) bool {
	if tag == nil {
		return false
	}
	raw, err := strconv.Unquote(tag.Value)
	if err != nil {
		return false
	}
	for _, part := range strings.Split(reflect.StructTag(raw).Get("binding"), ",") {
		if strings.TrimSpace(part) == target {
			return true
		}
	}
	return false
}
