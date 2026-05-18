package apperror

import (
	"errors"
	"testing"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

func TestMessageUsesDefaultMessage(t *testing.T) {
	if got := Message(CodeNotFound); got != MessageNotFound {
		t.Fatalf("Message(CodeNotFound) = %q, want %q", got, MessageNotFound)
	}
}

func TestFromMapsDomainErrors(t *testing.T) {
	got := From(derrors.ErrNotFound)
	if got.Code != CodeNotFound {
		t.Fatalf("Code = %d, want %d", got.Code, CodeNotFound)
	}
	if !errors.Is(got, derrors.ErrNotFound) {
		t.Fatalf("From() should wrap the original domain error")
	}
}
