package article

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerCreateRequiresGatewayUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	NewHandler(NewService(&fakeRepo{}, fakeIDs{next: 100})).RegisterRoutes(engine)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"title":"Hello","slug":"hello","content":"content"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestHandlerListOmitsContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	service := NewService(repo, fakeIDs{next: 100})
	entity, err := service.Create(httptest.NewRequest(http.MethodPost, "/", nil).Context(), CreateCommand{
		AuthorID: 1,
		Title:    "Hello",
		Slug:     "hello",
		Content:  "secret content",
		Publish:  true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	repo.entity = entity
	engine := gin.New()
	NewHandler(service).RegisterRoutes(engine)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/articles", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body struct {
		Data ArticleListResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(body.Data.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(body.Data.Items))
	}
	if body.Data.Items[0].Content != "" {
		t.Fatalf("list content = %q, want empty", body.Data.Items[0].Content)
	}
}
