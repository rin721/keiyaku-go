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
	if err := Verify(req.Method, req.URL.EscapedPath(), body, parts, "01234567890123456789012345678901", now, time.Minute); err != nil {
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
	parts.Signature = Sign(http.MethodPost, "/demo", parts.Timestamp, parts.Nonce, BodySHA256(nil), "01234567890123456789012345678901")

	if err := Verify(http.MethodPost, "/demo", nil, parts, "01234567890123456789012345678901", now, time.Minute); err == nil {
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
	parts.Signature = Sign(http.MethodPost, "/demo", parts.Timestamp, parts.Nonce, BodySHA256([]byte("left")), "01234567890123456789012345678901")

	if err := Verify(http.MethodPost, "/demo", []byte("right"), parts, "01234567890123456789012345678901", now, time.Minute); err == nil {
		t.Fatal("Verify() error is nil")
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
	got, parts, err := VerifySignedRequest(req, "01234567890123456789012345678901", 1024, now, time.Minute)
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
