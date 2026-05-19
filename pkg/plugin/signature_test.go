package plugin

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSignAndVerifyRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://host/api/v1/plugins/registrations", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	body := []byte(`{"ok":true}`)
	if err := SignRequest(req, "demo-plugin", "01234567890123456789012345678901", body, now, "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	parts := SignatureFromHeader(req.Header)
	if err := Verify(req.Method, req.URL.EscapedPath(), req.URL.RawQuery, body, parts, "01234567890123456789012345678901", now, time.Minute); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestVerifyRejectsExpiredTimestamp(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	parts := SignatureParts{
		PluginKey: "demo-plugin",
		Timestamp: now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		Nonce:     "nonce-1",
	}
	parts.Signature = Sign(http.MethodPost, "/demo", "", parts.Timestamp, parts.Nonce, BodySHA256(nil), "01234567890123456789012345678901")

	if err := Verify(http.MethodPost, "/demo", "", nil, parts, "01234567890123456789012345678901", now, time.Minute); err == nil {
		t.Fatal("Verify() error is nil")
	}
}

func TestVerifyRejectsBodyMismatch(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	parts := SignatureParts{
		PluginKey: "demo-plugin",
		Timestamp: now.Format(time.RFC3339Nano),
		Nonce:     "nonce-1",
	}
	parts.Signature = Sign(http.MethodPost, "/demo", "", parts.Timestamp, parts.Nonce, BodySHA256([]byte("left")), "01234567890123456789012345678901")

	if err := Verify(http.MethodPost, "/demo", "", []byte("right"), parts, "01234567890123456789012345678901", now, time.Minute); err == nil {
		t.Fatal("Verify() error is nil")
	}
}

func TestSignAndVerifyIncludesRawQuery(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	body := []byte(`{"ok":true}`)
	req, err := http.NewRequest(http.MethodPost, "http://host/demo?a=1", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if err := SignRequest(req, "demo-plugin", "01234567890123456789012345678901", body, now, "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	parts := SignatureFromHeader(req.Header)
	if err := Verify(req.Method, req.URL.EscapedPath(), "a=2", body, parts, "01234567890123456789012345678901", now, time.Minute); err == nil {
		t.Fatal("Verify() error is nil for changed raw query")
	}
}

func TestReadLimitedBodyRejectsOversizedBody(t *testing.T) {
	_, err := ReadLimitedBody(strings.NewReader("too-large"), 4)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("ReadLimitedBody() error = %v, want ErrBodyTooLarge", err)
	}
}

func TestVerifySignedRequestRestoresBody(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	req, err := http.NewRequest(http.MethodPost, "http://host/demo", strings.NewReader(`{"ok":true}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	body := []byte(`{"ok":true}`)
	if err := SignRequest(req, "demo-plugin", "01234567890123456789012345678901", body, now, "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	got, parts, err := VerifySignedRequest(req, VerifyRequestOptions{
		Secret:            "01234567890123456789012345678901",
		MaxBodyBytes:      1024,
		Now:               now,
		Skew:              time.Minute,
		ExpectedPluginKey: "demo-plugin",
		NonceStore:        NewMemoryNonceStore(),
	})
	if err != nil {
		t.Fatalf("VerifySignedRequest() error = %v", err)
	}
	if string(got) != string(body) || parts.PluginKey != "demo-plugin" {
		t.Fatalf("VerifySignedRequest() body=%q parts=%#v", got, parts)
	}
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if string(restored) != string(body) {
		t.Fatalf("restored body = %q, want %q", restored, body)
	}
}

func TestVerifySignedRequestRejectsExpectedPluginKeyMismatch(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	req, err := http.NewRequest(http.MethodPost, "http://host/demo", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if err := SignRequest(req, "demo-plugin", "01234567890123456789012345678901", nil, now, "nonce-1"); err != nil {
		t.Fatalf("SignRequest() error = %v", err)
	}
	_, _, err = VerifySignedRequest(req, VerifyRequestOptions{
		Secret:            "01234567890123456789012345678901",
		MaxBodyBytes:      1024,
		Now:               now,
		Skew:              time.Minute,
		ExpectedPluginKey: "other-plugin",
	})
	if err == nil {
		t.Fatal("VerifySignedRequest() error is nil")
	}
}

func TestVerifySignedRequestRejectsReusedNonce(t *testing.T) {
	now := time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC)
	store := NewMemoryNonceStore()
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodPost, "http://host/demo", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		if err := SignRequest(req, "demo-plugin", "01234567890123456789012345678901", nil, now, "nonce-1"); err != nil {
			t.Fatalf("SignRequest() error = %v", err)
		}
		_, _, err = VerifySignedRequest(req, VerifyRequestOptions{
			Secret:            "01234567890123456789012345678901",
			MaxBodyBytes:      1024,
			Now:               now,
			Skew:              time.Minute,
			ExpectedPluginKey: "demo-plugin",
			NonceStore:        store,
		})
		if i == 0 && err != nil {
			t.Fatalf("VerifySignedRequest() first error = %v", err)
		}
		if i == 1 && err == nil {
			t.Fatal("VerifySignedRequest() second error is nil")
		}
	}
}
