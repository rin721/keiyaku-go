package auth

import (
	"context"
	"testing"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/port"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainuser "github.com/rin721/keiyaku-go/internal/domain/user"
)

func TestServiceRegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	service := NewService(repo, fakeIDGenerator{next: 1001}, fakeHasher{}, fakeTokenIssuer{})

	registered, err := service.Register(ctx, RegisterCommand{
		Username:    "author_1",
		Email:       "author@example.com",
		Password:    "password-123",
		DisplayName: "Author",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.User.ID != 1001 {
		t.Fatalf("registered user id = %d", registered.User.ID)
	}
	if registered.Token.AccessToken == "" {
		t.Fatal("access token is empty")
	}

	loggedIn, err := service.Login(ctx, LoginCommand{Username: "author_1", Password: "password-123"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loggedIn.User.Username != "author_1" {
		t.Fatalf("logged in username = %q", loggedIn.User.Username)
	}
}

func TestServiceLoginRejectsBadPassword(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	service := NewService(repo, fakeIDGenerator{next: 1001}, fakeHasher{}, fakeTokenIssuer{})
	if _, err := service.Register(ctx, RegisterCommand{Username: "author_1", Email: "author@example.com", Password: "password-123"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := service.Login(ctx, LoginCommand{Username: "author_1", Password: "wrong-password"}); err == nil {
		t.Fatal("expected invalid credential error")
	}
}

type fakeIDGenerator struct {
	next int64
}

func (g fakeIDGenerator) NewID(context.Context) (int64, error) {
	return g.next, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(_ context.Context, plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakeHasher) Verify(_ context.Context, hash, plain string) (bool, bool, error) {
	return hash == "hashed:"+plain, false, nil
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) IssueToken(_ context.Context, subject port.TokenUser) (port.TokenPair, error) {
	return port.TokenPair{
		AccessToken:  "access:" + subject.Username,
		RefreshToken: "refresh:" + subject.Username,
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (fakeTokenIssuer) ParseAccessToken(context.Context, string) (port.TokenClaims, error) {
	return port.TokenClaims{}, nil
}

type fakeUserRepo struct {
	byID map[int64]*domainuser.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: make(map[int64]*domainuser.User)}
}

func (r *fakeUserRepo) Create(_ context.Context, entity *domainuser.User) error {
	if _, exists := r.byID[entity.ID]; exists {
		return derrors.ErrConflict
	}
	r.byID[entity.ID] = entity
	return nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id int64) (*domainuser.User, error) {
	entity, ok := r.byID[id]
	if !ok {
		return nil, derrors.ErrNotFound
	}
	return entity, nil
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (*domainuser.User, error) {
	for _, entity := range r.byID {
		if entity.Username == username {
			return entity, nil
		}
	}
	return nil, derrors.ErrNotFound
}
