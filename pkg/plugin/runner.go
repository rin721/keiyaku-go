package plugin

import (
	"context"
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
