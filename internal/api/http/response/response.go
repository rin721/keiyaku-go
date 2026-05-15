package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: apperror.CodeOK, Msg: "ok", Data: data})
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, Body{Code: apperror.CodeOK, Msg: "ok"})
}

func Error(c *gin.Context, err error) {
	appErr := apperror.From(err)
	c.JSON(statusFromCode(appErr.Code), Body{Code: appErr.Code, Msg: appErr.Msg})
}

func statusFromCode(code int) int {
	switch code {
	case apperror.CodeInvalidArgument:
		return http.StatusBadRequest
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
	default:
		if code >= 50000 {
			return http.StatusInternalServerError
		}
		return http.StatusBadRequest
	}
}
