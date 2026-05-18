package apperror

import (
	"errors"
	"fmt"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

type Code int

const (
	CodeOK Code = 0
)

const (
	CodeInvalidArgument Code = 10001
	CodeUnauthorized    Code = 10002
	CodeForbidden       Code = 10003
	CodeNotFound        Code = 10004
	CodeConflict        Code = 10005
	CodeTooManyRequests Code = 10006
)

const (
	CodeInvalidCredential Code = 20001
	CodeUserDisabled      Code = 20002
)

const (
	CodeInternal   Code = 50001
	CodeDependency Code = 50002
)

const (
	MessageOK                  = "ok"
	MessageInvalidArgument     = "invalid argument"
	MessageUnauthorized        = "unauthorized"
	MessageForbidden           = "forbidden"
	MessageNotFound            = "resource not found"
	MessageConflict            = "resource conflict"
	MessageTooManyRequests     = "too many requests"
	MessageInvalidCredential   = "invalid username or password"
	MessageUserDisabled        = "user disabled"
	MessageInternal            = "internal server error"
	MessageDependency          = "dependency error"
	MessageRouteNotFound       = "route not found"
	MessageServiceUnavailable  = "service temporarily unavailable"
	MessageMissingAuthClaims   = "missing auth claims"
	MessageInvalidRequestBody  = "invalid request body"
	MessageInvalidAccessToken  = "invalid access token"
	MessageMissingAuthHeader   = "missing authorization header"
	MessageInvalidAuthScheme   = "invalid authorization scheme"
	MessagePermissionDenied    = "permission denied"
	MessagePermissionNotReady  = "permission service is not ready"
	MessagePermissionCheckFail = "permission check failed"

	MessageAuthServiceNotReady    = "auth service is not ready"
	MessageUserServiceNotReady    = "user service is not ready"
	MessageArticleServiceNotReady = "article service is not ready"
	MessageAuthHandlerNotReady    = "auth handler is not ready"
	MessageUserHandlerNotReady    = "user handler is not ready"
	MessageArticleHandlerNotReady = "article handler is not ready"

	MessageInvalidUserID    = "invalid user id"
	MessageInvalidArticleID = "invalid article id"
	MessageMissingUser      = "missing authenticated user"
	MessagePasswordLength   = "password length must be between 8 and 128"
	MessageUsernameExists   = "username already exists"

	MessageCheckUserFailed      = "failed to check user"
	MessageLoadUserFailed       = "failed to load user"
	MessageCreateUserFailed     = "failed to create user"
	MessageAllocateUserIDFailed = "failed to allocate user id"
	MessageHashPasswordFailed   = "failed to hash password" // #nosec G101 -- user-facing failure message, not a credential.
	MessageVerifyPasswordFailed = "failed to verify password"
	MessageIssueTokenFailed     = "failed to issue token"

	MessageAllocateArticleIDFailed = "failed to allocate article id"
	MessageCreateArticleFailed     = "failed to create article"
	MessageListArticlesFailed      = "failed to list articles"
)

var defaultMessages = map[Code]string{
	CodeOK:                MessageOK,
	CodeInvalidArgument:   MessageInvalidArgument,
	CodeUnauthorized:      MessageUnauthorized,
	CodeForbidden:         MessageForbidden,
	CodeNotFound:          MessageNotFound,
	CodeConflict:          MessageConflict,
	CodeTooManyRequests:   MessageTooManyRequests,
	CodeInvalidCredential: MessageInvalidCredential,
	CodeUserDisabled:      MessageUserDisabled,
	CodeInternal:          MessageInternal,
	CodeDependency:        MessageDependency,
}

type Error struct {
	Code Code
	Msg  string
	Err  error
}

func New(code Code, msg string) *Error {
	if msg == "" {
		msg = Message(code)
	}
	return &Error{Code: code, Msg: msg}
}

func Wrap(code Code, msg string, err error) *Error {
	if msg == "" {
		msg = Message(code)
	}
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
		return Wrap(CodeInvalidArgument, MessageInvalidArgument, err)
	case errors.Is(err, derrors.ErrNotFound):
		return Wrap(CodeNotFound, MessageNotFound, err)
	case errors.Is(err, derrors.ErrConflict):
		return Wrap(CodeConflict, MessageConflict, err)
	case errors.Is(err, derrors.ErrUnauthorized):
		return Wrap(CodeUnauthorized, MessageUnauthorized, err)
	case errors.Is(err, derrors.ErrForbidden):
		return Wrap(CodeForbidden, MessageForbidden, err)
	case errors.Is(err, derrors.ErrInactiveUser):
		return Wrap(CodeUserDisabled, MessageUserDisabled, err)
	default:
		return Wrap(CodeInternal, MessageInternal, err)
	}
}

func Message(code Code) string {
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return MessageInternal
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
