package plugin

import (
	"strings"
	"testing"
)

func TestValidateManifestAcceptsV3HTTPManifest(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: DefaultSchemaVersion,
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      ProtocolHTTP,
		BaseURL:       "http://plugins.internal:9090",
		Routes: []Route{
			{
				RouteID:      "hello",
				Method:       MethodGet,
				MatchType:    MatchTypeExact,
				GatewayPath:  "/api/v1/extensions/demo/hello",
				UpstreamPath: "/hello",
				AuthPolicy:   AuthPolicyAuthenticated,
				Timeout:      "5s",
			},
		},
	}

	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("ValidateManifest() error = %v", err)
	}
}

func TestValidateManifestRejectsV2Manifest(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: "v2",
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      ProtocolHTTP,
		BaseURL:       "http://plugins.internal:9090",
		Routes: []Route{
			{RouteID: "hello", Method: MethodGet, MatchType: MatchTypeExact, GatewayPath: "/api/v1/extensions/demo/hello", UpstreamPath: "/hello", Timeout: "5s"},
		},
	}

	err := ValidateManifest(manifest)
	if err == nil || !strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("ValidateManifest() error = %v, want schema_version failure", err)
	}
}

func TestValidateManifestRejectsMissingRouteID(t *testing.T) {
	manifest := validManifest()
	manifest.Routes[0].RouteID = ""

	err := ValidateManifest(manifest)
	if err == nil || !strings.Contains(err.Error(), "route_id") {
		t.Fatalf("ValidateManifest() error = %v, want route_id failure", err)
	}
}

func TestValidateManifestRejectsUnsafeBaseURL(t *testing.T) {
	manifest := validManifest()
	manifest.BaseURL = "http://user:pass@plugins.internal:9090?token=secret"

	err := ValidateManifest(manifest)
	if err == nil {
		t.Fatal("ValidateManifest() error is nil")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("ValidateManifest() error = %v, want base_url failure", err)
	}
}

func TestValidateGatewayPathRequiresPublicPrefix(t *testing.T) {
	if err := ValidateGatewayPath("/api/v1/extensions/blog/articles", "/api/v1/extensions"); err != nil {
		t.Fatalf("ValidateGatewayPath() error = %v", err)
	}
	err := ValidateGatewayPath("/api/v1/users/me", "/api/v1/extensions")
	if err == nil || !strings.Contains(err.Error(), "public_prefix") {
		t.Fatalf("ValidateGatewayPath() error = %v, want public_prefix failure", err)
	}
}

func TestManifestHashIsStableAcrossRouteOrder(t *testing.T) {
	manifest := validManifest()
	manifest.Routes = append(manifest.Routes, Route{
		RouteID:      "items",
		Method:       MethodPost,
		MatchType:    MatchTypeExact,
		GatewayPath:  "/api/v1/extensions/demo/items",
		UpstreamPath: "/items",
		AuthPolicy:   AuthPolicyAuthenticated,
		Timeout:      "5s",
	})
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

func validManifest() Manifest {
	return Manifest{
		SchemaVersion: DefaultSchemaVersion,
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      ProtocolHTTP,
		BaseURL:       "http://plugins.internal:9090",
		Routes: []Route{
			{
				RouteID:      "hello",
				Method:       MethodGet,
				MatchType:    MatchTypeExact,
				GatewayPath:  "/api/v1/extensions/demo/hello",
				UpstreamPath: "/hello",
				AuthPolicy:   AuthPolicyAuthenticated,
				Timeout:      "5s",
			},
		},
	}
}
