package plugin

import (
	"net/url"
	"strings"
	"time"
)

type Protocol string

const (
	ProtocolHTTP Protocol = "http"
)

type ServiceStatus string

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusDisabled ServiceStatus = "disabled"
)

type InstanceStatus string

const (
	InstanceStatusActive       InstanceStatus = "active"
	InstanceStatusDisabled     InstanceStatus = "disabled"
	InstanceStatusIncompatible InstanceStatus = "incompatible"
)

type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type Method string

const (
	MethodAny    Method = "ANY"
	MethodGet    Method = "GET"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodPatch  Method = "PATCH"
	MethodDelete Method = "DELETE"
)

type MatchType string

const (
	MatchTypeExact  MatchType = "exact"
	MatchTypePrefix MatchType = "prefix"
)

type AuthPolicy string

const (
	AuthPolicyInherit       AuthPolicy = "inherit"
	AuthPolicyAuthenticated AuthPolicy = "authenticated"
	AuthPolicyRBAC          AuthPolicy = "rbac"
	AuthPolicyAdmin         AuthPolicy = "admin"
	AuthPolicyPublic        AuthPolicy = "public"
)

type AuditAction string

const (
	AuditActionRegister     AuditAction = "register"
	AuditActionHeartbeat    AuditAction = "heartbeat"
	AuditActionUnregister   AuditAction = "unregister"
	AuditActionHealthChange AuditAction = "health_change"
	AuditActionAdminDisable AuditAction = "admin_disable"
	AuditActionAdminEnable  AuditAction = "admin_enable"
	AuditActionRouteReplace AuditAction = "route_replace"
	AuditActionGatewayFail  AuditAction = "gateway_failure"
)

type Service struct {
	ID                  int64
	PluginKey           string
	Name                string
	Protocol            Protocol
	CurrentManifestHash string
	OpenAPIURL          string
	Status              ServiceStatus
	Metadata            map[string]string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DisabledAt          *time.Time
}

type Instance struct {
	ID                  int64
	PluginKey           string
	InstanceID          string
	Version             string
	BaseURL             string
	HealthPath          string
	ManifestHash        string
	Status              InstanceStatus
	HealthStatus        HealthStatus
	LastSeenAt          time.Time
	LeaseExpiresAt      time.Time
	LastCheckedAt       *time.Time
	ConsecutiveFailures int
	LastError           string
	LastErrorAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Route struct {
	ID                int64
	PluginKey         string
	ManifestHash      string
	RouteID           string
	Method            Method
	MatchType         MatchType
	GatewayPath       string
	UpstreamPath      string
	AuthPolicy        AuthPolicy
	Timeout           time.Duration
	ForwardAuthHeader bool
	Enabled           bool
	Metadata          map[string]string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RouteConflict struct {
	PluginKey   string
	RouteID     string
	Method      Method
	MatchType   MatchType
	GatewayPath string
}

type Registration struct {
	Service  Service
	Instance Instance
	Routes   []Route
}

type ResolvedRoute struct {
	Service  Service
	Instance Instance
	Route    Route
	Suffix   string
}

type AuditEvent struct {
	ID         int64
	PluginKey  string
	InstanceID string
	Action     AuditAction
	Message    string
	Metadata   map[string]string
	CreatedAt  time.Time
}

type GatewayMetric struct {
	PluginKey      string
	InstanceID     string
	RouteID        string
	GatewayPath    string
	UpstreamStatus int
	Duration       time.Duration
	GatewayError   string
	TraceID        string
}

type HealthMetric struct {
	PluginKey           string
	InstanceID          string
	PreviousStatus      HealthStatus
	CurrentStatus       HealthStatus
	ConsecutiveFailures int
}

func (i Instance) Routable(now time.Time, manifestHash string) bool {
	return i.Status == InstanceStatusActive &&
		i.ManifestHash == manifestHash &&
		!i.LeaseExpiresAt.Before(now) &&
		i.HealthStatus.Routable()
}

func (s HealthStatus) Routable() bool {
	switch s.Normalize() {
	case HealthStatusUnknown, HealthStatusHealthy:
		return true
	default:
		return false
	}
}

func (s HealthStatus) Normalize() HealthStatus {
	switch s {
	case HealthStatusHealthy, HealthStatusUnhealthy:
		return s
	default:
		return HealthStatusUnknown
	}
}

func (r Route) Matches(method string, path string) (string, bool) {
	if !r.Enabled {
		return "", false
	}
	if r.Method != MethodAny && string(r.Method) != method {
		return "", false
	}
	switch r.MatchType {
	case MatchTypeExact:
		if path == r.GatewayPath {
			return "", true
		}
	case MatchTypePrefix:
		if segmentPrefix(path, r.GatewayPath) {
			return strings.TrimPrefix(path, r.GatewayPath), true
		}
	}
	return "", false
}

func BuildUpstreamURL(baseURL string, upstreamPath string, suffix string, rawQuery string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = joinPath(upstreamPath, suffix)
	parsed.RawQuery = rawQuery
	return parsed.String(), nil
}

func segmentPrefix(path string, prefix string) bool {
	if prefix == "/" {
		return strings.HasPrefix(path, "/")
	}
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func joinPath(prefix string, suffix string) string {
	if prefix == "" {
		prefix = "/"
	}
	if suffix == "" {
		return prefix
	}
	if prefix == "/" {
		return suffix
	}
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(suffix, "/")
}
