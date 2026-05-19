package plugin

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestOutboundGuardTransportValidatesLiteralResolvedIP(t *testing.T) {
	wantErr := errors.New("blocked")
	transport := NewOutboundGuardTransport(nil, func(pluginKey string, ip net.IP) error {
		if pluginKey != "demo-plugin" {
			t.Fatalf("pluginKey = %q, want demo-plugin", pluginKey)
		}
		if !ip.Equal(net.ParseIP("127.0.0.1")) {
			t.Fatalf("ip = %s, want 127.0.0.1", ip.String())
		}
		return wantErr
	})

	err := transport.validateResolvedAddress(context.Background(), "demo-plugin", "127.0.0.1:9090")
	if !errors.Is(err, wantErr) {
		t.Fatalf("validateResolvedAddress() error = %v, want %v", err, wantErr)
	}
}
