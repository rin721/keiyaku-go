package iam

import (
	"testing"
	"time"
)

func TestRefreshSessionUsable(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	session, err := NewRefreshSession(1, 100, "refresh-1", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("new refresh session: %v", err)
	}
	if !session.Usable(now) {
		t.Fatal("expected active unexpired session to be usable")
	}
	session.Status = RefreshSessionRotated
	if session.Usable(now) {
		t.Fatal("expected rotated session to be unusable")
	}
}

func TestRefreshSessionExpired(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	session, err := NewRefreshSession(1, 100, "refresh-1", now.Add(-time.Minute), now)
	if err != nil {
		t.Fatalf("new refresh session: %v", err)
	}
	if session.Usable(now) {
		t.Fatal("expected expired session to be unusable")
	}
}
