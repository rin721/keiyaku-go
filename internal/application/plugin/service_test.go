package plugin

import (
	"context"
	"errors"
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

func TestResolveRouteSkipsUnhealthyInstances(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	repo.instances["demo-plugin"][0].HealthStatus = domainplugin.HealthStatusUnhealthy

	_, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/hello"})
	if err == nil {
		t.Fatal("ResolveRoute() error is nil")
	}
}

func TestHealthCheckMarksUnhealthyAfterThreshold(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:             true,
		RegistrationTokens:  []string{"01234567890123456789012345678901"},
		AllowedPluginKeys:   []string{"demo-plugin"},
		HeartbeatTTL:        time.Minute,
		RequestTimeout:      time.Second,
		HealthCheckInterval: time.Second,
		UnhealthyThreshold:  2,
		MaxAuditQueryLimit:  20,
		AllowedHosts:        []string{"plugins.internal"},
	}, WithAuditRepository(repo))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	probe := fakeHealthProbe{err: errors.New("probe failed")}
	if err := service.CheckHealth(context.Background(), probe); err != nil {
		t.Fatalf("CheckHealth() first error = %v", err)
	}
	if got := repo.instances["demo-plugin"][0].HealthStatus; got != domainplugin.HealthStatusUnknown {
		t.Fatalf("health after first failure = %q, want unknown", got)
	}
	if err := service.CheckHealth(context.Background(), probe); err != nil {
		t.Fatalf("CheckHealth() second error = %v", err)
	}
	if got := repo.instances["demo-plugin"][0].HealthStatus; got != domainplugin.HealthStatusUnhealthy {
		t.Fatalf("health after second failure = %q, want unhealthy", got)
	}
}

func TestRouteCacheInvalidatesOnDisableService(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:            true,
		RegistrationTokens: []string{"01234567890123456789012345678901"},
		AllowedPluginKeys:  []string{"demo-plugin"},
		HeartbeatTTL:       time.Minute,
		RequestTimeout:     time.Second,
		RouteCacheTTL:      time.Minute,
		AllowedHosts:       []string{"plugins.internal"},
	}, WithAuditRepository(repo))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/hello"}); err != nil {
		t.Fatalf("ResolveRoute() before disable error = %v", err)
	}
	if err := service.DisableService(context.Background(), "demo-plugin"); err != nil {
		t.Fatalf("DisableService() error = %v", err)
	}
	if _, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{PluginKey: "demo-plugin", Method: "GET", Path: "/hello"}); err == nil {
		t.Fatal("ResolveRoute() after disable error is nil")
	}
}

func TestRegisterRecordsAuditEvents(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:            true,
		RegistrationTokens: []string{"01234567890123456789012345678901"},
		AllowedPluginKeys:  []string{"demo-plugin"},
		HeartbeatTTL:       time.Minute,
		RequestTimeout:     time.Second,
		AllowedHosts:       []string{"plugins.internal"},
	}, WithAuditRepository(repo))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	events, err := service.ListAuditEvents(context.Background(), "demo-plugin", 10)
	if err != nil {
		t.Fatalf("ListAuditEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("audit events = %d, want 2", len(events))
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
	audits    map[string][]*domainplugin.AuditEvent
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		services:  map[string]*domainplugin.Service{},
		instances: map[string][]*domainplugin.Instance{},
		routes:    map[string][]*domainplugin.Route{},
		audits:    map[string][]*domainplugin.AuditEvent{},
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

func (r *memoryRepo) RecordPluginAudit(ctx context.Context, event domainplugin.AuditEvent) error {
	_ = ctx
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.ID = int64(len(r.audits[event.PluginKey]) + 1)
	eventCopy := event
	r.audits[event.PluginKey] = append(r.audits[event.PluginKey], &eventCopy)
	return nil
}

func (r *memoryRepo) ListPluginAuditEvents(ctx context.Context, pluginKey string, limit int) ([]*domainplugin.AuditEvent, error) {
	_ = ctx
	events := r.audits[pluginKey]
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	out := make([]*domainplugin.AuditEvent, 0, len(events))
	for i := len(events) - 1; i >= 0; i-- {
		out = append(out, events[i])
	}
	return out, nil
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

func (r *memoryRepo) SetServiceStatus(ctx context.Context, pluginKey string, status domainplugin.ServiceStatus, now time.Time) error {
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

func (r *memoryRepo) SetInstanceStatus(ctx context.Context, pluginKey string, instanceID string, status domainplugin.InstanceStatus, now time.Time) error {
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

func (r *memoryRepo) ListPluginServices(ctx context.Context) ([]*domainplugin.Service, error) {
	_ = ctx
	var services []*domainplugin.Service
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *memoryRepo) ListPluginInstances(ctx context.Context, pluginKey string) ([]*domainplugin.Instance, error) {
	_ = ctx
	return r.instances[pluginKey], nil
}

func (r *memoryRepo) ListHealthCheckTargets(ctx context.Context, now time.Time) ([]*domainplugin.Instance, error) {
	_ = ctx
	_ = now
	var instances []*domainplugin.Instance
	for pluginKey, service := range r.services {
		if service.Status != domainplugin.ServiceStatusActive {
			continue
		}
		for _, instance := range r.instances[pluginKey] {
			if instance.Status == domainplugin.InstanceStatusActive && instance.ManifestHash == service.CurrentManifestHash {
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
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

func (r *memoryRepo) UpdateInstanceHealth(ctx context.Context, pluginKey string, instanceID string, healthStatus domainplugin.HealthStatus, consecutiveFailures int, lastError string, checkedAt time.Time) (*domainplugin.Instance, error) {
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

type fakeHealthProbe struct {
	err error
}

func (p fakeHealthProbe) Probe(ctx context.Context, instance domainplugin.Instance) error {
	_ = ctx
	_ = instance
	return p.err
}
