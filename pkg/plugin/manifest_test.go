package plugin

import (
	"strings"
	"testing"
)

func TestValidateManifestAcceptsHTTPManifest(t *testing.T) {
	manifest := Manifest{
		PluginKey:  "demo-plugin",
		Name:       "Demo",
		Version:    "0.1.0",
		InstanceID: "demo-plugin-1",
		Protocol:   ProtocolHTTP,
		BaseURL:    "http://plugins.internal:9090",
		Routes: []Route{
			{
				Method:       MethodGet,
				MatchType:    MatchTypeExact,
				Path:         "/hello",
				UpstreamPath: "/hello",
				AuthPolicy:   AuthPolicyAuthenticated,
			},
		},
	}

	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("ValidateManifest() error = %v", err)
	}
}

func TestValidateManifestRejectsUnsafeBaseURL(t *testing.T) {
	manifest := Manifest{
		PluginKey:  "demo-plugin",
		Name:       "Demo",
		Version:    "0.1.0",
		InstanceID: "demo-plugin-1",
		Protocol:   ProtocolHTTP,
		BaseURL:    "http://user:pass@plugins.internal:9090?token=secret",
		Routes: []Route{
			{Method: MethodGet, MatchType: MatchTypeExact, Path: "/hello", UpstreamPath: "/hello"},
		},
	}

	err := ValidateManifest(manifest)
	if err == nil {
		t.Fatal("ValidateManifest() error is nil")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("ValidateManifest() error = %v, want base_url failure", err)
	}
}

func TestManifestHashIsStableAcrossRouteOrder(t *testing.T) {
	manifest := Manifest{
		PluginKey:  "demo-plugin",
		Name:       "Demo",
		Version:    "0.1.0",
		InstanceID: "demo-plugin-1",
		Protocol:   ProtocolHTTP,
		BaseURL:    "http://plugins.internal:9090",
		Routes: []Route{
			{Method: MethodPost, MatchType: MatchTypeExact, Path: "/items", UpstreamPath: "/items"},
			{Method: MethodGet, MatchType: MatchTypePrefix, Path: "/items", UpstreamPath: "/items"},
		},
	}
	other := manifest
	other.Routes = []Route{manifest.Routes[1], manifest.Routes[0]}

	left, err := ManifestHash(manifest)
	if err != nil {
		t.Fatalf("ManifestHash() error = %v", err)
	}
	right, err := ManifestHash(other)
	if err != nil {
		t.Fatalf("ManifestHash() error = %v", err)
	}
	if left != right {
		t.Fatalf("hash mismatch: %s != %s", left, right)
	}
}
