package article

import (
	"testing"
	"time"
)

func TestArticlePublish(t *testing.T) {
	now := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	got, err := New(1, 2, 0, "Hello", "hello", "", "content", []string{"Go", "go", " CMS "}, now)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("Tags = %#v", got.Tags)
	}
	if err := got.Publish(now.Add(time.Hour)); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if got.Status != StatusPublished {
		t.Fatalf("Status = %q", got.Status)
	}
	if got.PublishedAt == nil {
		t.Fatal("PublishedAt is nil")
	}
}

func TestArticleRejectsInvalidSlug(t *testing.T) {
	if _, err := New(1, 2, 0, "Hello", "Hello World", "", "content", nil, time.Now()); err == nil {
		t.Fatal("expected invalid slug error")
	}
}
