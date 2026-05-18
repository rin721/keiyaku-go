package api

import _ "embed"

// OpenAPIFileName is the embedded OpenAPI contract artifact name.
const OpenAPIFileName = "openapi.yaml"

//go:embed openapi.yaml
var openAPIYAML []byte

// OpenAPIYAML returns a copy of the embedded OpenAPI contract.
func OpenAPIYAML() []byte {
	return append([]byte(nil), openAPIYAML...)
}
