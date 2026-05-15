package user

import (
	"testing"
	"time"
)

func TestNewUserValidatesAndDefaults(t *testing.T) {
	now := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	got, err := New(1, "rin_721", "rin@example.com", "hash", "", now)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got.DisplayName != "rin_721" {
		t.Fatalf("DisplayName = %q", got.DisplayName)
	}
	if got.Status != StatusActive {
		t.Fatalf("Status = %q", got.Status)
	}
	if len(got.Roles) != 1 || got.Roles[0] != "author" {
		t.Fatalf("Roles = %#v", got.Roles)
	}
}

func TestNewUserRejectsInvalidUsername(t *testing.T) {
	if _, err := New(1, "no", "rin@example.com", "hash", "", time.Now()); err == nil {
		t.Fatal("expected invalid username error")
	}
}
