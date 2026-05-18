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

// Register handles user registration.
// @Summary Register user
// @Description Register a user and return access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register payload"
// @Success 200 {object} dto.AuthResponse "OK"
// @Failure 400 {object} response.Body "Invalid request"
// @Failure 409 {object} response.Body "Conflict"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageAuthHandlerNotReady))
		return
	}
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
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

// Login handles user login.
// @Summary Login user
// @Description Authenticate a user and return access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.AuthResponse "OK"
// @Failure 400 {object} response.Body "Invalid request"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 422 {object} response.Body "Invalid credential"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageAuthHandlerNotReady))
		return
	}
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
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
