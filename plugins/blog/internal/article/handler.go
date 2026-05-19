package article

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/articles", h.Create)
	router.GET("/articles", h.List)
	router.GET("/articles/:id", h.Get)
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
	Content     string     `json:"content,omitempty"`
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

type responseBody struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (h *Handler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		writeError(c, http.StatusInternalServerError, "article handler is not ready")
		return
	}
	authorID, ok := userIDFromHeader(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	entity, err := h.service.Create(c.Request.Context(), CreateCommand{
		AuthorID:   authorID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Slug:       req.Slug,
		Summary:    req.Summary,
		Content:    req.Content,
		Tags:       req.Tags,
		Publish:    req.Publish,
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	writeOK(c, newArticleResponse(entity, true))
}

func (h *Handler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		writeError(c, http.StatusInternalServerError, "article handler is not ready")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, "invalid article id")
		return
	}
	entity, err := h.service.GetPublished(c.Request.Context(), id)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	writeOK(c, newArticleResponse(entity, true))
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		writeError(c, http.StatusInternalServerError, "article handler is not ready")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.service.ListPublished(c.Request.Context(), ListQuery{Page: page, PageSize: size})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	items := make([]ArticleResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, newArticleResponse(item, false))
	}
	writeOK(c, ArticleListResponse{Items: items, Total: result.Total, Page: result.Page, Size: result.Size})
}

func userIDFromHeader(c *gin.Context) (int64, bool) {
	raw := c.GetHeader("X-Keiyaku-User-ID")
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

func newArticleResponse(entity *Article, includeContent bool) ArticleResponse {
	if entity == nil {
		return ArticleResponse{}
	}
	response := ArticleResponse{
		ID:          entity.ID,
		AuthorID:    entity.AuthorID,
		CategoryID:  entity.CategoryID,
		Title:       entity.Title,
		Slug:        entity.Slug,
		Summary:     entity.Summary,
		Status:      string(entity.Status),
		Tags:        append([]string(nil), entity.Tags...),
		PublishedAt: entity.PublishedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
	if includeContent {
		response.Content = entity.Content
	}
	return response
}

func writeOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, responseBody{Code: 0, Msg: "ok", Data: data})
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, responseBody{Code: status, Msg: msg})
}

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidArgument):
		writeError(c, http.StatusBadRequest, "invalid argument")
	case errors.Is(err, ErrMissingUser):
		writeError(c, http.StatusUnauthorized, "missing authenticated user")
	case errors.Is(err, ErrConflict):
		writeError(c, http.StatusConflict, "resource conflict")
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, "resource not found")
	default:
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
}
