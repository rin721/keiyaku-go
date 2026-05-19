package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	appplugin "github.com/rin721/keiyaku-go/internal/application/plugin"
	"github.com/rin721/keiyaku-go/internal/application/port"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
)

func TestPluginGatewayForwardsRequestWithoutAuthorizationByDefault(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/hello" {
			t.Fatalf("upstream path = %q, want /hello", got)
		}
		if got := r.URL.RawQuery; got != "q=1" {
			t.Fatalf("upstream query = %q, want q=1", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization forwarded = %q, want empty", got)
		}
		if got := r.Header.Get("X-Keiyaku-User-ID"); got != "42" {
			t.Fatalf("X-Keiyaku-User-ID = %q, want 42", got)
		}
		if got := r.Header.Get("X-Keiyaku-Plugin-Key"); got != "demo-plugin" {
			t.Fatalf("X-Keiyaku-Plugin-Key = %q, want demo-plugin", got)
		}
		if got := r.Header.Get("X-Forwarded-Host"); got == "evil.example" {
			t.Fatalf("X-Forwarded-Host was not rebuilt")
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("plugin-ok"))
	}))
	defer upstream.Close()

	repo := newHandlerMemoryRepo()
	service, err := appplugin.NewService(repo, appplugin.Config{
		Enabled: true,
		TrustedPlugins: map[string]appplugin.TrustedPluginConfig{
			"demo-plugin": {
				RegistrationSecret: "01234567890123456789012345678901",
				GatewaySecret:      "abcdefghijklmnopqrstuvwxyz123456",
				AllowedCIDRs:       []string{"127.0.0.1/32"},
				AllowLoopback:      true,
			},
		},
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), appplugin.RegisterCommand{
		Signature:     testHandlerSignature("demo-plugin", "POST", "/api/v1/plugins/registrations", nil),
		SchemaVersion: "v2",
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      "http",
		BaseURL:       upstream.URL,
		HealthPath:    "/healthz",
		Routes: []appplugin.RouteCommand{
			{RouteID: "hello", Method: "GET", MatchType: "exact", GatewayPath: "/api/v1/extensions/demo/hello", UpstreamPath: "/hello", AuthPolicy: "authenticated", Timeout: "1s"},
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	engine := gin.New()
	engine.Any("/api/v1/extensions/*proxy_path", NewPluginHandler(service, fakeTokenIssuer{}, nil).Gateway)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extensions/demo/hello?q=1", nil)
	req.Header.Set("Authorization", "Bearer ok")
	req.Header.Set("X-Keiyaku-User-ID", "999")
	req.Header.Set("X-Keiyaku-Plugin-Key", "spoof")
	req.Header.Set("X-Forwarded-Host", "evil.example")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != "plugin-ok" {
		t.Fatalf("body = %q, want plugin-ok", recorder.Body.String())
	}
}

func TestPluginGatewayForwardsAuthorizationWhenRouteAllows(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ok" {
			t.Fatalf("Authorization forwarded = %q, want Bearer ok", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	service := newRegisteredGatewayService(t, upstream.URL, appplugin.RouteCommand{
		RouteID:           "hello",
		Method:            "GET",
		MatchType:         "exact",
		GatewayPath:       "/api/v1/extensions/demo/hello",
		UpstreamPath:      "/hello",
		AuthPolicy:        "authenticated",
		Timeout:           "1s",
		ForwardAuthHeader: true,
	}, nil)
	engine := gin.New()
	engine.Any("/api/v1/extensions/*proxy_path", NewPluginHandler(service, fakeTokenIssuer{}, nil).Gateway)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extensions/demo/hello", nil)
	req.Header.Set("Authorization", "Bearer ok")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestPluginGatewayRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	upstreamCalled := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	service := newRegisteredGatewayService(t, upstream.URL, appplugin.RouteCommand{
		RouteID:      "submit",
		Method:       "POST",
		MatchType:    "exact",
		GatewayPath:  "/api/v1/extensions/demo/submit",
		UpstreamPath: "/submit",
		AuthPolicy:   "authenticated",
		Timeout:      "1s",
	}, func(config *appplugin.Config) {
		config.MaxGatewayBodyBytes = 4
	})
	engine := gin.New()
	engine.Any("/api/v1/extensions/*proxy_path", NewPluginHandler(service, fakeTokenIssuer{}, nil).Gateway)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/demo/submit", strings.NewReader("too-large"))
	req.Header.Set("Authorization", "Bearer ok")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusRequestEntityTooLarge, recorder.Body.String())
	}
	if upstreamCalled {
		t.Fatal("upstream was called for oversized request")
	}
}

func testHandlerSignature(pluginKey string, method string, path string, body []byte) appplugin.SignatureCommand {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	nonce := pluginKey + "-nonce-" + method + "-" + path
	return appplugin.SignatureCommand{
		PluginKey: pluginKey,
		Method:    method,
		Path:      path,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: pkgplugin.Sign(method, path, timestamp, nonce, pkgplugin.BodySHA256(body), "01234567890123456789012345678901"),
		Body:      body,
	}
}

func newRegisteredGatewayService(t *testing.T, upstreamURL string, route appplugin.RouteCommand, mutate func(*appplugin.Config)) *appplugin.Service {
	t.Helper()
	repo := newHandlerMemoryRepo()
	config := appplugin.Config{
		Enabled: true,
		TrustedPlugins: map[string]appplugin.TrustedPluginConfig{
			"demo-plugin": {
				RegistrationSecret: "01234567890123456789012345678901",
				GatewaySecret:      "abcdefghijklmnopqrstuvwxyz123456",
				AllowedCIDRs:       []string{"127.0.0.1/32"},
				AllowLoopback:      true,
			},
		},
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
	}
	if mutate != nil {
		mutate(&config)
	}
	service, err := appplugin.NewService(repo, config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), appplugin.RegisterCommand{
		Signature:     testHandlerSignature("demo-plugin", "POST", "/api/v1/plugins/registrations", nil),
		SchemaVersion: "v2",
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      "http",
		BaseURL:       upstreamURL,
		HealthPath:    "/healthz",
		Routes:        []appplugin.RouteCommand{route},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return service
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) IssueToken(ctx context.Context, subject port.TokenUser) (port.TokenPair, error) {
	_ = ctx
	_ = subject
	return port.TokenPair{}, nil
}

func (fakeTokenIssuer) ParseAccessToken(ctx context.Context, token string) (port.TokenClaims, error) {
	_ = ctx
	if token != "ok" {
		return port.TokenClaims{}, derrors.ErrUnauthorized
	}
	return port.TokenClaims{UserID: 42, Username: "alice", Roles: []string{"author"}}, nil
}

type handlerMemoryRepo struct {
	services  map[string]*domainplugin.Service
	instances map[string][]*domainplugin.Instance
	routes    map[string][]*domainplugin.Route
	nonces    map[string]time.Time
}

func newHandlerMemoryRepo() *handlerMemoryRepo {
	return &handlerMemoryRepo{
		services:  map[string]*domainplugin.Service{},
		instances: map[string][]*domainplugin.Instance{},
		routes:    map[string][]*domainplugin.Route{},
		nonces:    map[string]time.Time{},
	}
}

func (r *handlerMemoryRepo) UpsertRegistration(ctx context.Context, registration domainplugin.Registration) error {
	_ = ctx
	service := registration.Service
	r.services[service.PluginKey] = &service
	instance := registration.Instance
	r.instances[service.PluginKey] = []*domainplugin.Instance{&instance}
	r.routes[service.PluginKey] = nil
	for i := range registration.Routes {
		route := registration.Routes[i]
		r.routes[service.PluginKey] = append(r.routes[service.PluginKey], &route)
	}
	return nil
}

func (r *handlerMemoryRepo) FindRouteConflict(ctx context.Context, pluginKey string, routes []domainplugin.Route) (*domainplugin.RouteConflict, error) {
	_ = ctx
	for _, route := range routes {
		for otherPluginKey, service := range r.services {
			if otherPluginKey == pluginKey || service == nil || service.Status != domainplugin.ServiceStatusActive {
				continue
			}
			for _, existing := range r.routes[otherPluginKey] {
				if existing == nil || !existing.Enabled || existing.ManifestHash != service.CurrentManifestHash {
					continue
				}
				if existing.Method == route.Method && existing.MatchType == route.MatchType && existing.GatewayPath == route.GatewayPath {
					return &domainplugin.RouteConflict{
						PluginKey:   existing.PluginKey,
						RouteID:     existing.RouteID,
						Method:      existing.Method,
						MatchType:   existing.MatchType,
						GatewayPath: existing.GatewayPath,
					}, nil
				}
			}
		}
	}
	return nil, nil
}

func (r *handlerMemoryRepo) TouchInstance(ctx context.Context, pluginKey string, instanceID string, leaseExpiresAt time.Time, now time.Time) (*domainplugin.Instance, error) {
	_ = ctx
	for _, instance := range r.instances[pluginKey] {
		if instance.InstanceID == instanceID {
			instance.LeaseExpiresAt = leaseExpiresAt
			instance.LastSeenAt = now
			return instance, nil
		}
	}
	return nil, derrors.ErrNotFound
}

func (r *handlerMemoryRepo) DisableInstance(ctx context.Context, pluginKey string, instanceID string, now time.Time) error {
	_ = ctx
	_ = now
	for _, instance := range r.instances[pluginKey] {
		if instance.InstanceID == instanceID {
			instance.Status = domainplugin.InstanceStatusDisabled
			return nil
		}
	}
	return derrors.ErrNotFound
}

func (r *handlerMemoryRepo) SetServiceStatus(ctx context.Context, pluginKey string, status domainplugin.ServiceStatus, now time.Time) error {
	_ = ctx
	service := r.services[pluginKey]
	if service == nil {
		return derrors.ErrNotFound
	}
	service.Status = status
	service.UpdatedAt = now
	if status == domainplugin.ServiceStatusDisabled {
		service.DisabledAt = &now
	} else {
		service.DisabledAt = nil
	}
	return nil
}

func (r *handlerMemoryRepo) SetInstanceStatus(ctx context.Context, pluginKey string, instanceID string, status domainplugin.InstanceStatus, now time.Time) error {
	_ = ctx
	for _, instance := range r.instances[pluginKey] {
		if instance.InstanceID == instanceID {
			instance.Status = status
			instance.UpdatedAt = now
			return nil
		}
	}
	return derrors.ErrNotFound
}

func (r *handlerMemoryRepo) ListPluginServices(ctx context.Context) ([]*domainplugin.Service, error) {
	_ = ctx
	var services []*domainplugin.Service
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *handlerMemoryRepo) ListPluginInstances(ctx context.Context, pluginKey string) ([]*domainplugin.Instance, error) {
	_ = ctx
	return r.instances[pluginKey], nil
}

func (r *handlerMemoryRepo) ListHealthCheckTargets(ctx context.Context, now time.Time) ([]*domainplugin.Instance, error) {
	_ = ctx
	_ = now
	var instances []*domainplugin.Instance
	for pluginKey, service := range r.services {
		if service.Status != domainplugin.ServiceStatusActive {
			continue
		}
		for _, instance := range r.instances[pluginKey] {
			if instance.Status == domainplugin.InstanceStatusActive && instance.ManifestHash == service.CurrentManifestHash && !instance.LeaseExpiresAt.Before(now) {
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
}

func (r *handlerMemoryRepo) GetPluginService(ctx context.Context, pluginKey string) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	_ = ctx
	service := r.services[pluginKey]
	if service == nil {
		return nil, nil, nil, derrors.ErrNotFound
	}
	return service, r.instances[pluginKey], r.routes[pluginKey], nil
}

func (r *handlerMemoryRepo) FindRoutable(ctx context.Context, now time.Time) ([]*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	_ = ctx
	var services []*domainplugin.Service
	var instances []*domainplugin.Instance
	var routes []*domainplugin.Route
	for pluginKey, service := range r.services {
		if service == nil || service.Status != domainplugin.ServiceStatusActive {
			continue
		}
		services = append(services, service)
		for _, instance := range r.instances[pluginKey] {
			if instance.Routable(now, service.CurrentManifestHash) {
				instances = append(instances, instance)
			}
		}
		routes = append(routes, r.routes[pluginKey]...)
	}
	return services, instances, routes, nil
}

func (r *handlerMemoryRepo) UpdateInstanceHealth(ctx context.Context, pluginKey string, instanceID string, healthStatus domainplugin.HealthStatus, consecutiveFailures int, lastError string, checkedAt time.Time) (*domainplugin.Instance, error) {
	_ = ctx
	for _, instance := range r.instances[pluginKey] {
		if instance.InstanceID == instanceID {
			instance.HealthStatus = healthStatus.Normalize()
			instance.ConsecutiveFailures = consecutiveFailures
			instance.LastError = lastError
			instance.LastCheckedAt = &checkedAt
			if lastError == "" {
				instance.LastErrorAt = nil
			} else {
				instance.LastErrorAt = &checkedAt
			}
			return instance, nil
		}
	}
	return nil, derrors.ErrNotFound
}

func (r *handlerMemoryRepo) UseSignatureNonce(ctx context.Context, pluginKey string, nonce string, expiresAt time.Time, now time.Time) error {
	_ = ctx
	_ = expiresAt
	_ = now
	key := pluginKey + ":" + nonce
	if _, ok := r.nonces[key]; ok {
		return derrors.ErrConflict
	}
	r.nonces[key] = expiresAt
	return nil
}

func (r *handlerMemoryRepo) PruneSignatureNonces(ctx context.Context, now time.Time) (int64, error) {
	_ = ctx
	var count int64
	for key, expiresAt := range r.nonces {
		if expiresAt.Before(now) {
			delete(r.nonces, key)
			count++
		}
	}
	return count, nil
}

func (r *handlerMemoryRepo) PrunePluginAuditEvents(ctx context.Context, before time.Time) (int64, error) {
	_ = ctx
	_ = before
	return 0, nil
}

func (r *handlerMemoryRepo) DisableStalePluginInstances(ctx context.Context, staleBefore time.Time, now time.Time) (int64, error) {
	_ = ctx
	var count int64
	for _, instances := range r.instances {
		for _, instance := range instances {
			if instance.Status == domainplugin.InstanceStatusActive && instance.LeaseExpiresAt.Before(staleBefore) {
				instance.Status = domainplugin.InstanceStatusDisabled
				instance.UpdatedAt = now
				count++
			}
		}
	}
	return count, nil
}
