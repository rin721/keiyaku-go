package article

import (
	"testing"
	"time"
)

func TestArticlePublish(t *testing.T) {
	now := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)
	entity, err := New(1, 2, 0, "Hello", "hello", "summary", "content", []string{"Go", "go", " "}, now)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := entity.Publish(now.Add(time.Minute)); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if entity.Status != StatusPublished {
		t.Fatalf("status = %s, want published", entity.Status)
	}
	if entity.PublishedAt == nil {
		t.Fatal("PublishedAt is nil")
	}
	if got := len(entity.Tags); got != 1 {
		t.Fatalf("tags length = %d, want 1", got)
	}
}

func TestArticleRejectsInvalidSlug(t *testing.T) {
	_, err := New(1, 2, 0, "Hello", "Hello", "", "content", nil, time.Now())
	if err == nil {
		t.Fatal("New() error is nil")
	}
}
