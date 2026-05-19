package plugin

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
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
	cmd := validRegisterCommand("demo-plugin")
	cmd.OpenAPIURL = "https://plugins.internal/openapi.json"

	result, err := service.Register(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.PluginKey != "demo-plugin" || result.ManifestHash == "" {
		t.Fatalf("Register() plugin_key=%q manifest_hash=%q", result.PluginKey, result.ManifestHash)
	}
	if len(repo.routes["demo-plugin"]) != 2 {
		t.Fatalf("stored routes = %d, want 2", len(repo.routes["demo-plugin"]))
	}
	if got := repo.services["demo-plugin"].OpenAPIURL; got != "https://plugins.internal/openapi.json" {
		t.Fatalf("openapi_url = %q, want manifest value", got)
	}
}

func TestRegisterRejectsGatewayPathOutsidePublicPrefix(t *testing.T) {
	service := newTestService(t)
	cmd := validRegisterCommand("demo-plugin")
	cmd.Routes[0].GatewayPath = "/api/v1/users/me"

	_, err := service.Register(context.Background(), cmd)
	if err == nil {
		t.Fatal("Register() error is nil")
	}
}

func TestRegisterRejectsGatewayPathOutsideTrustedPrefix(t *testing.T) {
	service := newTestService(t)
	cmd := validRegisterCommand("demo-plugin")
	cmd.Routes[0].GatewayPath = "/api/v1/extensions/other-plugin/hello"

	_, err := service.Register(context.Background(), cmd)
	if err == nil {
		t.Fatal("Register() error is nil")
	}
}

func TestRegisterRejectsUntrustedAuthPolicyAndMethod(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled: true,
		TrustedPlugins: map[string]TrustedPluginConfig{
			"demo-plugin": {
				RegistrationSecret:  "01234567890123456789012345678901",
				GatewaySecret:       "abcdefghijklmnopqrstuvwxyz123456",
				AllowedHosts:        []string{"plugins.internal"},
				AllowedAuthPolicies: []string{"authenticated"},
				AllowedMethods:      []string{"GET"},
				AllowInsecureHTTP:   true,
			},
		},
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	cmd := validRegisterCommand("demo-plugin")
	cmd.Routes[0].AuthPolicy = "public"
	if _, err := service.Register(context.Background(), cmd); err == nil {
		t.Fatal("Register() public auth_policy error is nil")
	}
	cmd = validRegisterCommand("demo-plugin")
	cmd.Routes[0].Method = "POST"
	if _, err := service.Register(context.Background(), cmd); err == nil {
		t.Fatal("Register() POST method error is nil")
	}
}

func TestValidateOutboundURLRejectsInsecureAndLoopbackByDefault(t *testing.T) {
	service, err := NewService(newMemoryRepo(), Config{
		Enabled: true,
		TrustedPlugins: map[string]TrustedPluginConfig{
			"demo-plugin": {
				RegistrationSecret: "01234567890123456789012345678901",
				GatewaySecret:      "abcdefghijklmnopqrstuvwxyz123456",
				AllowedHosts:       []string{"plugins.internal"},
				AllowedCIDRs:       []string{"127.0.0.1/32"},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if err := service.ValidateOutboundURL("demo-plugin", "http://plugins.internal/hello?q=1"); err == nil {
		t.Fatal("ValidateOutboundURL() insecure http error is nil")
	}
	if err := service.ValidateOutboundURL("demo-plugin", "https://127.0.0.1/hello?q=1"); err == nil {
		t.Fatal("ValidateOutboundURL() loopback error is nil")
	}
	if err := service.ValidateResolvedOutboundIP("demo-plugin", net.ParseIP("10.0.0.5")); err == nil {
		t.Fatal("ValidateResolvedOutboundIP() private IP error is nil")
	}
	if err := service.ValidateResolvedOutboundIP("demo-plugin", net.ParseIP("8.8.8.8")); err != nil {
		t.Fatalf("ValidateResolvedOutboundIP() public IP error = %v", err)
	}
}

func TestRegisterRejectsReusedSignatureNonce(t *testing.T) {
	service := newTestService(t)
	cmd := validRegisterCommand("demo-plugin")
	if _, err := service.Register(context.Background(), cmd); err != nil {
		t.Fatalf("Register() first error = %v", err)
	}
	cmd.InstanceID = "demo-plugin-2"
	_, err := service.Register(context.Background(), cmd)
	if err == nil {
		t.Fatal("Register() second error is nil")
	}
}

func TestResolveRoutePrefersExactMethodAndSegmentPrefix(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	resolved, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/items/42"})
	if err != nil {
		t.Fatalf("ResolveRoute() error = %v", err)
	}
	if resolved.Route.GatewayPath != "/api/v1/extensions/demo-plugin/items" || resolved.Suffix != "/42" {
		t.Fatalf("resolved route path=%q suffix=%q", resolved.Route.GatewayPath, resolved.Suffix)
	}

	_, err = service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/items-extra"})
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

	_, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/hello"})
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

	_, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/hello"})
	if err == nil {
		t.Fatal("ResolveRoute() error is nil")
	}
}

func TestHealthCheckMarksUnhealthyAfterThreshold(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:             true,
		TrustedPlugins:      testTrustedPlugins(),
		HeartbeatTTL:        time.Minute,
		RequestTimeout:      time.Second,
		HealthCheckInterval: time.Second,
		UnhealthyThreshold:  2,
		MaxAuditQueryLimit:  20,
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
		Enabled:        true,
		TrustedPlugins: testTrustedPlugins(),
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
		RouteCacheTTL:  time.Minute,
	}, WithAuditRepository(repo))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/hello"}); err != nil {
		t.Fatalf("ResolveRoute() before disable error = %v", err)
	}
	if err := service.DisableService(context.Background(), "demo-plugin"); err != nil {
		t.Fatalf("DisableService() error = %v", err)
	}
	if _, err := service.ResolveRoute(context.Background(), ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/hello"}); err == nil {
		t.Fatal("ResolveRoute() after disable error is nil")
	}
}

func TestRegisterRecordsAuditEvents(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:        true,
		TrustedPlugins: testTrustedPlugins(),
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
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

func TestRegisterRejectsCrossPluginRouteConflict(t *testing.T) {
	repo := newMemoryRepo()
	otherTrust := testTrustedPlugins()["demo-plugin"]
	otherTrust.AllowedGatewayPrefixes = []string{"/api/v1/extensions/demo-plugin"}
	service, err := NewService(repo, Config{
		Enabled: true,
		TrustedPlugins: map[string]TrustedPluginConfig{
			"demo-plugin":  testTrustedPlugins()["demo-plugin"],
			"other-plugin": otherTrust,
		},
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() first error = %v", err)
	}
	cmd := validRegisterCommand("other-plugin")
	cmd.Routes[0].GatewayPath = "/api/v1/extensions/demo-plugin/hello"
	cmd.Routes[1].GatewayPath = "/api/v1/extensions/demo-plugin/items"
	_, err = service.Register(context.Background(), cmd)
	if err == nil {
		t.Fatal("Register() conflicting route error is nil")
	}
}

func TestRegisterRejectsRouteTimeoutAboveMax(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewService(repo, Config{
		Enabled:         true,
		TrustedPlugins:  testTrustedPlugins(),
		HeartbeatTTL:    time.Minute,
		RequestTimeout:  time.Second,
		MaxRouteTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	cmd := validRegisterCommand("demo-plugin")
	cmd.Routes[0].Timeout = "2s"
	if _, err := service.Register(context.Background(), cmd); err == nil {
		t.Fatal("Register() timeout error is nil")
	}
}

func TestHealthCheckSkipsExpiredLease(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	repo.instances["demo-plugin"][0].LeaseExpiresAt = time.Now().UTC().Add(-time.Minute)

	if err := service.CheckHealth(context.Background(), fakeHealthProbe{err: errors.New("probe failed")}); err != nil {
		t.Fatalf("CheckHealth() error = %v", err)
	}
	if got := repo.instances["demo-plugin"][0].ConsecutiveFailures; got != 0 {
		t.Fatalf("consecutive failures = %d, want 0", got)
	}
}

func TestDiagnoseReturnsRoutabilityReasons(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	repo.instances["demo-plugin"][0].HealthStatus = domainplugin.HealthStatusUnhealthy

	diagnostics, err := service.Diagnose(context.Background(), "demo-plugin", ResolveRouteQuery{Method: "GET", Path: "/api/v1/extensions/demo-plugin/hello"})
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}
	if !diagnostics.RouteMatched || diagnostics.MatchedRoute == nil {
		t.Fatal("Diagnose() did not match route")
	}
	if diagnostics.RoutableInstances != 0 || len(diagnostics.InstanceDiagnostics) != 1 {
		t.Fatalf("diagnostics routable=%d instances=%d", diagnostics.RoutableInstances, len(diagnostics.InstanceDiagnostics))
	}
	if got := diagnostics.InstanceDiagnostics[0].Reasons; len(got) == 0 || got[0] != "health_unhealthy" {
		t.Fatalf("diagnostic reasons = %#v, want health_unhealthy", got)
	}
}

func TestMaintainPrunesAndDisablesStaleInstances(t *testing.T) {
	repo := newMemoryRepo()
	service := newTestServiceWithRepo(t, repo)
	if _, err := service.Register(context.Background(), validRegisterCommand("demo-plugin")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	repo.nonces["demo-plugin:old"] = now.Add(-time.Minute)
	repo.audits["demo-plugin"] = append(repo.audits["demo-plugin"], &domainplugin.AuditEvent{PluginKey: "demo-plugin", CreatedAt: now.AddDate(0, 0, -40)})
	repo.instances["demo-plugin"][0].LeaseExpiresAt = now.Add(-2 * time.Minute)

	result, err := service.Maintain(context.Background())
	if err != nil {
		t.Fatalf("Maintain() error = %v", err)
	}
	if result.PrunedSignatureNonces == 0 || result.PrunedAuditEvents == 0 || result.DisabledStaleInstances == 0 {
		t.Fatalf("maintenance result = %#v, want all counters > 0", result)
	}
	if got := repo.instances["demo-plugin"][0].Status; got != domainplugin.InstanceStatusDisabled {
		t.Fatalf("stale instance status = %q, want disabled", got)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	return newTestServiceWithRepo(t, newMemoryRepo())
}

func newTestServiceWithRepo(t *testing.T, repo *memoryRepo) *Service {
	t.Helper()
	service, err := NewService(repo, Config{
		Enabled:        true,
		TrustedPlugins: testTrustedPlugins(),
		HeartbeatTTL:   time.Minute,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func validRegisterCommand(pluginKey string) RegisterCommand {
	return RegisterCommand{
		Signature:     testSignature(pluginKey, "POST", "/api/v1/plugins/registrations", nil),
		SchemaVersion: pkgplugin.DefaultSchemaVersion,
		PluginKey:     pluginKey,
		Name:          "Demo",
		Version:       "0.1.0",
		InstanceID:    "demo-plugin-1",
		Protocol:      "http",
		BaseURL:       "http://plugins.internal:9090",
		HealthPath:    "/healthz",
		Routes: []RouteCommand{
			{RouteID: "hello", Method: "GET", MatchType: "exact", GatewayPath: "/api/v1/extensions/" + pluginKey + "/hello", UpstreamPath: "/hello", AuthPolicy: "authenticated", Timeout: "1s"},
			{RouteID: "items", Method: "GET", MatchType: "prefix", GatewayPath: "/api/v1/extensions/" + pluginKey + "/items", UpstreamPath: "/api/items", AuthPolicy: "authenticated", Timeout: "1s"},
		},
	}
}

func testTrustedPlugins() map[string]TrustedPluginConfig {
	return map[string]TrustedPluginConfig{
		"demo-plugin": {
			RegistrationSecret: "01234567890123456789012345678901",
			GatewaySecret:      "abcdefghijklmnopqrstuvwxyz123456",
			AllowedHosts:       []string{"plugins.internal"},
			AllowInsecureHTTP:  true,
		},
	}
}

func testSignature(pluginKey string, method string, path string, body []byte) SignatureCommand {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	nonce := pluginKey + "-nonce-" + method + "-" + path
	return SignatureCommand{
		PluginKey: pluginKey,
		Method:    method,
		Path:      path,
		RawQuery:  "",
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: pkgplugin.Sign(method, path, "", timestamp, nonce, pkgplugin.BodySHA256(body), "01234567890123456789012345678901"),
		Body:      body,
	}
}

type memoryRepo struct {
	services  map[string]*domainplugin.Service
	instances map[string][]*domainplugin.Instance
	routes    map[string][]*domainplugin.Route
	audits    map[string][]*domainplugin.AuditEvent
	nonces    map[string]time.Time
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		services:  map[string]*domainplugin.Service{},
		instances: map[string][]*domainplugin.Instance{},
		routes:    map[string][]*domainplugin.Route{},
		audits:    map[string][]*domainplugin.AuditEvent{},
		nonces:    map[string]time.Time{},
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

func (r *memoryRepo) FindRouteConflict(ctx context.Context, pluginKey string, routes []domainplugin.Route) (*domainplugin.RouteConflict, error) {
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
				if testRoutesOverlap(*existing, route) {
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

func testRoutesOverlap(left domainplugin.Route, right domainplugin.Route) bool {
	if left.Method != right.Method && left.Method != domainplugin.MethodAny && right.Method != domainplugin.MethodAny {
		return false
	}
	switch {
	case left.MatchType == domainplugin.MatchTypeExact && right.MatchType == domainplugin.MatchTypeExact:
		return left.GatewayPath == right.GatewayPath
	case left.MatchType == domainplugin.MatchTypePrefix && right.MatchType == domainplugin.MatchTypePrefix:
		return testPathPrefix(left.GatewayPath, right.GatewayPath) || testPathPrefix(right.GatewayPath, left.GatewayPath)
	case left.MatchType == domainplugin.MatchTypeExact && right.MatchType == domainplugin.MatchTypePrefix:
		return testPathPrefix(left.GatewayPath, right.GatewayPath)
	case left.MatchType == domainplugin.MatchTypePrefix && right.MatchType == domainplugin.MatchTypeExact:
		return testPathPrefix(right.GatewayPath, left.GatewayPath)
	default:
		return false
	}
}

func testPathPrefix(path string, prefix string) bool {
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, strings.TrimRight(prefix, "/")+"/")
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
			if instance.Status == domainplugin.InstanceStatusActive && instance.ManifestHash == service.CurrentManifestHash && !instance.LeaseExpiresAt.Before(now) {
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

func (r *memoryRepo) FindRoutable(ctx context.Context, now time.Time) ([]*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
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

func (r *memoryRepo) UseSignatureNonce(ctx context.Context, pluginKey string, nonce string, expiresAt time.Time, now time.Time) error {
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

func (r *memoryRepo) PruneSignatureNonces(ctx context.Context, now time.Time) (int64, error) {
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

func (r *memoryRepo) PrunePluginAuditEvents(ctx context.Context, before time.Time) (int64, error) {
	_ = ctx
	var count int64
	for pluginKey, events := range r.audits {
		kept := events[:0]
		for _, event := range events {
			if event.CreatedAt.Before(before) {
				count++
				continue
			}
			kept = append(kept, event)
		}
		r.audits[pluginKey] = kept
	}
	return count, nil
}

func (r *memoryRepo) DisableStalePluginInstances(ctx context.Context, staleBefore time.Time, now time.Time) (int64, error) {
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

type fakeHealthProbe struct {
	err error
}

func (p fakeHealthProbe) Probe(ctx context.Context, instance domainplugin.Instance) error {
	_ = ctx
	_ = instance
	return p.err
}
