package article

import (
	"context"
	"fmt"
	"time"
)

type Repository interface {
	Create(ctx context.Context, entity *Article, revision Revision) error
	FindPublishedByID(ctx context.Context, id int64) (*Article, error)
	ListPublished(ctx context.Context, pagination Pagination) ([]*Article, int64, error)
}

type IDGenerator interface {
	NewID(ctx context.Context) (int64, error)
}

type Service struct {
	articles Repository
	ids      IDGenerator
	now      func() time.Time
}

func NewService(articles Repository, ids IDGenerator) *Service {
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
	Items []*Article
	Total int64
	Page  int
	Size  int
}

func (s *Service) Create(ctx context.Context, cmd CreateCommand) (*Article, error) {
	if s == nil || s.articles == nil || s.ids == nil {
		return nil, fmt.Errorf("article service is not ready")
	}
	if cmd.AuthorID <= 0 {
		return nil, ErrMissingUser
	}
	id, err := s.ids.NewID(ctx)
	if err != nil {
		return nil, fmt.Errorf("allocate article id: %w", err)
	}
	entity, err := New(id, cmd.AuthorID, cmd.CategoryID, cmd.Title, cmd.Slug, cmd.Summary, cmd.Content, cmd.Tags, s.now())
	if err != nil {
		return nil, err
	}
	if cmd.Publish {
		if err := entity.Publish(s.now()); err != nil {
			return nil, err
		}
	}
	if err := s.articles.Create(ctx, entity, entity.FirstRevision()); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) GetPublished(ctx context.Context, id int64) (*Article, error) {
	if s == nil || s.articles == nil {
		return nil, fmt.Errorf("article service is not ready")
	}
	if id <= 0 {
		return nil, ErrInvalidArgument
	}
	return s.articles.FindPublishedByID(ctx, id)
}

func (s *Service) ListPublished(ctx context.Context, query ListQuery) (*ListResult, error) {
	if s == nil || s.articles == nil {
		return nil, fmt.Errorf("article service is not ready")
	}
	pagination := NewPagination(query.Page, query.PageSize)
	items, total, err := s.articles.ListPublished(ctx, pagination)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total, Page: pagination.Page, Size: pagination.PageSize}, nil
}
