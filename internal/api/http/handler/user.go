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

// Me handles current user profile lookup for the IAM service.
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
