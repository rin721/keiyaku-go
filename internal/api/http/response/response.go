package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/types"
)

type Body = types.Response

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, types.OK(data))
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, types.EmptyOK())
}

func Error(c *gin.Context, err error) {
	appErr := apperror.From(err)
	c.JSON(appErr.Code.HTTPStatus(), types.NewResponse(appErr.Code, appErr.Msg, nil))
}
