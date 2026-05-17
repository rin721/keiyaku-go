package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/i18n"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
)

type Body = types.Response

func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, types.CodeOK, types.MessageOK, data)
}

func NoContent(c *gin.Context) {
	JSON(c, http.StatusOK, types.CodeOK, types.MessageOK, nil)
}

func Error(c *gin.Context, err error) {
	appErr := apperror.From(err)
	JSON(c, appErr.Code.HTTPStatus(), appErr.Code, appErr.Msg, nil)
}

func JSON(c *gin.Context, status int, code types.Code, msg string, data interface{}) {
	c.JSON(status, types.NewResponse(code, i18n.Message(c, code, msg), data))
}
