package plugin

import (
	"strings"
	"testing"
)

func TestDecodeRegistryResponseEnvelope(t *testing.T) {
	var result RegisterResponse
	err := decodeRegistryResponse(strings.NewReader(`{"code":0,"msg":"ok","data":{"plugin_key":"demo-plugin","instance_id":"demo-1","manifest_hash":"abc"}}`), &result)
	if err != nil {
		t.Fatalf("decodeRegistryResponse() error = %v", err)
	}
	if result.PluginKey != "demo-plugin" || result.InstanceID != "demo-1" || result.ManifestHash != "abc" {
		t.Fatalf("decoded response plugin_key=%q instance_id=%q manifest_hash=%q", result.PluginKey, result.InstanceID, result.ManifestHash)
	}
}
