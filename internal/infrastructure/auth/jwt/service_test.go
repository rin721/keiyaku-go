package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/port"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
)

func TestRefreshTokenCarriesTokenID(t *testing.T) {
	service := NewService(config.JWTConfig{
		Secret:          "01234567890123456789012345678901",
		Issuer:          "keiyaku-test",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: time.Hour,
	})
	pair, err := service.IssueTokenWithRefreshID(context.Background(), port.TokenUser{
		ID:       100,
		Username: "alice",
		Roles:    []string{"author"},
	}, "refresh-session-1")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	claims, err := service.ParseRefreshToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if claims.TokenID != "refresh-session-1" {
		t.Fatalf("unexpected refresh token id: %q", claims.TokenID)
	}
	if !pair.RefreshExpiresAt.After(pair.ExpiresAt) {
		t.Fatal("expected refresh token to expire after access token")
	}
}
