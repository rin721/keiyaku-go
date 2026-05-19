package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL            string
	PluginKey          string
	RegistrationSecret string
	HTTPClient         *http.Client
	Timeout            time.Duration
}

func NewClient(baseURL string, pluginKey string, registrationSecret string) *Client {
	return &Client{
		BaseURL:            strings.TrimRight(baseURL, "/"),
		PluginKey:          strings.TrimSpace(pluginKey),
		RegistrationSecret: registrationSecret,
		Timeout:            5 * time.Second,
	}
}

func (c *Client) Register(ctx context.Context, manifest Manifest) (RegisterResponse, error) {
	manifest = NormalizeManifest(manifest)
	if err := ValidateManifest(manifest); err != nil {
		return RegisterResponse{}, err
	}
	if err := c.validateIdentity(manifest.PluginKey); err != nil {
		return RegisterResponse{}, err
	}
	var response RegisterResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/plugins/registrations", manifest, &response); err != nil {
		return RegisterResponse{}, err
	}
	return response, nil
}

func (c *Client) Heartbeat(ctx context.Context, pluginKey string, instanceID string) (HeartbeatResponse, error) {
	if err := c.validateIdentity(pluginKey); err != nil {
		return HeartbeatResponse{}, err
	}
	if !ValidPluginKey(instanceID) {
		return HeartbeatResponse{}, validationError("instance_id must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	var response HeartbeatResponse
	path := fmt.Sprintf("/api/v1/plugins/%s/instances/%s/heartbeat", pluginKey, instanceID)
	if err := c.doJSON(ctx, http.MethodPost, path, nil, &response); err != nil {
		return HeartbeatResponse{}, err
	}
	return response, nil
}

func (c *Client) Unregister(ctx context.Context, pluginKey string, instanceID string) error {
	if err := c.validateIdentity(pluginKey); err != nil {
		return err
	}
	if !ValidPluginKey(instanceID) {
		return validationError("instance_id must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	path := fmt.Sprintf("/api/v1/plugins/%s/instances/%s", pluginKey, instanceID)
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) doJSON(ctx context.Context, method string, path string, body interface{}, out interface{}) error {
	if c == nil {
		return httpError(method, "client is nil", nil)
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		return httpError(method, "base URL is required", nil)
	}
	var bodyBytes []byte
	if body != nil {
		content, err := json.Marshal(body)
		if err != nil {
			return httpError(method, "encode request", err)
		}
		bodyBytes = content
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return httpError(method, "build request", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := SignRequest(req, c.PluginKey, c.RegistrationSecret, bodyBytes, time.Time{}, ""); err != nil {
		return httpError(method, "sign request", err)
	}
	client := c.HTTPClient
	if client == nil {
		timeout := c.Timeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return httpError(method, "send request", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		content, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return httpStatusError(method, resp.StatusCode, fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(content))), ErrUnexpectedReply)
	}
	if out == nil {
		return nil
	}
	if err := decodeRegistryResponse(resp.Body, out); err != nil {
		return httpError(method, "decode response", err)
	}
	return nil
}

func (c *Client) validateIdentity(pluginKey string) error {
	if c == nil {
		return httpError("client", "client is nil", nil)
	}
	if !ValidPluginKey(pluginKey) {
		return validationError("plugin_key must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	if c.PluginKey != pluginKey {
		return validationError("client plugin key does not match request plugin key", ErrInvalidManifest)
	}
	if strings.TrimSpace(c.RegistrationSecret) == "" {
		return validationError("registration secret is required", ErrInvalidManifest)
	}
	return nil
}

func decodeRegistryResponse(reader io.Reader, out interface{}) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(content, &envelope); err == nil && len(envelope.Data) > 0 {
		return json.Unmarshal(envelope.Data, out)
	}
	return json.Unmarshal(content, out)
}
