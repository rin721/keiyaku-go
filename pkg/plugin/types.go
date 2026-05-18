package plugin

import "time"

const (
	DefaultSchemaVersion = "v1"
	DefaultProtocol      = ProtocolHTTP
)

type Protocol string

const (
	ProtocolHTTP Protocol = "http"
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

type Manifest struct {
	SchemaVersion     string            `json:"schema_version"`
	PluginKey         string            `json:"plugin_key"`
	Name              string            `json:"name"`
	Version           string            `json:"version"`
	InstanceID        string            `json:"instance_id"`
	Protocol          Protocol          `json:"protocol"`
	BaseURL           string            `json:"base_url"`
	HealthPath        string            `json:"health_path"`
	OpenAPIURL        string            `json:"openapi_url,omitempty"`
	Routes            []Route           `json:"routes"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	RegistrationToken string            `json:"-"`
}

type Route struct {
	Method            Method            `json:"method"`
	MatchType         MatchType         `json:"match_type"`
	Path              string            `json:"path"`
	UpstreamPath      string            `json:"upstream_path"`
	AuthPolicy        AuthPolicy        `json:"auth_policy"`
	TimeoutMS         int               `json:"timeout_ms"`
	ForwardAuthHeader bool              `json:"forward_auth_header"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

type RegisterResponse struct {
	PluginKey    string    `json:"plugin_key"`
	InstanceID   string    `json:"instance_id"`
	ManifestHash string    `json:"manifest_hash"`
	LeaseUntil   time.Time `json:"lease_until"`
}

type HeartbeatResponse struct {
	PluginKey  string    `json:"plugin_key"`
	InstanceID string    `json:"instance_id"`
	LeaseUntil time.Time `json:"lease_until"`
}
