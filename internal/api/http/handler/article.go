package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/dto"
	"github.com/rin721/keiyaku-go/internal/api/http/middleware"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	apparticle "github.com/rin721/keiyaku-go/internal/application/article"
)

type ArticleHandler struct {
	service *apparticle.Service
}

func NewArticleHandler(service *apparticle.Service) *ArticleHandler {
	return &ArticleHandler{service: service}
}

func (h *ArticleHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageArticleHandlerNotReady))
		return
	}
	claims, ok := middleware.Claims(c)
	if !ok {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageMissingAuthClaims))
		return
	}
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
		return
	}
	entity, err := h.service.Create(c.Request.Context(), apparticle.CreateCommand{
		AuthorID:   claims.UserID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Slug:       req.Slug,
		Summary:    req.Summary,
		Content:    req.Content,
		Tags:       req.Tags,
		Publish:    req.Publish,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewArticleResponse(entity))
}

func (h *ArticleHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageArticleHandlerNotReady))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidArticleID))
		return
	}
	entity, err := h.service.GetPublished(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewArticleResponse(entity))
}

func (h *ArticleHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessageArticleHandlerNotReady))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.service.ListPublished(c.Request.Context(), apparticle.ListQuery{Page: page, PageSize: size})
	if err != nil {
		response.Error(c, err)
		return
	}
	items := make([]dto.ArticleResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.NewArticleResponse(item))
	}
	response.OK(c, dto.ArticleListResponse{Items: items, Total: result.Total, Page: result.Page, Size: result.Size})
}
