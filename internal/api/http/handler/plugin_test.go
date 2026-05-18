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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("plugin-ok"))
	}))
	defer upstream.Close()

	repo := newHandlerMemoryRepo()
	service, err := appplugin.NewService(repo, appplugin.Config{
		Enabled:            true,
		RegistrationTokens: []string{"01234567890123456789012345678901"},
		AllowedPluginKeys:  []string{"demo-plugin"},
		HeartbeatTTL:       time.Minute,
		RequestTimeout:     time.Second,
		AllowedCIDRs:       []string{"127.0.0.1/32"},
		AllowLoopback:      true,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), appplugin.RegisterCommand{
		Token:         "01234567890123456789012345678901",
		SchemaVersion: "v1",
		PluginKey:     "demo-plugin",
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      "http",
		BaseURL:       upstream.URL,
		HealthPath:    "/healthz",
		Routes: []appplugin.RouteCommand{
			{Method: "GET", MatchType: "exact", Path: "/hello", UpstreamPath: "/hello", AuthPolicy: "authenticated"},
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	engine := gin.New()
	engine.Any("/api/v1/extensions/:plugin_key/*proxy_path", NewPluginHandler(service, fakeTokenIssuer{}, nil).Gateway)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extensions/demo-plugin/hello?q=1", nil)
	req.Header.Set("Authorization", "Bearer ok")

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != "plugin-ok" {
		t.Fatalf("body = %q, want plugin-ok", recorder.Body.String())
	}
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
}

func newHandlerMemoryRepo() *handlerMemoryRepo {
	return &handlerMemoryRepo{
		services:  map[string]*domainplugin.Service{},
		instances: map[string][]*domainplugin.Instance{},
		routes:    map[string][]*domainplugin.Route{},
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

func (r *handlerMemoryRepo) ListPluginServices(ctx context.Context) ([]*domainplugin.Service, error) {
	_ = ctx
	var services []*domainplugin.Service
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *handlerMemoryRepo) GetPluginService(ctx context.Context, pluginKey string) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	_ = ctx
	service := r.services[pluginKey]
	if service == nil {
		return nil, nil, nil, derrors.ErrNotFound
	}
	return service, r.instances[pluginKey], r.routes[pluginKey], nil
}

func (r *handlerMemoryRepo) FindRoutable(ctx context.Context, pluginKey string, now time.Time) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	_ = ctx
	service := r.services[pluginKey]
	if service == nil {
		return nil, nil, nil, derrors.ErrNotFound
	}
	var instances []*domainplugin.Instance
	for _, instance := range r.instances[pluginKey] {
		if instance.Routable(now, service.CurrentManifestHash) {
			instances = append(instances, instance)
		}
	}
	return service, instances, r.routes[pluginKey], nil
}
