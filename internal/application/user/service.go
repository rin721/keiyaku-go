package user

import (
	"context"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	domainuser "github.com/rin721/keiyaku-go/internal/domain/user"
)

type Service struct {
	users port.UserRepository
}

func NewService(users port.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) GetProfile(ctx context.Context, id int64) (*domainuser.User, error) {
	if s == nil || s.users == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessageUserServiceNotReady)
	}
	if id <= 0 {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidUserID)
	}
	entity, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return entity, entity.EnsureActive()
}
