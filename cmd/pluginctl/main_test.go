package main

import (
	"strings"
	"testing"
)

func TestScaffoldMainVerifiesGatewaySignature(t *testing.T) {
	content := scaffoldMain("example.com/demo", "demo-plugin", "Demo")
	for _, want := range []string{
		"KEIYAKU_PLUGIN_GATEWAY_SECRET",
		"pluginsdk.VerifySignedRequest",
		"verifyGateway(w, r, gatewaySecret)",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("scaffold main missing %q", want)
		}
	}
}
