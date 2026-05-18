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

type Service struct {
	ID                  int64
	PluginKey           string
	Name                string
	Protocol            Protocol
	CurrentManifestHash string
	Status              ServiceStatus
	Metadata            map[string]string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DisabledAt          *time.Time
}

type Instance struct {
	ID             int64
	PluginKey      string
	InstanceID     string
	Version        string
	BaseURL        string
	HealthPath     string
	ManifestHash   string
	Status         InstanceStatus
	LastSeenAt     time.Time
	LeaseExpiresAt time.Time
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Route struct {
	ID                int64
	PluginKey         string
	ManifestHash      string
	Method            Method
	MatchType         MatchType
	Path              string
	UpstreamPath      string
	AuthPolicy        AuthPolicy
	Timeout           time.Duration
	ForwardAuthHeader bool
	Enabled           bool
	Metadata          map[string]string
	CreatedAt         time.Time
	UpdatedAt         time.Time
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

func (i Instance) Routable(now time.Time, manifestHash string) bool {
	return i.Status == InstanceStatusActive &&
		i.ManifestHash == manifestHash &&
		!i.LeaseExpiresAt.Before(now)
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
		if path == r.Path {
			return "", true
		}
	case MatchTypePrefix:
		if segmentPrefix(path, r.Path) {
			return strings.TrimPrefix(path, r.Path), true
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
