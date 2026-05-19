package plugin

import (
	"context"
	"errors"
	"net/http"
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

func TestLifecycleRunnerRegistersHeartbeatsAndUnregisters(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &fakeLifecycleClient{cancelOnHeartbeat: cancel}
	runner := LifecycleRunner{
		Client:              client,
		Manifest:            validManifest(),
		HeartbeatInterval:   time.Millisecond,
		RetryInitialBackoff: time.Millisecond,
		RetryMaxBackoff:     time.Millisecond,
		UnregisterTimeout:   time.Second,
	}

	err := runner.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want canceled", err)
	}
	if client.registers != 1 || client.heartbeats != 1 || client.unregisters != 1 {
		t.Fatalf("registers=%d heartbeats=%d unregisters=%d", client.registers, client.heartbeats, client.unregisters)
	}
}

func TestLifecycleRunnerReregistersOnMissingHeartbeat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &fakeLifecycleClient{heartbeatErrs: []error{httpStatusError("heartbeat", http.StatusNotFound, "not found", ErrUnexpectedReply)}, cancelOnRegister: cancel}
	runner := LifecycleRunner{
		Client:              client,
		Manifest:            validManifest(),
		HeartbeatInterval:   time.Millisecond,
		RetryInitialBackoff: time.Millisecond,
		RetryMaxBackoff:     time.Millisecond,
		UnregisterTimeout:   time.Second,
	}

	err := runner.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want canceled", err)
	}
	if client.registers < 2 {
		t.Fatalf("registers=%d, want re-register", client.registers)
	}
}

type fakeLifecycleClient struct {
	registers         int
	heartbeats        int
	unregisters       int
	heartbeatErrs     []error
	cancelOnHeartbeat context.CancelFunc
	cancelOnRegister  context.CancelFunc
}

func (c *fakeLifecycleClient) Register(ctx context.Context, manifest Manifest) (RegisterResponse, error) {
	_ = ctx
	c.registers++
	if c.cancelOnRegister != nil && c.registers > 1 {
		c.cancelOnRegister()
	}
	return RegisterResponse{PluginKey: manifest.PluginKey, InstanceID: manifest.InstanceID, LeaseUntil: time.Now().UTC().Add(time.Minute)}, nil
}

func (c *fakeLifecycleClient) Heartbeat(ctx context.Context, pluginKey string, instanceID string) (HeartbeatResponse, error) {
	_ = ctx
	c.heartbeats++
	if len(c.heartbeatErrs) > 0 {
		err := c.heartbeatErrs[0]
		c.heartbeatErrs = c.heartbeatErrs[1:]
		return HeartbeatResponse{}, err
	}
	if c.cancelOnHeartbeat != nil {
		c.cancelOnHeartbeat()
	}
	return HeartbeatResponse{PluginKey: pluginKey, InstanceID: instanceID, LeaseUntil: time.Now().UTC().Add(time.Minute)}, nil
}

func (c *fakeLifecycleClient) Unregister(ctx context.Context, pluginKey string, instanceID string) error {
	_ = ctx
	_ = pluginKey
	_ = instanceID
	c.unregisters++
	return nil
}
