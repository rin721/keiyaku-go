package password

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(_ context.Context, plain string) (string, error) {
	if h == nil {
		return "", fmt.Errorf("bcrypt hasher is not ready")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

func (h *BcryptHasher) Verify(_ context.Context, hash, plain string) (bool, bool, error) {
	if h == nil {
		return false, false, fmt.Errorf("bcrypt hasher is not ready")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, false, nil
		}
		return false, false, fmt.Errorf("bcrypt verify: %w", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true, false, fmt.Errorf("bcrypt cost: %w", err)
	}
	return true, cost != h.cost, nil
}
