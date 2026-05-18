package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/dto"
	"github.com/rin721/keiyaku-go/internal/api/http/middleware"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	appuser "github.com/rin721/keiyaku-go/internal/application/user"
)

type UserHandler struct {
	service *appuser.Service
}

func NewUserHandler(service *appuser.Service) *UserHandler {
	return &UserHandler{service: service}
}

// Me handles current user profile lookup.
// @Summary Current user profile
// @Description Get the authenticated user's profile.
// @Tags User
// @Produce json
// @Security bearerAuth
// @Success 200 {object} dto.UserResponse "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "User not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageUserHandlerNotReady))
		return
	}
	claims, ok := middleware.Claims(c)
	if !ok {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageMissingAuthClaims))
		return
	}
	entity, err := h.service.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewUserResponse(entity))
}
