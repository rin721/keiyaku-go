package plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
)

const defaultRequestTimeout = 5 * time.Second
const defaultUnhealthyThreshold = 3
const defaultMaxAuditQueryLimit = 200
const defaultAuditRetentionDays = 30
const defaultMaxRegistrationBodyBytes int64 = 1 << 20
const defaultMaxGatewayBodyBytes int64 = 10 << 20

var safeIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

type Config struct {
	Enabled                  bool
	TrustedPlugins           map[string]TrustedPluginConfig
	PublicPrefix             string
	HeartbeatTTL             time.Duration
	RequestTimeout           time.Duration
	MaxRegistrationBodyBytes int64
	MaxGatewayBodyBytes      int64
	MaxRouteTimeout          time.Duration
	HealthCheckInterval      time.Duration
	HealthCheckTimeout       time.Duration
	UnhealthyThreshold       int
	RouteCacheTTL            time.Duration
	MaintenanceInterval      time.Duration
	AuditRetentionDays       int
	MaxAuditQueryLimit       int
	AllowPublicRoutes        bool
	SignatureSkew            time.Duration
}

type TrustedPluginConfig struct {
	RegistrationSecret string
	GatewaySecret      string
	AllowedHosts       []string
	AllowedCIDRs       []string
	AllowLoopback      bool
}

type Service struct {
	repo      port.PluginRegistryRepository
	auditRepo port.PluginAuditRepository
	metrics   port.PluginMetrics
	config    Config
	now       func() time.Time
	mu        sync.Mutex
	counters  map[string]int
	trusted   map[string]trustedPlugin
	cache     *routeCacheEntry
}

type Option func(*Service)

func WithAuditRepository(repo port.PluginAuditRepository) Option {
	return func(service *Service) {
		service.auditRepo = repo
	}
}

func WithMetrics(metrics port.PluginMetrics) Option {
	return func(service *Service) {
		service.metrics = metrics
	}
}

type routeCacheEntry struct {
	services  []*domainplugin.Service
	instances []*domainplugin.Instance
	routes    []*domainplugin.Route
	expiresAt time.Time
}

type trustedPlugin struct {
	registrationSecret string
	gatewaySecret      string
	allowedHosts       []string
	allowedCIDRs       []*net.IPNet
	allowLoopback      bool
}

func NewService(repo port.PluginRegistryRepository, config Config, options ...Option) (*Service, error) {
	if config.PublicPrefix == "" {
		config.PublicPrefix = "/api/v1/extensions"
	}
	config.PublicPrefix = pkgplugin.NormalizePublicPrefix(config.PublicPrefix)
	if config.HeartbeatTTL <= 0 {
		config.HeartbeatTTL = 30 * time.Second
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	if config.MaxRegistrationBodyBytes <= 0 {
		config.MaxRegistrationBodyBytes = defaultMaxRegistrationBodyBytes
	}
	if config.MaxGatewayBodyBytes <= 0 {
		config.MaxGatewayBodyBytes = defaultMaxGatewayBodyBytes
	}
	if config.MaxRouteTimeout <= 0 {
		config.MaxRouteTimeout = config.RequestTimeout
	}
	if config.SignatureSkew <= 0 {
		config.SignatureSkew = pkgplugin.DefaultSignatureSkew
	}
	if config.AuditRetentionDays <= 0 {
		config.AuditRetentionDays = defaultAuditRetentionDays
	}
	trusted := make(map[string]trustedPlugin, len(config.TrustedPlugins))
	for pluginKey, rawTrust := range config.TrustedPlugins {
		pluginKey = strings.TrimSpace(pluginKey)
		if !safeIDPattern.MatchString(pluginKey) {
			return nil, fmt.Errorf("plugins.trusted_plugins contains invalid plugin key %q", pluginKey)
		}
		var cidrs []*net.IPNet
		for _, raw := range rawTrust.AllowedCIDRs {
			_, cidr, err := net.ParseCIDR(strings.TrimSpace(raw))
			if err != nil {
				return nil, fmt.Errorf("parse plugins.trusted_plugins.%s.allowed_cidrs %q: %w", pluginKey, raw, err)
			}
			cidrs = append(cidrs, cidr)
		}
		trusted[pluginKey] = trustedPlugin{
			registrationSecret: rawTrust.RegistrationSecret,
			gatewaySecret:      rawTrust.GatewaySecret,
			allowedHosts:       rawTrust.AllowedHosts,
			allowedCIDRs:       cidrs,
			allowLoopback:      rawTrust.AllowLoopback,
		}
	}
	service := &Service{
		repo:     repo,
		config:   config,
		now:      func() time.Time { return time.Now().UTC() },
		counters: map[string]int{},
		trusted:  trusted,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service, nil
}

type RegisterCommand struct {
	Signature     SignatureCommand
	SchemaVersion string
	PluginKey     string
	Name          string
	Version       string
	InstanceID    string
	Protocol      string
	BaseURL       string
	HealthPath    string
	OpenAPIURL    string
	Routes        []RouteCommand
	Metadata      map[string]string
}

type RouteCommand struct {
	RouteID           string
	Method            string
	MatchType         string
	GatewayPath       string
	UpstreamPath      string
	AuthPolicy        string
	Timeout           string
	ForwardAuthHeader bool
	Metadata          map[string]string
}

type SignatureCommand struct {
	PluginKey string
	Method    string
	Path      string
	Timestamp string
	Nonce     string
	Signature string
	Body      []byte
}

type RegisterResult struct {
	PluginKey    string
	InstanceID   string
	ManifestHash string
	LeaseUntil   time.Time
}

type HeartbeatResult struct {
	PluginKey  string
	InstanceID string
	LeaseUntil time.Time
}

type PluginDetail struct {
	Service   *domainplugin.Service
	Instances []*domainplugin.Instance
	Routes    []*domainplugin.Route
}

type PluginDiagnostics struct {
	Service             *domainplugin.Service
	MatchedRoute        *domainplugin.Route
	RouteSuffix         string
	RouteMatched        bool
	CheckedAt           time.Time
	RoutableInstances   int
	InstanceDiagnostics []InstanceDiagnostic
}

type InstanceDiagnostic struct {
	Instance *domainplugin.Instance
	Routable bool
	Reasons  []string
}

type MaintenanceResult struct {
	PrunedSignatureNonces  int64
	PrunedAuditEvents      int64
	DisabledStaleInstances int64
}

type ResolveRouteQuery struct {
	Method string
	Path   string
}

func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	manifest, err := commandToManifest(cmd)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	if err := s.validateControlSignature(ctx, manifest.PluginKey, cmd.Signature); err != nil {
		return nil, err
	}
	if err := pkgplugin.ValidateManifest(manifest); err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	if err := s.validateManifestGatewayPaths(manifest); err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	if err := s.validateBaseURL(manifest.PluginKey, manifest.BaseURL); err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	hash, err := pkgplugin.ManifestHash(manifest)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	now := s.now()
	leaseUntil := now.Add(s.config.HeartbeatTTL)
	registration := domainplugin.Registration{
		Service: domainplugin.Service{
			PluginKey:           manifest.PluginKey,
			Name:                manifest.Name,
			Protocol:            domainplugin.Protocol(manifest.Protocol),
			CurrentManifestHash: hash,
			Status:              domainplugin.ServiceStatusActive,
			Metadata:            manifest.Metadata,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		Instance: domainplugin.Instance{
			PluginKey:      manifest.PluginKey,
			InstanceID:     manifest.InstanceID,
			Version:        manifest.Version,
			BaseURL:        strings.TrimRight(manifest.BaseURL, "/"),
			HealthPath:     manifest.HealthPath,
			ManifestHash:   hash,
			Status:         domainplugin.InstanceStatusActive,
			HealthStatus:   domainplugin.HealthStatusUnknown,
			LastSeenAt:     now,
			LeaseExpiresAt: leaseUntil,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Routes: make([]domainplugin.Route, 0, len(manifest.Routes)),
	}
	for _, route := range manifest.Routes {
		route = pkgplugin.NormalizeRoute(route)
		timeout, err := pkgplugin.ParseRouteTimeout(route.Timeout)
		if err != nil {
			return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
		}
		if s.config.MaxRouteTimeout > 0 && timeout > s.config.MaxRouteTimeout {
			return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, fmt.Errorf("route timeout %s exceeds plugins.max_route_timeout %s", timeout, s.config.MaxRouteTimeout))
		}
		registration.Routes = append(registration.Routes, domainplugin.Route{
			PluginKey:         manifest.PluginKey,
			ManifestHash:      hash,
			RouteID:           route.RouteID,
			Method:            domainplugin.Method(route.Method),
			MatchType:         domainplugin.MatchType(route.MatchType),
			GatewayPath:       route.GatewayPath,
			UpstreamPath:      route.UpstreamPath,
			AuthPolicy:        domainplugin.AuthPolicy(route.AuthPolicy),
			Timeout:           timeout,
			ForwardAuthHeader: route.ForwardAuthHeader,
			Enabled:           true,
			Metadata:          route.Metadata,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}
	if conflict, err := s.repo.FindRouteConflict(ctx, manifest.PluginKey, registration.Routes); err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
	} else if conflict != nil {
		return nil, apperror.Wrap(apperror.CodeConflict, apperror.MessageConflict, fmt.Errorf("plugin route conflicts with %s/%s %s %s %s", conflict.PluginKey, conflict.RouteID, conflict.Method, conflict.MatchType, conflict.GatewayPath))
	}
	if err := s.repo.UpsertRegistration(ctx, registration); err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
	}
	s.invalidateRouteCache(manifest.PluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  manifest.PluginKey,
		InstanceID: manifest.InstanceID,
		Action:     domainplugin.AuditActionRegister,
		Message:    "plugin instance registered",
		Metadata: map[string]string{
			"manifest_hash": hash,
			"version":       manifest.Version,
		},
	})
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey: manifest.PluginKey,
		Action:    domainplugin.AuditActionRouteReplace,
		Message:   "plugin routes replaced",
		Metadata: map[string]string{
			"manifest_hash": hash,
			"route_count":   strconv.Itoa(len(registration.Routes)),
		},
	})
	return &RegisterResult{PluginKey: manifest.PluginKey, InstanceID: manifest.InstanceID, ManifestHash: hash, LeaseUntil: leaseUntil}, nil
}

func (s *Service) Heartbeat(ctx context.Context, signature SignatureCommand, pluginKey string, instanceID string) (*HeartbeatResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if !safeIDPattern.MatchString(pluginKey) || !safeIDPattern.MatchString(instanceID) {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	if err := s.validateControlSignature(ctx, pluginKey, signature); err != nil {
		return nil, err
	}
	now := s.now()
	leaseUntil := now.Add(s.config.HeartbeatTTL)
	instance, err := s.repo.TouchInstance(ctx, pluginKey, instanceID, leaseUntil, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  pluginKey,
		InstanceID: instance.InstanceID,
		Action:     domainplugin.AuditActionHeartbeat,
		Message:    "plugin instance heartbeat accepted",
		Metadata: map[string]string{
			"lease_until": leaseUntil.Format(time.RFC3339Nano),
		},
	})
	return &HeartbeatResult{PluginKey: pluginKey, InstanceID: instance.InstanceID, LeaseUntil: leaseUntil}, nil
}

func (s *Service) Unregister(ctx context.Context, signature SignatureCommand, pluginKey string, instanceID string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return err
	}
	if !safeIDPattern.MatchString(pluginKey) || !safeIDPattern.MatchString(instanceID) {
		return apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	if err := s.validateControlSignature(ctx, pluginKey, signature); err != nil {
		return err
	}
	if err := s.repo.DisableInstance(ctx, pluginKey, instanceID, s.now()); err != nil {
		return mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  pluginKey,
		InstanceID: instanceID,
		Action:     domainplugin.AuditActionUnregister,
		Message:    "plugin instance unregistered",
	})
	return nil
}

func (s *Service) List(ctx context.Context) ([]*domainplugin.Service, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	return s.repo.ListPluginServices(ctx)
}

func (s *Service) Get(ctx context.Context, pluginKey string) (*PluginDetail, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	service, instances, routes, err := s.repo.GetPluginService(ctx, pluginKey)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	return &PluginDetail{Service: service, Instances: instances, Routes: routes}, nil
}

func (s *Service) ListInstances(ctx context.Context, pluginKey string) ([]*domainplugin.Instance, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	_, instances, _, err := s.repo.GetPluginService(ctx, pluginKey)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	return instances, nil
}

func (s *Service) DisableService(ctx context.Context, pluginKey string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) {
		return apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	now := s.now()
	if err := s.repo.SetServiceStatus(ctx, pluginKey, domainplugin.ServiceStatusDisabled, now); err != nil {
		return mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey: pluginKey,
		Action:    domainplugin.AuditActionAdminDisable,
		Message:   "plugin service disabled",
	})
	return nil
}

func (s *Service) EnableService(ctx context.Context, pluginKey string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) {
		return apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	now := s.now()
	if err := s.repo.SetServiceStatus(ctx, pluginKey, domainplugin.ServiceStatusActive, now); err != nil {
		return mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey: pluginKey,
		Action:    domainplugin.AuditActionAdminEnable,
		Message:   "plugin service enabled",
	})
	return nil
}

func (s *Service) DisableInstance(ctx context.Context, pluginKey string, instanceID string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) || !safeIDPattern.MatchString(instanceID) {
		return apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	now := s.now()
	if err := s.repo.SetInstanceStatus(ctx, pluginKey, instanceID, domainplugin.InstanceStatusDisabled, now); err != nil {
		return mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  pluginKey,
		InstanceID: instanceID,
		Action:     domainplugin.AuditActionAdminDisable,
		Message:    "plugin instance disabled",
	})
	return nil
}

func (s *Service) EnableInstance(ctx context.Context, pluginKey string, instanceID string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) || !safeIDPattern.MatchString(instanceID) {
		return apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	now := s.now()
	if err := s.repo.SetInstanceStatus(ctx, pluginKey, instanceID, domainplugin.InstanceStatusActive, now); err != nil {
		return mapPluginRepoError(err)
	}
	s.invalidateRouteCache(pluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  pluginKey,
		InstanceID: instanceID,
		Action:     domainplugin.AuditActionAdminEnable,
		Message:    "plugin instance enabled",
	})
	return nil
}

func (s *Service) Diagnose(ctx context.Context, pluginKey string, query ResolveRouteQuery) (*PluginDiagnostics, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	service, instances, routes, err := s.repo.GetPluginService(ctx, pluginKey)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	now := s.now()
	result := &PluginDiagnostics{
		Service:             service,
		CheckedAt:           now,
		InstanceDiagnostics: make([]InstanceDiagnostic, 0, len(instances)),
	}
	if strings.TrimSpace(query.Method) != "" && strings.TrimSpace(query.Path) != "" {
		if route, suffix, ok := bestRoute(strings.ToUpper(strings.TrimSpace(query.Method)), strings.TrimSpace(query.Path), routes); ok {
			result.MatchedRoute = route
			result.RouteSuffix = suffix
			result.RouteMatched = true
		}
	}
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		reasons := instanceRoutabilityReasons(service, instance, now)
		routable := len(reasons) == 0
		if routable {
			result.RoutableInstances++
		}
		result.InstanceDiagnostics = append(result.InstanceDiagnostics, InstanceDiagnostic{
			Instance: instance,
			Routable: routable,
			Reasons:  reasons,
		})
	}
	return result, nil
}

func (s *Service) ListAuditEvents(ctx context.Context, pluginKey string, limit int) ([]*domainplugin.AuditEvent, error) {
	if s == nil || s.repo == nil || s.auditRepo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !safeIDPattern.MatchString(pluginKey) {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	if _, _, _, err := s.repo.GetPluginService(ctx, pluginKey); err != nil {
		return nil, mapPluginRepoError(err)
	}
	limit = s.normalizeAuditLimit(limit)
	events, err := s.auditRepo.ListPluginAuditEvents(ctx, pluginKey, limit)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	return events, nil
}

func (s *Service) ResolveRoute(ctx context.Context, query ResolveRouteQuery) (*domainplugin.ResolvedRoute, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	now := s.now()
	services, instances, routes, err := s.findRoutable(ctx, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	route, suffix, ok := bestRoute(query.Method, query.Path, routes)
	if !ok {
		return nil, apperror.New(apperror.CodeNotFound, apperror.MessagePluginRouteNotFound)
	}
	service := serviceByPluginKey(services, route.PluginKey)
	if service == nil || service.Status != domainplugin.ServiceStatusActive {
		return nil, apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginUnavailable)
	}
	instances = filterRoutableInstances(instancesByPluginKey(instances, route.PluginKey), now, service.CurrentManifestHash)
	if len(instances) == 0 {
		return nil, apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginUnavailable)
	}
	instance := s.pickInstance(route.PluginKey, instances)
	return &domainplugin.ResolvedRoute{Service: *service, Instance: *instance, Route: *route, Suffix: suffix}, nil
}

func (s *Service) GatewaySigningSecret(pluginKey string) string {
	if s == nil {
		return ""
	}
	trust, ok := s.trusted[pluginKey]
	if !ok {
		return ""
	}
	return trust.gatewaySecret
}

func (s *Service) AllowPublicRoutes() bool {
	return s != nil && s.config.AllowPublicRoutes
}

func (s *Service) RequestTimeout() time.Duration {
	if s == nil || s.config.RequestTimeout <= 0 {
		return defaultRequestTimeout
	}
	return s.config.RequestTimeout
}

func (s *Service) MaxRegistrationBodyBytes() int64 {
	if s == nil || s.config.MaxRegistrationBodyBytes <= 0 {
		return defaultMaxRegistrationBodyBytes
	}
	return s.config.MaxRegistrationBodyBytes
}

func (s *Service) MaxGatewayBodyBytes() int64 {
	if s == nil || s.config.MaxGatewayBodyBytes <= 0 {
		return defaultMaxGatewayBodyBytes
	}
	return s.config.MaxGatewayBodyBytes
}

func (s *Service) EffectiveRouteTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		timeout = s.RequestTimeout()
	}
	if s != nil && s.config.MaxRouteTimeout > 0 && timeout > s.config.MaxRouteTimeout {
		return s.config.MaxRouteTimeout
	}
	return timeout
}

func (s *Service) HealthCheckInterval() time.Duration {
	if s == nil {
		return 0
	}
	return s.config.HealthCheckInterval
}

func (s *Service) HealthCheckTimeout() time.Duration {
	if s == nil {
		return 0
	}
	return s.config.HealthCheckTimeout
}

func (s *Service) MaintenanceInterval() time.Duration {
	if s == nil {
		return 0
	}
	return s.config.MaintenanceInterval
}

func (s *Service) CheckHealth(ctx context.Context, probe port.PluginHealthProbe) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if probe == nil || !s.config.Enabled || s.config.HealthCheckInterval <= 0 {
		return nil
	}
	targets, err := s.repo.ListHealthCheckTargets(ctx, s.now())
	if err != nil {
		return mapPluginRepoError(err)
	}
	for _, target := range targets {
		if target == nil {
			continue
		}
		s.checkInstanceHealth(ctx, probe, *target)
	}
	return nil
}

func (s *Service) Maintain(ctx context.Context) (*MaintenanceResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !s.config.Enabled {
		return &MaintenanceResult{}, nil
	}
	now := s.now()
	result := &MaintenanceResult{}
	prunedNonces, err := s.repo.PruneSignatureNonces(ctx, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	result.PrunedSignatureNonces = prunedNonces
	if s.config.AuditRetentionDays > 0 {
		prunedAudits, err := s.repo.PrunePluginAuditEvents(ctx, now.AddDate(0, 0, -s.config.AuditRetentionDays))
		if err != nil {
			return nil, mapPluginRepoError(err)
		}
		result.PrunedAuditEvents = prunedAudits
	}
	staleBefore := now.Add(-s.config.HeartbeatTTL)
	disabled, err := s.repo.DisableStalePluginInstances(ctx, staleBefore, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	result.DisabledStaleInstances = disabled
	if disabled > 0 {
		s.invalidateRouteCache("")
	}
	return result, nil
}

func (s *Service) RecordGatewayRequest(ctx context.Context, metric domainplugin.GatewayMetric) {
	if s == nil || s.metrics == nil {
		return
	}
	s.metrics.RecordPluginGateway(ctx, metric)
}

func (s *Service) RecordGatewayFailure(ctx context.Context, metric domainplugin.GatewayMetric) {
	if s == nil {
		return
	}
	switch metric.GatewayError {
	case "upstream_connect", "upstream_request", "upstream_timeout", "plugin_unavailable":
	default:
		return
	}
	metadata := map[string]string{
		"gateway_error": metric.GatewayError,
		"trace_id":      metric.TraceID,
	}
	if metric.RouteID != "" {
		metadata["route_id"] = metric.RouteID
	}
	if metric.GatewayPath != "" {
		metadata["gateway_path"] = metric.GatewayPath
	}
	if metric.UpstreamStatus > 0 {
		metadata["upstream_status"] = strconv.Itoa(metric.UpstreamStatus)
	}
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  metric.PluginKey,
		InstanceID: metric.InstanceID,
		Action:     domainplugin.AuditActionGatewayFail,
		Message:    "plugin gateway failure",
		Metadata:   metadata,
	})
}

func (s *Service) ensureEnabled() error {
	if !s.config.Enabled {
		return apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginRegistrationDisabled)
	}
	return nil
}

func (s *Service) validateControlSignature(ctx context.Context, pluginKey string, signature SignatureCommand) error {
	if !safeIDPattern.MatchString(pluginKey) || signature.PluginKey != pluginKey {
		return apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature)
	}
	trust, ok := s.trusted[pluginKey]
	if !ok || strings.TrimSpace(trust.registrationSecret) == "" {
		return apperror.New(apperror.CodeForbidden, apperror.MessagePluginKeyNotTrusted)
	}
	parts := pkgplugin.SignatureParts{
		PluginKey: signature.PluginKey,
		Timestamp: signature.Timestamp,
		Nonce:     signature.Nonce,
		Signature: signature.Signature,
	}
	if err := pkgplugin.Verify(signature.Method, signature.Path, signature.Body, parts, trust.registrationSecret, s.now(), s.config.SignatureSkew); err != nil {
		return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature, err)
	}
	timestamp, err := pkgplugin.ParseSignatureTimestamp(signature.Timestamp)
	if err != nil {
		return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature, err)
	}
	expiresAt := timestamp.Add(s.config.SignatureSkew)
	if expiresAt.Before(s.now()) {
		expiresAt = s.now().Add(time.Minute)
	}
	if err := s.repo.UseSignatureNonce(ctx, pluginKey, signature.Nonce, expiresAt, s.now()); err != nil {
		if errors.Is(err, derrors.ErrConflict) {
			return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessagePluginNonceReused, err)
		}
		return apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
	}
	return nil
}

func (s *Service) validateManifestGatewayPaths(manifest pkgplugin.Manifest) error {
	for _, route := range manifest.Routes {
		route = pkgplugin.NormalizeRoute(route)
		if err := pkgplugin.ValidateGatewayPath(route.GatewayPath, s.config.PublicPrefix); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) checkInstanceHealth(ctx context.Context, probe port.PluginHealthProbe, instance domainplugin.Instance) {
	previous := instance.HealthStatus.Normalize()
	err := probe.Probe(ctx, instance)
	now := s.now()
	nextStatus := previous
	failures := instance.ConsecutiveFailures
	lastError := ""
	if err != nil {
		failures++
		lastError = err.Error()
		if failures >= s.unhealthyThreshold() {
			nextStatus = domainplugin.HealthStatusUnhealthy
		}
	} else {
		failures = 0
		nextStatus = domainplugin.HealthStatusHealthy
	}
	updated, updateErr := s.repo.UpdateInstanceHealth(ctx, instance.PluginKey, instance.InstanceID, nextStatus, failures, lastError, now)
	if updateErr != nil || updated == nil {
		return
	}
	current := updated.HealthStatus.Normalize()
	if previous == current {
		return
	}
	s.invalidateRouteCache(instance.PluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  instance.PluginKey,
		InstanceID: instance.InstanceID,
		Action:     domainplugin.AuditActionHealthChange,
		Message:    "plugin instance health changed",
		Metadata: map[string]string{
			"from":                 string(previous),
			"to":                   string(current),
			"consecutive_failures": strconv.Itoa(updated.ConsecutiveFailures),
		},
	})
	if s.metrics != nil {
		s.metrics.RecordPluginHealth(ctx, domainplugin.HealthMetric{
			PluginKey:           instance.PluginKey,
			InstanceID:          instance.InstanceID,
			PreviousStatus:      previous,
			CurrentStatus:       current,
			ConsecutiveFailures: updated.ConsecutiveFailures,
		})
	}
}

func (s *Service) unhealthyThreshold() int {
	if s == nil || s.config.UnhealthyThreshold <= 0 {
		return defaultUnhealthyThreshold
	}
	return s.config.UnhealthyThreshold
}

func (s *Service) normalizeAuditLimit(limit int) int {
	maxLimit := defaultMaxAuditQueryLimit
	if s != nil && s.config.MaxAuditQueryLimit > 0 {
		maxLimit = s.config.MaxAuditQueryLimit
	}
	if limit <= 0 || limit > maxLimit {
		return maxLimit
	}
	return limit
}

func (s *Service) findRoutable(ctx context.Context, now time.Time) ([]*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	if s.config.RouteCacheTTL > 0 {
		if services, instances, routes, ok := s.routeCacheGet(now); ok {
			return services, instances, routes, nil
		}
	}
	services, instances, routes, err := s.repo.FindRoutable(ctx, now)
	if err != nil {
		return nil, nil, nil, err
	}
	if s.config.RouteCacheTTL > 0 {
		s.routeCacheSet(services, instances, routes, now.Add(s.config.RouteCacheTTL))
	}
	return cloneServices(services), cloneInstances(instances), cloneRoutes(routes), nil
}

func (s *Service) routeCacheGet(now time.Time) ([]*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache == nil || !now.Before(s.cache.expiresAt) {
		s.cache = nil
		return nil, nil, nil, false
	}
	return cloneServices(s.cache.services), cloneInstances(s.cache.instances), cloneRoutes(s.cache.routes), true
}

func (s *Service) routeCacheSet(services []*domainplugin.Service, instances []*domainplugin.Instance, routes []*domainplugin.Route, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = &routeCacheEntry{
		services:  cloneServices(services),
		instances: cloneInstances(instances),
		routes:    cloneRoutes(routes),
		expiresAt: expiresAt,
	}
}

func (s *Service) invalidateRouteCache(pluginKey string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = nil
}

func (s *Service) recordAudit(ctx context.Context, event domainplugin.AuditEvent) {
	if s == nil || s.auditRepo == nil || event.PluginKey == "" {
		return
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = s.now()
	}
	event.Metadata = safeAuditMetadata(event.Metadata)
	_ = s.auditRepo.RecordPluginAudit(ctx, event)
}

func (s *Service) validateBaseURL(pluginKey string, raw string) error {
	trust, ok := s.trusted[pluginKey]
	if !ok {
		return fmt.Errorf("plugin key %q is not trusted", pluginKey)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("base_url must not include userinfo, query, or fragment")
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("base_url host is required")
	}
	if hostAllowed(trust.allowedHosts, host) {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("host %q is not in trusted plugin allowed_hosts", host)
	}
	if ip.IsLoopback() && !trust.allowLoopback {
		return fmt.Errorf("loopback plugin base_url is not allowed")
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local plugin base_url is not allowed")
	}
	for _, cidr := range trust.allowedCIDRs {
		if cidr.Contains(ip) {
			return nil
		}
	}
	return fmt.Errorf("ip %q is not in trusted plugin allowed_cidrs", host)
}

func hostAllowed(allowedHosts []string, host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(allowed, ".")))
		if allowed == "" {
			continue
		}
		if host == allowed {
			return true
		}
		if strings.HasPrefix(allowed, "*.") && strings.HasSuffix(host, strings.TrimPrefix(allowed, "*")) {
			return true
		}
	}
	return false
}

func (s *Service) pickInstance(pluginKey string, instances []*domainplugin.Instance) *domainplugin.Instance {
	if len(instances) == 1 {
		return instances[0]
	}
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].InstanceID < instances[j].InstanceID
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.counters[pluginKey] % len(instances)
	s.counters[pluginKey]++
	return instances[index]
}

func filterRoutableInstances(instances []*domainplugin.Instance, now time.Time, manifestHash string) []*domainplugin.Instance {
	filtered := make([]*domainplugin.Instance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		if instance.Routable(now, manifestHash) {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

func instanceRoutabilityReasons(service *domainplugin.Service, instance *domainplugin.Instance, now time.Time) []string {
	var reasons []string
	if service == nil {
		reasons = append(reasons, "service_missing")
	} else {
		if service.Status != domainplugin.ServiceStatusActive {
			reasons = append(reasons, "service_disabled")
		}
		if instance != nil && instance.ManifestHash != service.CurrentManifestHash {
			reasons = append(reasons, "manifest_mismatch")
		}
	}
	if instance == nil {
		return append(reasons, "instance_missing")
	}
	if instance.Status != domainplugin.InstanceStatusActive {
		reasons = append(reasons, "instance_"+string(instance.Status))
	}
	if instance.LeaseExpiresAt.Before(now) {
		reasons = append(reasons, "lease_expired")
	}
	if !instance.HealthStatus.Routable() {
		reasons = append(reasons, "health_"+string(instance.HealthStatus.Normalize()))
	}
	return reasons
}

func serviceByPluginKey(services []*domainplugin.Service, pluginKey string) *domainplugin.Service {
	for _, service := range services {
		if service != nil && service.PluginKey == pluginKey {
			return service
		}
	}
	return nil
}

func instancesByPluginKey(instances []*domainplugin.Instance, pluginKey string) []*domainplugin.Instance {
	filtered := make([]*domainplugin.Instance, 0, len(instances))
	for _, instance := range instances {
		if instance != nil && instance.PluginKey == pluginKey {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

func cloneServices(services []*domainplugin.Service) []*domainplugin.Service {
	clones := make([]*domainplugin.Service, 0, len(services))
	for _, service := range services {
		clone := cloneService(service)
		if clone != nil {
			clones = append(clones, clone)
		}
	}
	return clones
}

func cloneService(service *domainplugin.Service) *domainplugin.Service {
	if service == nil {
		return nil
	}
	clone := *service
	clone.Metadata = cloneStringMap(service.Metadata)
	return &clone
}

func cloneInstances(instances []*domainplugin.Instance) []*domainplugin.Instance {
	clones := make([]*domainplugin.Instance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		clone := *instance
		clones = append(clones, &clone)
	}
	return clones
}

func cloneRoutes(routes []*domainplugin.Route) []*domainplugin.Route {
	clones := make([]*domainplugin.Route, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		clone := *route
		clone.Metadata = cloneStringMap(route.Metadata)
		clones = append(clones, &clone)
	}
	return clones
}

func cloneStringMap(value map[string]string) map[string]string {
	if len(value) == 0 {
		return map[string]string{}
	}
	clone := make(map[string]string, len(value))
	for key, item := range value {
		clone[key] = item
	}
	return clone
}

func safeAuditMetadata(value map[string]string) map[string]string {
	if len(value) == 0 {
		return map[string]string{}
	}
	safe := make(map[string]string, len(value))
	for key, item := range value {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") ||
			strings.Contains(lower, "authorization") ||
			strings.Contains(lower, "cookie") ||
			strings.Contains(lower, "secret") ||
			strings.Contains(lower, "password") {
			continue
		}
		safe[key] = item
	}
	return safe
}

func commandToManifest(cmd RegisterCommand) (pkgplugin.Manifest, error) {
	manifest := pkgplugin.Manifest{
		SchemaVersion: cmd.SchemaVersion,
		PluginKey:     strings.TrimSpace(cmd.PluginKey),
		Name:          strings.TrimSpace(cmd.Name),
		Version:       strings.TrimSpace(cmd.Version),
		InstanceID:    strings.TrimSpace(cmd.InstanceID),
		Protocol:      pkgplugin.Protocol(strings.TrimSpace(cmd.Protocol)),
		BaseURL:       strings.TrimSpace(cmd.BaseURL),
		HealthPath:    strings.TrimSpace(cmd.HealthPath),
		OpenAPIURL:    strings.TrimSpace(cmd.OpenAPIURL),
		Metadata:      cmd.Metadata,
	}
	for _, route := range cmd.Routes {
		manifest.Routes = append(manifest.Routes, pkgplugin.Route{
			RouteID:           strings.TrimSpace(route.RouteID),
			Method:            pkgplugin.Method(strings.ToUpper(strings.TrimSpace(route.Method))),
			MatchType:         pkgplugin.MatchType(strings.TrimSpace(route.MatchType)),
			GatewayPath:       strings.TrimSpace(route.GatewayPath),
			UpstreamPath:      strings.TrimSpace(route.UpstreamPath),
			AuthPolicy:        pkgplugin.AuthPolicy(strings.TrimSpace(route.AuthPolicy)),
			Timeout:           strings.TrimSpace(route.Timeout),
			ForwardAuthHeader: route.ForwardAuthHeader,
			Metadata:          route.Metadata,
		})
	}
	return pkgplugin.NormalizeManifest(manifest), nil
}

func bestRoute(method string, path string, routes []*domainplugin.Route) (*domainplugin.Route, string, bool) {
	type candidate struct {
		route       *domainplugin.Route
		suffix      string
		methodScore int
		matchScore  int
		pathLen     int
	}
	var candidates []candidate
	for _, route := range routes {
		if route == nil {
			continue
		}
		suffix, ok := route.Matches(method, path)
		if !ok {
			continue
		}
		methodScore := 0
		if route.Method != domainplugin.MethodAny {
			methodScore = 1
		}
		matchScore := 0
		if route.MatchType == domainplugin.MatchTypeExact {
			matchScore = 1
		}
		candidates = append(candidates, candidate{
			route:       route,
			suffix:      suffix,
			methodScore: methodScore,
			matchScore:  matchScore,
			pathLen:     len(route.GatewayPath),
		})
	}
	if len(candidates) == 0 {
		return nil, "", false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].methodScore != candidates[j].methodScore {
			return candidates[i].methodScore > candidates[j].methodScore
		}
		if candidates[i].matchScore != candidates[j].matchScore {
			return candidates[i].matchScore > candidates[j].matchScore
		}
		return candidates[i].pathLen > candidates[j].pathLen
	})
	return candidates[0].route, candidates[0].suffix, true
}

func mapPluginRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, derrors.ErrNotFound) {
		return apperror.Wrap(apperror.CodeNotFound, apperror.MessageNotFound, err)
	}
	return apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
}
