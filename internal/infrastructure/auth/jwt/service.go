package jwt

import (
	"context"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/rin721/keiyaku-go/internal/application/port"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
)

type Service struct {
	cfg config.JWTConfig
	now func() time.Time
}

type claims struct {
	UserID   int64    `json:"uid"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwtlib.RegisteredClaims
}

func NewService(cfg config.JWTConfig) *Service {
	return &Service{cfg: cfg, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) IssueToken(ctx context.Context, subject port.TokenUser) (port.TokenPair, error) {
	return s.IssueTokenWithRefreshID(ctx, subject, "")
}

func (s *Service) IssueTokenWithRefreshID(_ context.Context, subject port.TokenUser, refreshTokenID string) (port.TokenPair, error) {
	if s == nil || s.cfg.Secret == "" {
		return port.TokenPair{}, fmt.Errorf("jwt service is not ready")
	}
	now := s.now()
	accessExpiresAt := now.Add(s.cfg.AccessTokenTTL)
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)
	access, err := s.sign(subject, accessExpiresAt, "access", "")
	if err != nil {
		return port.TokenPair{}, err
	}
	refresh, err := s.sign(subject, refreshExpiresAt, "refresh", refreshTokenID)
	if err != nil {
		return port.TokenPair{}, err
	}
	return port.TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Service) ParseAccessToken(_ context.Context, raw string) (port.TokenClaims, error) {
	return s.parseToken(raw, "access")
}

func (s *Service) ParseRefreshToken(_ context.Context, raw string) (port.TokenClaims, error) {
	return s.parseToken(raw, "refresh")
}

func (s *Service) parseToken(raw string, audience string) (port.TokenClaims, error) {
	if s == nil || s.cfg.Secret == "" {
		return port.TokenClaims{}, fmt.Errorf("jwt service is not ready")
	}
	parsed := claims{}
	token, err := jwtlib.ParseWithClaims(raw, &parsed, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected jwt method: %s", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	}, jwtlib.WithIssuer(s.cfg.Issuer), jwtlib.WithAudience(audience))
	if err != nil {
		return port.TokenClaims{}, fmt.Errorf("parse jwt: %w", err)
	}
	if !token.Valid {
		return port.TokenClaims{}, fmt.Errorf("invalid jwt")
	}
	return port.TokenClaims{
		UserID:    parsed.UserID,
		Username:  parsed.Username,
		Roles:     parsed.Roles,
		ExpiresAt: parsed.ExpiresAt.Time,
		TokenID:   parsed.ID,
	}, nil
}

func (s *Service) sign(subject port.TokenUser, expiresAt time.Time, audience string, tokenID string) (string, error) {
	now := s.now()
	claims := claims{
		UserID:   subject.ID,
		Username: subject.Username,
		Roles:    append([]string(nil), subject.Roles...),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   fmt.Sprintf("%d", subject.ID),
			Audience:  []string{audience},
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ID:        tokenID,
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}
