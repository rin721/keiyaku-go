package plugin

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestHeartbeatRunnerRejectsNilClient(t *testing.T) {
	err := (HeartbeatRunner{PluginKey: "demo-plugin", InstanceID: "demo-plugin-1"}).Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("Run() error = %v, want client error", err)
	}
}

func TestHeartbeatRunnerStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := HeartbeatRunner{
		Client: &Client{
			BaseURL:            "http://127.0.0.1:1",
			PluginKey:          "demo-plugin",
			RegistrationSecret: "01234567890123456789012345678901",
		},
		PluginKey:  "demo-plugin",
		InstanceID: "demo-plugin-1",
		Interval:   time.Millisecond,
		OnError:    func(error) {},
	}

	if err := runner.Run(ctx); err == nil {
		t.Fatal("Run() error is nil")
	}
}
