package plugin

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RegistryClient interface {
	Register(ctx context.Context, manifest Manifest) (RegisterResponse, error)
	Heartbeat(ctx context.Context, pluginKey string, instanceID string) (HeartbeatResponse, error)
	Unregister(ctx context.Context, pluginKey string, instanceID string) error
}

type HeartbeatRunner struct {
	Client     *Client
	PluginKey  string
	InstanceID string
	Interval   time.Duration
	OnError    func(error)
}

func (r HeartbeatRunner) Run(ctx context.Context) error {
	if r.Client == nil {
		return fmt.Errorf("plugin heartbeat client is required")
	}
	if !ValidPluginKey(r.PluginKey) {
		return validationError("plugin_key must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	if !ValidPluginKey(r.InstanceID) {
		return validationError("instance_id must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	interval := r.Interval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	if _, err := r.Client.Heartbeat(ctx, r.PluginKey, r.InstanceID); err != nil {
		if r.OnError != nil {
			r.OnError(err)
		}
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := r.Client.Heartbeat(ctx, r.PluginKey, r.InstanceID); err != nil && r.OnError != nil {
				r.OnError(err)
			}
		}
	}
}

type LifecycleRunner struct {
	Client              RegistryClient
	Manifest            Manifest
	HeartbeatInterval   time.Duration
	RegisterTimeout     time.Duration
	UnregisterTimeout   time.Duration
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
	RetryJitter         time.Duration
	OnRegister          func(RegisterResponse)
	OnHeartbeat         func(HeartbeatResponse)
	OnUnregister        func(error)
	OnError             func(error)
}

func (r LifecycleRunner) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	manifest := NormalizeManifest(r.Manifest)
	if err := ValidateManifest(manifest); err != nil {
		return err
	}
	if r.Client == nil {
		return fmt.Errorf("plugin lifecycle client is required")
	}
	if err := r.register(ctx, manifest); err != nil {
		return err
	}
	defer r.unregister(manifest)

	interval := r.HeartbeatInterval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	backoff := r.initialBackoff()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			result, err := r.Client.Heartbeat(ctx, manifest.PluginKey, manifest.InstanceID)
			if err == nil {
				backoff = r.initialBackoff()
				if r.OnHeartbeat != nil {
					r.OnHeartbeat(result)
				}
				continue
			}
			r.reportError(err)
			if shouldReregister(err) {
				if err := r.sleep(ctx, backoff); err != nil {
					return err
				}
				if err := r.register(ctx, manifest); err != nil {
					r.reportError(err)
				}
				backoff = r.nextBackoff(backoff)
			}
		}
	}
}

func (r LifecycleRunner) register(ctx context.Context, manifest Manifest) error {
	timeout := r.RegisterTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	registerCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := r.Client.Register(registerCtx, manifest)
	if err != nil {
		return err
	}
	if r.OnRegister != nil {
		r.OnRegister(result)
	}
	return nil
}

func (r LifecycleRunner) unregister(manifest Manifest) {
	timeout := r.UnregisterTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := r.Client.Unregister(ctx, manifest.PluginKey, manifest.InstanceID)
	if r.OnUnregister != nil {
		r.OnUnregister(err)
	}
}

func (r LifecycleRunner) sleep(ctx context.Context, duration time.Duration) error {
	duration += jitterDuration(r.RetryJitter)
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r LifecycleRunner) reportError(err error) {
	if err != nil && r.OnError != nil {
		r.OnError(err)
	}
}

func (r LifecycleRunner) initialBackoff() time.Duration {
	if r.RetryInitialBackoff > 0 {
		return r.RetryInitialBackoff
	}
	return time.Second
}

func (r LifecycleRunner) nextBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		current = r.initialBackoff()
	}
	next := current * 2
	maxBackoff := r.RetryMaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}

func jitterDuration(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}

func shouldReregister(err error) bool {
	return IsHTTPStatus(err, http.StatusNotFound) || IsHTTPStatus(err, http.StatusConflict)
}

func IsContextDone(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
