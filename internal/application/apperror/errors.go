package apperror

import (
	"errors"
	"fmt"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

const (
	CodeOK              = 0
	CodeInvalidArgument = 10001
	CodeUnauthorized    = 10002
	CodeForbidden       = 10003
	CodeNotFound        = 10004
	CodeConflict        = 10005

	CodeInvalidCredential = 20001
	CodeUserDisabled      = 20002

	CodeInternal   = 50001
	CodeDependency = 50002
)

type Error struct {
	Code int
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Msg
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

func Wrap(code int, msg string, err error) *Error {
	return &Error{Code: code, Msg: msg, Err: err}
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
		return Wrap(CodeInvalidArgument, "invalid argument", err)
	case errors.Is(err, derrors.ErrNotFound):
		return Wrap(CodeNotFound, "resource not found", err)
	case errors.Is(err, derrors.ErrConflict):
		return Wrap(CodeConflict, "resource conflict", err)
	case errors.Is(err, derrors.ErrUnauthorized):
		return Wrap(CodeUnauthorized, "unauthorized", err)
	case errors.Is(err, derrors.ErrForbidden):
		return Wrap(CodeForbidden, "forbidden", err)
	case errors.Is(err, derrors.ErrInactiveUser):
		return Wrap(CodeUserDisabled, "user disabled", err)
	default:
		return Wrap(CodeInternal, "internal server error", err)
	}
}
