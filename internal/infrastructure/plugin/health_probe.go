package plugin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	"github.com/rin721/keiyaku-go/internal/observability/trace"
)

type HTTPHealthProbe struct {
	client  *http.Client
	timeout time.Duration
}

func NewHTTPHealthProbe(timeout time.Duration) *HTTPHealthProbe {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &HTTPHealthProbe{
		client:  &http.Client{},
		timeout: timeout,
	}
}

func (p *HTTPHealthProbe) Probe(ctx context.Context, instance domainplugin.Instance) error {
	if p == nil {
		return fmt.Errorf("plugin health probe is not ready")
	}
	timeout := p.timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	healthPath := instance.HealthPath
	if healthPath == "" {
		healthPath = "/healthz"
	}
	target, err := domainplugin.BuildUpstreamURL(instance.BaseURL, healthPath, "", "")
	if err != nil {
		return fmt.Errorf("build plugin health URL: %w", err)
	}
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("create plugin health request: %w", err)
	}
	if traceID := trace.IDFromContext(ctx); traceID != "" {
		req.Header.Set(trace.HeaderName, traceID)
	}
	client := p.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("call plugin health endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("plugin health endpoint returned %d", resp.StatusCode)
	}
	return nil
}
