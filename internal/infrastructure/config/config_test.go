package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResolvesI18NAndRBACPathsRelativeToConfigFile(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "i18n", "en-US.yaml"), `"ok": "ok"`)
	mustWriteFile(t, filepath.Join(dir, "i18n", "zh-CN.yaml"), `"ok": "成功"`)
	mustWriteFile(t, filepath.Join(dir, "rbac", "model.conf"), "[request_definition]\nr = sub, obj, act\n")
	mustWriteFile(t, filepath.Join(dir, "rbac", "policy.csv"), "")
	configPath := filepath.Join(dir, "config.yaml")
	mustWriteFile(t, configPath, `
app:
  name: keiyaku-go
server:
  addr: ":8080"
mysql:
  dsn: "keiyaku:keiyaku@tcp(127.0.0.1:3306)/keiyaku?charset=utf8mb4&parseTime=True&loc=UTC"
redis:
  addr: "127.0.0.1:6379"
jwt:
  secret: "change-me-with-env-KEIYAKU_JWT_SECRET-32-bytes-min"
i18n:
  default: en-US
  supported:
    - en-US
    - zh-CN
  files:
    en-US: i18n/en-US.yaml
    zh-CN: i18n/zh-CN.yaml
rbac:
  model_path: rbac/model.conf
  policy_path: rbac/policy.csv
`)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.I18N.Files["en-US"], filepath.Join(dir, "i18n", "en-US.yaml"); got != want {
		t.Fatalf("en-US file = %q, want %q", got, want)
	}
	if got, want := cfg.I18N.Files["zh-CN"], filepath.Join(dir, "i18n", "zh-CN.yaml"); got != want {
		t.Fatalf("zh-CN file = %q, want %q", got, want)
	}
	if got, want := cfg.RBAC.ModelPath, filepath.Join(dir, "rbac", "model.conf"); got != want {
		t.Fatalf("rbac model path = %q, want %q", got, want)
	}
	if got, want := cfg.RBAC.PolicyPath, filepath.Join(dir, "rbac", "policy.csv"); got != want {
		t.Fatalf("rbac policy path = %q, want %q", got, want)
	}
}

func TestI18NConfigValidateRequiresDefaultSupportedFile(t *testing.T) {
	err := (I18NConfig{
		Default:   "en-US",
		Supported: []string{"en-US", "zh-CN"},
		Files: map[string]string{
			"en-US": "i18n/en-US.yaml",
		},
	}).Validate()
	if err == nil {
		t.Fatal("I18NConfig.Validate() error is nil")
	}
}

func TestRBACConfigValidateRequiresPaths(t *testing.T) {
	if err := (RBACConfig{ModelPath: "rbac/model.conf"}).Validate(); err == nil {
		t.Fatal("RBACConfig.Validate() error is nil")
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
