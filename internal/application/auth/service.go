package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	"github.com/rin721/keiyaku-go/internal/domain/user"
)

type Service struct {
	users  port.UserRepository
	ids    port.IDGenerator
	hashes port.PasswordHasher
	tokens port.TokenIssuer
	now    func() time.Time
}

func NewService(users port.UserRepository, ids port.IDGenerator, hashes port.PasswordHasher, tokens port.TokenIssuer) *Service {
	return &Service{users: users, ids: ids, hashes: hashes, tokens: tokens, now: func() time.Time { return time.Now().UTC() }}
}

type RegisterCommand struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

type LoginCommand struct {
	Username string
	Password string
}

type Result struct {
	User  *user.User
	Token port.TokenPair
}

func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (*Result, error) {
	if s == nil || s.users == nil || s.ids == nil || s.hashes == nil || s.tokens == nil {
		return nil, apperror.New(apperror.CodeInternal, "auth service is not ready")
	}
	if len(cmd.Password) < 8 || len(cmd.Password) > 128 {
		return nil, apperror.New(apperror.CodeInvalidArgument, "password length must be between 8 and 128")
	}
	username := strings.TrimSpace(cmd.Username)
	if _, err := s.users.FindByUsername(ctx, username); err == nil {
		return nil, apperror.New(apperror.CodeConflict, "username already exists")
	} else if !errors.Is(err, derrors.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to check user", err)
	}

	id, err := s.ids.NewID(ctx)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to allocate user id", err)
	}
	hash, err := s.hashes.Hash(ctx, cmd.Password)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to hash password", err)
	}
	entity, err := user.New(id, username, cmd.Email, hash, cmd.DisplayName, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, entity); err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to create user", err)
	}
	token, err := s.tokens.IssueToken(ctx, port.TokenUser{ID: entity.ID, Username: entity.Username, Roles: entity.Roles})
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to issue token", err)
	}
	return &Result{User: entity, Token: token}, nil
}

func (s *Service) Login(ctx context.Context, cmd LoginCommand) (*Result, error) {
	if s == nil || s.users == nil || s.hashes == nil || s.tokens == nil {
		return nil, apperror.New(apperror.CodeInternal, "auth service is not ready")
	}
	entity, err := s.users.FindByUsername(ctx, strings.TrimSpace(cmd.Username))
	if err != nil {
		if errors.Is(err, derrors.ErrNotFound) {
			return nil, apperror.New(apperror.CodeInvalidCredential, "invalid username or password")
		}
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to load user", err)
	}
	if err := entity.EnsureActive(); err != nil {
		return nil, err
	}
	matched, _, err := s.hashes.Verify(ctx, entity.PasswordHash, cmd.Password)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to verify password", err)
	}
	if !matched {
		return nil, apperror.New(apperror.CodeInvalidCredential, "invalid username or password")
	}
	token, err := s.tokens.IssueToken(ctx, port.TokenUser{ID: entity.ID, Username: entity.Username, Roles: entity.Roles})
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to issue token", err)
	}
	return &Result{User: entity, Token: token}, nil
}
