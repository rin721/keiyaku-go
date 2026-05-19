package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
)

func TestNormalizeOptionsDefaults(t *testing.T) {
	got := normalizeOptions(Options{})

	if got.RateLimit.RequestsPerSecond != 100 {
		t.Fatalf("default rate limit = %v, want 100", got.RateLimit.RequestsPerSecond)
	}
	if got.RateLimit.Burst != 100 {
		t.Fatalf("default rate limit burst = %d, want 100", got.RateLimit.Burst)
	}
	if got.CircuitBreaker.Name != "http-api" {
		t.Fatalf("default circuit breaker name = %q, want http-api", got.CircuitBreaker.Name)
	}
	if got.CircuitBreaker.FailureThreshold != 5 {
		t.Fatalf("default failure threshold = %d, want 5", got.CircuitBreaker.FailureThreshold)
	}
	if got.CircuitBreaker.OpenTimeout != 5*time.Second {
		t.Fatalf("default open timeout = %s, want 5s", got.CircuitBreaker.OpenTimeout)
	}
}

func TestNewWithZeroOptionsRegistersHealthz(t *testing.T) {
	engine := New(Deps{})
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body response.Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != apperror.CodeOK {
		t.Fatalf("code = %d, want %d", body.Code, apperror.CodeOK)
	}
}

func TestNewInjectsAPIDocsByDefault(t *testing.T) {
	engine := New(Deps{})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("docs status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("docs body does not include Swagger UI bootstrap: %q", body)
	}

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "openapi: 3.0.3") {
		t.Fatalf("openapi body does not include embedded contract: %q", body)
	}
}

func TestNewCanDisableAPIDocs(t *testing.T) {
	engine := New(Deps{
		Options: Options{
			APIDocs: APIDocsOptions{Disabled: true},
		},
	})
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestNewDoesNotRegisterLegacyArticleRoutes(t *testing.T) {
	engine := New(Deps{})
	for _, target := range []string{"/api/v1/articles", "/api/v1/articles/1"} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		engine.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", target, recorder.Code, http.StatusNotFound)
		}
	}
}
