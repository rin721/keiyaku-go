package user

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	Status       Status
	Roles        []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(id int64, username, email, passwordHash, displayName string, now time.Time) (*User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)

	if id <= 0 || !ValidUsername(username) || !validEmail(email) || passwordHash == "" {
		return nil, derrors.ErrInvalidArgument
	}
	if displayName == "" {
		displayName = username
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return &User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Status:       StatusActive,
		Roles:        []string{"author"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ValidUsername(username string) bool {
	n := utf8.RuneCountInString(username)
	if n < 3 || n > 32 {
		return false
	}
	for _, r := range username {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func (u *User) EnsureActive() error {
	if u == nil {
		return derrors.ErrNotFound
	}
	if u.Status != StatusActive {
		return derrors.ErrInactiveUser
	}
	return nil
}

func validEmail(value string) bool {
	if value == "" || len(value) > 254 {
		return false
	}
	_, err := mail.ParseAddress(value)
	return err == nil
}
