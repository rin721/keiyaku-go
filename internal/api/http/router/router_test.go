package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
