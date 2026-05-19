package casbin

import (
	"path/filepath"
	"testing"

	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
)

func TestAuthorizerLoadsConfigFiles(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	authorizer, err := NewAuthorizer(config.RBACConfig{
		ModelPath:  filepath.Join(root, "configs", "rbac", "model.conf"),
		PolicyPath: filepath.Join(root, "configs", "rbac", "policy.csv"),
	})
	if err != nil {
		t.Fatalf("NewAuthorizer() error = %v", err)
	}

	tests := []struct {
		name   string
		role   string
		object string
		action string
		want   bool
	}{
		{name: "admin can delete", role: "admin", object: "/api/v1/extensions/blog/articles/1", action: "DELETE", want: true},
		{name: "author can create blog article", role: "author", object: "/api/v1/extensions/blog/articles", action: "POST", want: true},
		{name: "author cannot delete blog article", role: "author", object: "/api/v1/extensions/blog/articles/1", action: "DELETE", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authorizer.Allow(tt.role, tt.object, tt.action)
			if err != nil {
				t.Fatalf("Allow() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}
