package port

import (
	"context"
	"time"

	"github.com/rin721/keiyaku-go/internal/domain/article"
	"github.com/rin721/keiyaku-go/internal/domain/shared"
	"github.com/rin721/keiyaku-go/internal/domain/user"
)

type IDGenerator interface {
	NewID(ctx context.Context) (int64, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Verify(ctx context.Context, hash, plain string) (matched bool, needsRehash bool, err error)
}

type TokenUser struct {
	ID       int64
	Username string
	Roles    []string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenClaims struct {
	UserID    int64
	Username  string
	Roles     []string
	ExpiresAt time.Time
}

type TokenIssuer interface {
	IssueToken(ctx context.Context, subject TokenUser) (TokenPair, error)
	ParseAccessToken(ctx context.Context, token string) (TokenClaims, error)
}

type Authorizer interface {
	Allow(role string, object string, action string) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, entity *user.User) error
	FindByID(ctx context.Context, id int64) (*user.User, error)
	FindByUsername(ctx context.Context, username string) (*user.User, error)
}

type ArticleRepository interface {
	Create(ctx context.Context, entity *article.Article) error
	FindPublishedByID(ctx context.Context, id int64) (*article.Article, error)
	ListPublished(ctx context.Context, pagination shared.Pagination) ([]*article.Article, int64, error)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
