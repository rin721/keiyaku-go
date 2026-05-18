package openapi

const (
	DefaultHandlerDir = "internal/api/http/handler"
	DefaultDTODir     = "internal/api/http/dto"
	DefaultOutputPath = "api/openapi.yaml"
	DefaultBasePath   = "/api/v1"
	DefaultTitle      = "Keiyaku-Go CMS API"
	DefaultVersion    = "1.0.0"
)

type Config struct {
	HandlerDir string
	DTODir     string
	Output     string
	BasePath   string
	Title      string
	Version    string
	Check      bool
}

type Document struct {
	Title    string
	Version  string
	BasePath string
	Paths    map[string]map[string]Operation
	Schemas  map[string]*Schema
}

type Operation struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Tags        []string
	Accepts     []string
	Produces    []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   []Response
	Security    []string
}

type Parameter struct {
	Name        string
	In          string
	TypeName    string
	Required    bool
	Description string
	Schema      *Schema
}

type RequestBody struct {
	Description string
	Required    bool
	Schema      *Schema
}

type Response struct {
	Code        string
	Description string
	Schema      *Schema
}

type Schema struct {
	Ref        string
	Type       string
	Format     string
	Nullable   bool
	Items      *Schema
	Properties []Property
	Required   []string
	MinLength  *int
	MaxLength  *int
	Minimum    *int
	Maximum    *int
}

type Property struct {
	Name   string
	Schema *Schema
}
