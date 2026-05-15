package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/rin721/keiyaku-go/types"
)

type contextKey struct{}

const HeaderName = types.HeaderTraceID

func WithID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = NewID()
	}
	return context.WithValue(ctx, contextKey{}, id)
}

func IDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(b[:])
}
