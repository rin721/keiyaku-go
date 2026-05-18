package apidocs

// Options controls Swagger UI and OpenAPI route injection.
type Options struct {
	Disabled bool
	Path     string
	SpecPath string
	Title    string
	Spec     []byte
}
