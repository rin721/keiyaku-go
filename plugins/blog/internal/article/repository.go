package article

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
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
	return "blog_articles"
}

type RevisionModel struct {
	ArticleID int64     `gorm:"column:article_id;primaryKey"`
	Version   int       `gorm:"column:version;primaryKey"`
	Title     string    `gorm:"column:title"`
	Summary   string    `gorm:"column:summary"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (RevisionModel) TableName() string {
	return "blog_article_revisions"
}

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, entity *Article, revision Revision) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("article repository is not ready")
	}
	articleModel, err := articleToModel(entity)
	if err != nil {
		return err
	}
	revisionModel := RevisionModel{
		ArticleID: revision.ArticleID,
		Version:   revision.Version,
		Title:     revision.Title,
		Summary:   revision.Summary,
		Content:   revision.Content,
		CreatedAt: revision.CreatedAt,
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(articleModel).Error; err != nil {
			return err
		}
		if err := tx.Create(&revisionModel).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if isDuplicate(err) {
			return ErrConflict
		}
		return fmt.Errorf("create article: %w", err)
	}
	return nil
}

func (r *MySQLRepository) FindPublishedByID(ctx context.Context, id int64) (*Article, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("article repository is not ready")
	}
	var model ArticleModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, string(StatusPublished)).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find article by id: %w", err)
	}
	return articleFromModel(&model)
}

func (r *MySQLRepository) ListPublished(ctx context.Context, pagination Pagination) ([]*Article, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("article repository is not ready")
	}
	query := r.db.WithContext(ctx).Model(&ArticleModel{}).Where("status = ?", string(StatusPublished))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count published articles: %w", err)
	}
	var models []ArticleModel
	if err := query.Select("id", "author_id", "category_id", "title", "slug", "summary", "status", "tags_json", "published_at", "created_at", "updated_at").
		Order("published_at DESC, id DESC").
		Offset(pagination.Offset()).
		Limit(pagination.PageSize).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list published articles: %w", err)
	}
	items := make([]*Article, 0, len(models))
	for i := range models {
		item, err := articleFromModel(&models[i])
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func articleToModel(entity *Article) (*ArticleModel, error) {
	if entity == nil {
		return nil, ErrInvalidArgument
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

func articleFromModel(model *ArticleModel) (*Article, error) {
	if model == nil {
		return nil, ErrNotFound
	}
	var tags []string
	if model.TagsJSON != "" {
		if err := json.Unmarshal([]byte(model.TagsJSON), &tags); err != nil {
			return nil, fmt.Errorf("unmarshal article tags: %w", err)
		}
	}
	return &Article{
		ID:          model.ID,
		AuthorID:    model.AuthorID,
		CategoryID:  model.CategoryID,
		Title:       model.Title,
		Slug:        model.Slug,
		Summary:     model.Summary,
		Content:     model.Content,
		Status:      Status(model.Status),
		Tags:        tags,
		PublishedAt: model.PublishedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
