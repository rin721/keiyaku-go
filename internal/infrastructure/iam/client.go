package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/port"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
)

type Client struct {
	baseURL      string
	serviceToken string
	httpClient   *http.Client
	timeout      time.Duration
}

func NewClient(cfg config.IAMConfig) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		serviceToken: cfg.ServiceToken,
		httpClient:   &http.Client{Timeout: timeout},
		timeout:      timeout,
	}
}

func (c *Client) IssueToken(context.Context, port.TokenUser) (port.TokenPair, error) {
	return port.TokenPair{}, fmt.Errorf("iam client cannot issue tokens")
}

func (c *Client) ParseAccessToken(ctx context.Context, token string) (port.TokenClaims, error) {
	var out struct {
		UserID    int64     `json:"user_id"`
		Username  string    `json:"username"`
		Roles     []string  `json:"roles"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/internal/v1/tokens/introspect", map[string]string{"access_token": token}, &out); err != nil {
		return port.TokenClaims{}, err
	}
	return port.TokenClaims{UserID: out.UserID, Username: out.Username, Roles: out.Roles, ExpiresAt: out.ExpiresAt}, nil
}

func (c *Client) Allow(role string, object string, action string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	var out struct {
		Allowed bool `json:"allowed"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/internal/v1/authorize", map[string]string{
		"role":   role,
		"object": object,
		"action": action,
	}, &out)
	return out.Allowed, err
}

func (c *Client) Health(ctx context.Context) error {
	if c == nil || c.baseURL == "" {
		return fmt.Errorf("iam client is not ready")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("iam health status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, body interface{}, out interface{}) error {
	if c == nil || c.baseURL == "" {
		return fmt.Errorf("iam client is not ready")
	}
	content, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode iam request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("build iam request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.serviceToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.serviceToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call iam: %w", err)
	}
	defer resp.Body.Close()
	content, err = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read iam response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("iam status %d", resp.StatusCode)
	}
	var envelope struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(content, &envelope); err == nil && len(envelope.Data) > 0 {
		if envelope.Code != 0 {
			return fmt.Errorf("iam error: %s", envelope.Msg)
		}
		return json.Unmarshal(envelope.Data, out)
	}
	return json.Unmarshal(content, out)
}
