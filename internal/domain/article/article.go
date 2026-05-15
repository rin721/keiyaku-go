package article

import (
	"strings"
	"time"
	"unicode/utf8"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Article struct {
	ID          int64
	AuthorID    int64
	CategoryID  int64
	Title       string
	Slug        string
	Summary     string
	Content     string
	Status      Status
	Tags        []string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(id, authorID, categoryID int64, title, slug, summary, content string, tags []string, now time.Time) (*Article, error) {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	summary = strings.TrimSpace(summary)
	content = strings.TrimSpace(content)
	if id <= 0 || authorID <= 0 || !validTitle(title) || !validSlug(slug) || content == "" {
		return nil, derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &Article{
		ID:         id,
		AuthorID:   authorID,
		CategoryID: categoryID,
		Title:      title,
		Slug:       slug,
		Summary:    summary,
		Content:    content,
		Status:     StatusDraft,
		Tags:       normalizeTags(tags),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (a *Article) Publish(now time.Time) error {
	if a == nil {
		return derrors.ErrNotFound
	}
	if a.Status == StatusArchived {
		return derrors.ErrConflict
	}
	if !validTitle(a.Title) || strings.TrimSpace(a.Content) == "" {
		return derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	a.Status = StatusPublished
	a.PublishedAt = &now
	a.UpdatedAt = now
	return nil
}

func validTitle(title string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(title))
	return n >= 1 && n <= 160
}

func validSlug(slug string) bool {
	n := utf8.RuneCountInString(slug)
	if n < 1 || n > 180 {
		return false
	}
	for _, r := range slug {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' {
			continue
		}
		return false
	}
	return true
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
