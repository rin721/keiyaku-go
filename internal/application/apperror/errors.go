package apperror

import (
	"errors"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	"github.com/rin721/keiyaku-go/types"
)

const (
	CodeOK              = types.CodeOK
	CodeInvalidArgument = types.CodeInvalidArgument
	CodeUnauthorized    = types.CodeUnauthorized
	CodeForbidden       = types.CodeForbidden
	CodeNotFound        = types.CodeNotFound
	CodeConflict        = types.CodeConflict
	CodeTooManyRequests = types.CodeTooManyRequests

	CodeInvalidCredential = types.CodeInvalidCredential
	CodeUserDisabled      = types.CodeUserDisabled

	CodeInternal   = types.CodeInternal
	CodeDependency = types.CodeDependency
)

type Error = types.AppError

func New(code types.Code, msg string) *Error {
	return types.NewError(code, msg)
}

func Wrap(code types.Code, msg string, err error) *Error {
	return types.WrapError(code, msg, err)
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	switch {
	case errors.Is(err, derrors.ErrInvalidArgument):
		return Wrap(CodeInvalidArgument, types.MessageInvalidArgument, err)
	case errors.Is(err, derrors.ErrNotFound):
		return Wrap(CodeNotFound, types.MessageNotFound, err)
	case errors.Is(err, derrors.ErrConflict):
		return Wrap(CodeConflict, types.MessageConflict, err)
	case errors.Is(err, derrors.ErrUnauthorized):
		return Wrap(CodeUnauthorized, types.MessageUnauthorized, err)
	case errors.Is(err, derrors.ErrForbidden):
		return Wrap(CodeForbidden, types.MessageForbidden, err)
	case errors.Is(err, derrors.ErrInactiveUser):
		return Wrap(CodeUserDisabled, types.MessageUserDisabled, err)
	default:
		return Wrap(CodeInternal, types.MessageInternal, err)
	}
}
