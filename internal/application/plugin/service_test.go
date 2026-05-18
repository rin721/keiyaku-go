package plugin

import (
	"context"
	"testing"
	"time"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
)

func TestRegisterRejectsDisallowedPluginKey(t *testing.T) {
	service := newTestService(t)

	_, err := service.Register(context.Background(), validRegisterCommand("other-plugin"))
	if err == nil {
		t.Fatal("Register() error is nil")
	}
}

func TestRegisterStoresManifestRoutes(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)

	result, err := service.Register(context.Background(), validRegisterCommand("demo-plugin"))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.PluginKey != "demo-plugin" || result.ManifestHash == "" {
		t.Fatalf("Register() plugin_key=%q manifest_hash=%q", result.PluginKey, result.ManifestHash)
	}
	if len(repo.routes["demo-plugin"]) != 2 {
		t.Fatalf("stored routes = %d, want 2", len(repo.routes["demo-plugin"]))
	}
}

func TestResolveRoutePrefersExactMethodAndSegmentPrefix(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	resolved, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/items/42"})
	if err != nil {
		t.Fatalf("ResolveRoute() error = %v", err)
	}
	if resolved.Route.Path != "/items" || resolved.Suffix != "/42" {
		t.Fatalf("resolved route path=%q suffix=%q", resolved.Route.Path, resolved.Suffix)
	}

	_, err = service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/items-extra"})
	if err == nil {
		t.Fatal("ResolveRoute() for non-segment prefix error is nil")
	}
}

func TestResolveRouteSkipsExpiredInstances(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	repo.instances["demo-plugin"][0].LeaseExpiresAt = time.Now().UTC().Add(-time.Second)

	_, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/hello"})
	if err == nil {
		t.Fatal("ResolveRoute() error is nil")
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	return newTestServiceWithRepo(t, newMemoryRepo())
}

func newTestServiceWithRepo(t *testing.T, repo *memoryRepo) *Service {
	t.Helper()
	service, err := NewService(repo, Config{
		Enabled:            true,
		RegistrationTokens: []string{"01234567890123456789012345678901"},
		AllowedPluginKeys:  []string{"demo-plugin"},
		HeartbeatTTL:       time.Minute,
		RequestTimeout:     time.Second,
		AllowedHosts:       []string{"plugins.internal"},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func validRegisterCommand(pluginKey string) RegisterCommand {
	return RegisterCommand{
		Token:         "01234567890123456789012345678901",
		SchemaVersion: "v1",
		PluginKey:     pluginKey,
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      "http",
		BaseURL:       "http://plugins.internal:9090",
		HealthPath:    "/healthz",
		Routes: []RouteCommand{
			{Method: "GET", MatchType: "exact", Path: "/hello", UpstreamPath: "/hello", AuthPolicy: "authenticated"},
			{Method: "GET", MatchType: "prefix", Path: "/items", UpstreamPath: "/api/items", AuthPolicy: "authenticated"},
		},
	}
}

type memoryRepo struct {
	services  map[string]*domainplugin.Service
	instances map[string][]*domainplugin.Instance
	routes    map[string][]*domainplugin.Route
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		services:  map[string]*domainplugin.Service{},
		instances: map[string][]*domainplugin.Instance{},
		routes:    map[string][]*domainplugin.Route{},
	}
}

func (r *memoryRepo) UpsertRegistration(ctx context.Context, registration domainplugin.Registration) error {
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

func (r *memoryRepo) TouchInstance(ctx context.Context, pluginKey string, instanceID string, leaseExpiresAt time.Time, now time.Time) (*domainplugin.Instance, error) {
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

func (r *memoryRepo) DisableInstance(ctx context.Context, pluginKey string, instanceID string, now time.Time) error {
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

func (r *memoryRepo) ListPluginServices(ctx context.Context) ([]*domainplugin.Service, error) {
	_ = ctx
	var services []*domainplugin.Service
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *memoryRepo) GetPluginService(ctx context.Context, pluginKey string) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	_ = ctx
	service := r.services[pluginKey]
	if service == nil {
		return nil, nil, nil, derrors.ErrNotFound
	}
	return service, r.instances[pluginKey], r.routes[pluginKey], nil
}

func (r *memoryRepo) FindRoutable(ctx context.Context, pluginKey string, now time.Time) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
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
