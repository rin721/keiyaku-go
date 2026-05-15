package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/dto"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/auth"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, "auth handler is not ready"))
		return
	}
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, "invalid request body", err))
		return
	}
	result, err := h.service.Register(c.Request.Context(), auth.RegisterCommand{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.AuthResponse{
		User:         dto.NewUserResponse(result.User),
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		ExpiresAt:    result.Token.ExpiresAt,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, "auth handler is not ready"))
		return
	}
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, "invalid request body", err))
		return
	}
	result, err := h.service.Login(c.Request.Context(), auth.LoginCommand{Username: req.Username, Password: req.Password})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.AuthResponse{
		User:         dto.NewUserResponse(result.User),
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		ExpiresAt:    result.Token.ExpiresAt,
	})
}
