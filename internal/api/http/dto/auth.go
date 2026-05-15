package dto

import (
	"time"

	domainarticle "github.com/rin721/keiyaku-go/internal/domain/article"
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

type CreateArticleRequest struct {
	CategoryID int64    `json:"category_id"`
	Title      string   `json:"title" binding:"required,max=160"`
	Slug       string   `json:"slug" binding:"required,max=180"`
	Summary    string   `json:"summary" binding:"omitempty,max=512"`
	Content    string   `json:"content" binding:"required"`
	Tags       []string `json:"tags"`
	Publish    bool     `json:"publish"`
}

type ArticleResponse struct {
	ID          int64      `json:"id"`
	AuthorID    int64      `json:"author_id"`
	CategoryID  int64      `json:"category_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Summary     string     `json:"summary"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ArticleListResponse struct {
	Items []ArticleResponse `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
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

func NewArticleResponse(entity *domainarticle.Article) ArticleResponse {
	if entity == nil {
		return ArticleResponse{}
	}
	return ArticleResponse{
		ID:          entity.ID,
		AuthorID:    entity.AuthorID,
		CategoryID:  entity.CategoryID,
		Title:       entity.Title,
		Slug:        entity.Slug,
		Summary:     entity.Summary,
		Content:     entity.Content,
		Status:      string(entity.Status),
		Tags:        append([]string(nil), entity.Tags...),
		PublishedAt: entity.PublishedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
