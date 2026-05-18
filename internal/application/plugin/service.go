package plugin

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
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

var safeIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

type Config struct {
	Enabled              bool
	RegistrationTokens   []string
	AllowedPluginKeys    []string
	PublicPrefix         string
	HeartbeatTTL         time.Duration
	RequestTimeout       time.Duration
	AllowedHosts         []string
	AllowedCIDRs         []string
	AllowLoopback        bool
	AllowPublicRoutes    bool
	GatewaySigningSecret string
}

type Service struct {
	repo     port.PluginRegistryRepository
	config   Config
	now      func() time.Time
	mu       sync.Mutex
	counters map[string]int
	cidrs    []*net.IPNet
}

func NewService(repo port.PluginRegistryRepository, config Config) (*Service, error) {
	if config.PublicPrefix == "" {
		config.PublicPrefix = "/api/v1/extensions"
	}
	if config.HeartbeatTTL <= 0 {
		config.HeartbeatTTL = 30 * time.Second
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	var cidrs []*net.IPNet
	for _, raw := range config.AllowedCIDRs {
		_, cidr, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("parse plugins.allowed_cidrs %q: %w", raw, err)
		}
		cidrs = append(cidrs, cidr)
	}
	return &Service{
		repo:     repo,
		config:   config,
		now:      func() time.Time { return time.Now().UTC() },
		counters: map[string]int{},
		cidrs:    cidrs,
	}, nil
}

type RegisterCommand struct {
	Token         string
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
	Method            string
	MatchType         string
	Path              string
	UpstreamPath      string
	AuthPolicy        string
	TimeoutMS         int
	ForwardAuthHeader bool
	Metadata          map[string]string
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

type ResolveRouteQuery struct {
	PluginKey string
	Method    string
	Path      string
}

func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if !s.validToken(cmd.Token) {
		return nil, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidPluginToken)
	}
	if !s.allowedPluginKey(cmd.PluginKey) {
		return nil, apperror.New(apperror.CodeForbidden, apperror.MessagePluginKeyNotAllowed)
	}
	manifest, err := commandToManifest(cmd)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	if err := pkgplugin.ValidateManifest(manifest); err != nil {
		return nil, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest, err)
	}
	if err := s.validateBaseURL(manifest.BaseURL); err != nil {
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
			LastSeenAt:     now,
			LeaseExpiresAt: leaseUntil,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Routes: make([]domainplugin.Route, 0, len(manifest.Routes)),
	}
	for _, route := range manifest.Routes {
		route = pkgplugin.NormalizeRoute(route)
		timeout := time.Duration(route.TimeoutMS) * time.Millisecond
		if timeout <= 0 {
			timeout = s.config.RequestTimeout
		}
		registration.Routes = append(registration.Routes, domainplugin.Route{
			PluginKey:         manifest.PluginKey,
			ManifestHash:      hash,
			Method:            domainplugin.Method(route.Method),
			MatchType:         domainplugin.MatchType(route.MatchType),
			Path:              route.Path,
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
	if err := s.repo.UpsertRegistration(ctx, registration); err != nil {
		return nil, apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
	}
	return &RegisterResult{PluginKey: manifest.PluginKey, InstanceID: manifest.InstanceID, ManifestHash: hash, LeaseUntil: leaseUntil}, nil
}

func (s *Service) Heartbeat(ctx context.Context, token string, pluginKey string, instanceID string) (*HeartbeatResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	if !s.validToken(token) {
		return nil, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidPluginToken)
	}
	if !safeIDPattern.MatchString(pluginKey) || !safeIDPattern.MatchString(instanceID) {
		return nil, apperror.New(apperror.CodeInvalidArgument, apperror.MessageInvalidPluginManifest)
	}
	now := s.now()
	leaseUntil := now.Add(s.config.HeartbeatTTL)
	instance, err := s.repo.TouchInstance(ctx, pluginKey, instanceID, leaseUntil, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	return &HeartbeatResult{PluginKey: pluginKey, InstanceID: instance.InstanceID, LeaseUntil: leaseUntil}, nil
}

func (s *Service) Unregister(ctx context.Context, token string, pluginKey string, instanceID string) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return err
	}
	if !s.validToken(token) {
		return apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidPluginToken)
	}
	if err := s.repo.DisableInstance(ctx, pluginKey, instanceID, s.now()); err != nil {
		return mapPluginRepoError(err)
	}
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

func (s *Service) ResolveRoute(ctx context.Context, query ResolveRouteQuery) (*domainplugin.ResolvedRoute, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	service, instances, routes, err := s.repo.FindRoutable(ctx, query.PluginKey, s.now())
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	if service == nil || service.Status != domainplugin.ServiceStatusActive {
		return nil, apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginUnavailable)
	}
	if len(instances) == 0 {
		return nil, apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginUnavailable)
	}
	route, suffix, ok := bestRoute(query.Method, query.Path, routes)
	if !ok {
		return nil, apperror.New(apperror.CodeNotFound, apperror.MessagePluginRouteNotFound)
	}
	instance := s.pickInstance(query.PluginKey, instances)
	return &domainplugin.ResolvedRoute{Service: *service, Instance: *instance, Route: *route, Suffix: suffix}, nil
}

func (s *Service) GatewaySigningSecret() string {
	if s == nil {
		return ""
	}
	return s.config.GatewaySigningSecret
}

func (s *Service) AllowPublicRoutes() bool {
	return s != nil && s.config.AllowPublicRoutes
}

func (s *Service) ensureEnabled() error {
	if !s.config.Enabled {
		return apperror.New(apperror.CodeServiceUnavailable, apperror.MessagePluginRegistrationDisabled)
	}
	return nil
}

func (s *Service) validToken(token string) bool {
	if token == "" || len(s.config.RegistrationTokens) == 0 {
		return false
	}
	for _, expected := range s.config.RegistrationTokens {
		if expected == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
			return true
		}
	}
	return false
}

func (s *Service) allowedPluginKey(pluginKey string) bool {
	if !safeIDPattern.MatchString(pluginKey) {
		return false
	}
	if len(s.config.AllowedPluginKeys) == 0 {
		return false
	}
	for _, allowed := range s.config.AllowedPluginKeys {
		if pluginKey == allowed {
			return true
		}
	}
	return false
}

func (s *Service) validateBaseURL(raw string) error {
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
	if s.hostAllowed(host) {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("host %q is not in plugins.allowed_hosts", host)
	}
	if ip.IsLoopback() && !s.config.AllowLoopback {
		return fmt.Errorf("loopback plugin base_url is not allowed")
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local plugin base_url is not allowed")
	}
	for _, cidr := range s.cidrs {
		if cidr.Contains(ip) {
			return nil
		}
	}
	return fmt.Errorf("ip %q is not in plugins.allowed_cidrs", host)
}

func (s *Service) hostAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, allowed := range s.config.AllowedHosts {
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
			Method:            pkgplugin.Method(strings.ToUpper(strings.TrimSpace(route.Method))),
			MatchType:         pkgplugin.MatchType(strings.TrimSpace(route.MatchType)),
			Path:              strings.TrimSpace(route.Path),
			UpstreamPath:      strings.TrimSpace(route.UpstreamPath),
			AuthPolicy:        pkgplugin.AuthPolicy(strings.TrimSpace(route.AuthPolicy)),
			TimeoutMS:         route.TimeoutMS,
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
			pathLen:     len(route.Path),
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
