package article

import (
	"context"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	domainarticle "github.com/rin721/keiyaku-go/internal/domain/article"
	"github.com/rin721/keiyaku-go/internal/domain/shared"
)

type Service struct {
	articles port.ArticleRepository
	ids      port.IDGenerator
	now      func() time.Time
}

func NewService(articles port.ArticleRepository, ids port.IDGenerator) *Service {
	return &Service{articles: articles, ids: ids, now: func() time.Time { return time.Now().UTC() }}
}

type CreateCommand struct {
	AuthorID   int64
	CategoryID int64
	Title      string
	Slug       string
	Summary    string
	Content    string
	Tags       []string
	Publish    bool
}

type ListQuery struct {
	Page     int
	PageSize int
}

type ListResult struct {
	Items []*domainarticle.Article
	Total int64
	Page  int
	Size  int
}

func (s *Service) Create(ctx context.Context, cmd CreateCommand) (*domainarticle.Article, error) {
	if s == nil || s.articles == nil || s.ids == nil {
		return nil, apperror.New(apperror.CodeInternal, "article service is not ready")
	}
	if cmd.AuthorID <= 0 {
		return nil, apperror.New(apperror.CodeUnauthorized, "missing authenticated user")
	}
	id, err := s.ids.NewID(ctx)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to allocate article id", err)
	}
	entity, err := domainarticle.New(id, cmd.AuthorID, cmd.CategoryID, cmd.Title, cmd.Slug, cmd.Summary, cmd.Content, cmd.Tags, s.now())
	if err != nil {
		return nil, err
	}
	if cmd.Publish {
		if err := entity.Publish(s.now()); err != nil {
			return nil, err
		}
	}
	if err := s.articles.Create(ctx, entity); err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to create article", err)
	}
	return entity, nil
}

func (s *Service) GetPublished(ctx context.Context, id int64) (*domainarticle.Article, error) {
	if s == nil || s.articles == nil {
		return nil, apperror.New(apperror.CodeInternal, "article service is not ready")
	}
	if id <= 0 {
		return nil, apperror.New(apperror.CodeInvalidArgument, "invalid article id")
	}
	return s.articles.FindPublishedByID(ctx, id)
}

func (s *Service) ListPublished(ctx context.Context, query ListQuery) (*ListResult, error) {
	if s == nil || s.articles == nil {
		return nil, apperror.New(apperror.CodeInternal, "article service is not ready")
	}
	pagination := shared.NewPagination(query.Page, query.PageSize)
	items, total, err := s.articles.ListPublished(ctx, pagination)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, "failed to list articles", err)
	}
	return &ListResult{Items: items, Total: total, Page: pagination.Page, Size: pagination.PageSize}, nil
}
