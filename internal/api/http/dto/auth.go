package dto

import (
	"time"

	domainuser "github.com/rin721/keiyaku-go/internal/domain/user"
)

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=32"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	DisplayName string `json:"display_name" binding:"omitempty,max=64"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

type UserResponse struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewUserResponse(entity *domainuser.User) UserResponse {
	if entity == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:          entity.ID,
		Username:    entity.Username,
		Email:       entity.Email,
		DisplayName: entity.DisplayName,
		Status:      string(entity.Status),
		Roles:       append([]string(nil), entity.Roles...),
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
