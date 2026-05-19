package plugin

import (
	"context"
	"fmt"
	"time"
)

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
