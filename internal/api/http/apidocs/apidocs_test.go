package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInjectRegistersSwaggerUIAndSpec(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	spec := []byte("openapi: 3.0.3\ninfo:\n  title: Demo\n")

	Inject(engine, Options{Spec: spec})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, DefaultPath, nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("docs status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("docs content-type = %q, want text/html", contentType)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "SwaggerUIBundle") || !strings.Contains(body, "openapi.yaml") {
		t.Fatalf("docs body does not include swagger bootstrap and spec path: %q", body)
	}

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, DefaultSpecPath, nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("spec status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/yaml") {
		t.Fatalf("spec content-type = %q, want application/yaml", contentType)
	}
	if got := recorder.Body.String(); got != string(spec) {
		t.Fatalf("spec body = %q, want %q", got, string(spec))
	}
}

func TestInjectSkipsDisabledDocs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	Inject(engine, Options{Disabled: true, Spec: []byte("openapi: 3.0.3\n")})

	if routes := engine.Routes(); len(routes) != 0 {
		t.Fatalf("registered routes = %d, want 0", len(routes))
	}
}

func TestNormalizeOptions(t *testing.T) {
	got := normalizeOptions(Options{
		Path:     " api-docs/ ",
		SpecPath: " spec/openapi.yaml ",
		Title:    " Custom Docs ",
	})

	if got.Path != "/api-docs" {
		t.Fatalf("path = %q, want /api-docs", got.Path)
	}
	if got.SpecPath != "/spec/openapi.yaml" {
		t.Fatalf("spec path = %q, want /spec/openapi.yaml", got.SpecPath)
	}
	if got.Title != "Custom Docs" {
		t.Fatalf("title = %q, want Custom Docs", got.Title)
	}
}
