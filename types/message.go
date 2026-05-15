package types

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
	MessageHashPasswordFailed   = "failed to hash password"
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

func Message(code Code) string {
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return MessageInternal
}
