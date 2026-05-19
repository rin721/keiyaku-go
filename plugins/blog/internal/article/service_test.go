package article

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceCreateWritesFirstRevision(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, fakeIDs{next: 100})
	service.now = func() time.Time { return time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC) }

	entity, err := service.Create(context.Background(), CreateCommand{
		AuthorID: 1,
		Title:    "Hello",
		Slug:     "hello",
		Content:  "content",
		Tags:     []string{"go"},
		Publish:  true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if entity.ID != 100 {
		t.Fatalf("id = %d, want 100", entity.ID)
	}
	if repo.revision.Version != 1 || repo.revision.ArticleID != entity.ID {
		t.Fatalf("revision = %#v", repo.revision)
	}
}

func TestServiceCreateRequiresUser(t *testing.T) {
	service := NewService(&fakeRepo{}, fakeIDs{next: 100})
	_, err := service.Create(context.Background(), CreateCommand{Title: "Hello", Slug: "hello", Content: "content"})
	if !errors.Is(err, ErrMissingUser) {
		t.Fatalf("Create() error = %v, want ErrMissingUser", err)
	}
}

type fakeIDs struct {
	next int64
}

func (f fakeIDs) NewID(context.Context) (int64, error) {
	return f.next, nil
}

type fakeRepo struct {
	entity   *Article
	revision Revision
}

func (r *fakeRepo) Create(_ context.Context, entity *Article, revision Revision) error {
	r.entity = entity
	r.revision = revision
	return nil
}

func (r *fakeRepo) FindPublishedByID(_ context.Context, id int64) (*Article, error) {
	if r.entity == nil || r.entity.ID != id || r.entity.Status != StatusPublished {
		return nil, ErrNotFound
	}
	return r.entity, nil
}

func (r *fakeRepo) ListPublished(context.Context, Pagination) ([]*Article, int64, error) {
	if r.entity == nil || r.entity.Status != StatusPublished {
		return nil, 0, nil
	}
	return []*Article{r.entity}, 1, nil
}
