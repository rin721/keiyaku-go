package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/i18n"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
)

type Body struct {
	Code apperror.Code `json:"code"`
	Msg  string        `json:"msg"`
	Data interface{}   `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, apperror.CodeOK, apperror.MessageOK, data)
}

func NoContent(c *gin.Context) {
	JSON(c, http.StatusOK, apperror.CodeOK, apperror.MessageOK, nil)
}

func Error(c *gin.Context, err error) {
	appErr := apperror.From(err)
	JSON(c, statusForCode(appErr.Code), appErr.Code, appErr.Msg, nil)
}

func JSON(c *gin.Context, status int, code apperror.Code, msg string, data interface{}) {
	if msg == "" {
		msg = apperror.Message(code)
	}
	c.JSON(status, Body{Code: code, Msg: i18n.Message(c, code, msg), Data: data})
}

func statusForCode(code apperror.Code) int {
	switch code {
	case apperror.CodeOK:
		return http.StatusOK
	case apperror.CodeInvalidArgument, apperror.CodeTooManyRequests:
		return http.StatusBadRequest
	case apperror.CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeInvalidCredential, apperror.CodeUserDisabled:
		return http.StatusUnprocessableEntity
	case apperror.CodeBadGateway:
		return http.StatusBadGateway
	case apperror.CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case apperror.CodeGatewayTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
