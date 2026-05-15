package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainarticle "github.com/rin721/keiyaku-go/internal/domain/article"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	"github.com/rin721/keiyaku-go/internal/domain/shared"
	"gorm.io/gorm"
)

type ArticleModel struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	AuthorID    int64      `gorm:"column:author_id"`
	CategoryID  int64      `gorm:"column:category_id"`
	Title       string     `gorm:"column:title"`
	Slug        string     `gorm:"column:slug"`
	Summary     string     `gorm:"column:summary"`
	Content     string     `gorm:"column:content"`
	Status      string     `gorm:"column:status"`
	TagsJSON    string     `gorm:"column:tags_json"`
	PublishedAt *time.Time `gorm:"column:published_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (ArticleModel) TableName() string {
	return "articles"
}

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, entity *domainarticle.Article) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("article repository is not ready")
	}
	model, err := articleToModel(entity)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isDuplicate(err) {
			return derrors.ErrConflict
		}
		return fmt.Errorf("create article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) FindPublishedByID(ctx context.Context, id int64) (*domainarticle.Article, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("article repository is not ready")
	}
	var model ArticleModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, string(domainarticle.StatusPublished)).
		First(&model).Error; err != nil {
		if IsNotFound(err) {
			return nil, derrors.ErrNotFound
		}
		return nil, fmt.Errorf("find article by id: %w", err)
	}
	return articleFromModel(&model)
}

func (r *ArticleRepository) ListPublished(ctx context.Context, pagination shared.Pagination) ([]*domainarticle.Article, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("article repository is not ready")
	}
	query := r.db.WithContext(ctx).Model(&ArticleModel{}).Where("status = ?", string(domainarticle.StatusPublished))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count published articles: %w", err)
	}
	var models []ArticleModel
	if err := query.Order("published_at DESC, id DESC").
		Offset(pagination.Offset()).
		Limit(pagination.PageSize).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list published articles: %w", err)
	}
	items := make([]*domainarticle.Article, 0, len(models))
	for i := range models {
		item, err := articleFromModel(&models[i])
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func articleToModel(entity *domainarticle.Article) (*ArticleModel, error) {
	if entity == nil {
		return nil, derrors.ErrInvalidArgument
	}
	tags, err := json.Marshal(entity.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal article tags: %w", err)
	}
	return &ArticleModel{
		ID:          entity.ID,
		AuthorID:    entity.AuthorID,
		CategoryID:  entity.CategoryID,
		Title:       entity.Title,
		Slug:        entity.Slug,
		Summary:     entity.Summary,
		Content:     entity.Content,
		Status:      string(entity.Status),
		TagsJSON:    string(tags),
		PublishedAt: entity.PublishedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

func articleFromModel(model *ArticleModel) (*domainarticle.Article, error) {
	if model == nil {
		return nil, derrors.ErrNotFound
	}
	var tags []string
	if model.TagsJSON != "" {
		if err := json.Unmarshal([]byte(model.TagsJSON), &tags); err != nil {
			return nil, fmt.Errorf("unmarshal article tags: %w", err)
		}
	}
	return &domainarticle.Article{
		ID:          model.ID,
		AuthorID:    model.AuthorID,
		CategoryID:  model.CategoryID,
		Title:       model.Title,
		Slug:        model.Slug,
		Summary:     model.Summary,
		Content:     model.Content,
		Status:      domainarticle.Status(model.Status),
		Tags:        tags,
		PublishedAt: model.PublishedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}
