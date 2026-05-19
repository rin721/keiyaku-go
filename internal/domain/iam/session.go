package iam

import (
	"strings"
	"time"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

type RefreshSessionStatus string

const (
	RefreshSessionActive  RefreshSessionStatus = "active"
	RefreshSessionRotated RefreshSessionStatus = "rotated"
	RefreshSessionRevoked RefreshSessionStatus = "revoked"
)

type RefreshSession struct {
	ID                  int64
	UserID              int64
	RefreshTokenID      string
	Status              RefreshSessionStatus
	ReplacedBySessionID int64
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewRefreshSession(id int64, userID int64, refreshTokenID string, expiresAt time.Time, now time.Time) (*RefreshSession, error) {
	refreshTokenID = strings.TrimSpace(refreshTokenID)
	if id <= 0 || userID <= 0 || refreshTokenID == "" || expiresAt.IsZero() {
		return nil, derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &RefreshSession{
		ID:             id,
		UserID:         userID,
		RefreshTokenID: refreshTokenID,
		Status:         RefreshSessionActive,
		ExpiresAt:      expiresAt.UTC(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (s *RefreshSession) Usable(now time.Time) bool {
	if s == nil {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.Status == RefreshSessionActive && s.RevokedAt == nil && s.ExpiresAt.After(now)
}
