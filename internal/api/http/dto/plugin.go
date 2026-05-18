package dto

import (
	"time"

	appplugin "github.com/rin721/keiyaku-go/internal/application/plugin"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
)

type PluginRegistrationRequest struct {
	SchemaVersion string            `json:"schema_version" binding:"required"`
	PluginKey     string            `json:"plugin_key" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Version       string            `json:"version" binding:"required"`
	InstanceID    string            `json:"instance_id" binding:"required"`
	Protocol      string            `json:"protocol" binding:"required"`
	BaseURL       string            `json:"base_url" binding:"required"`
	HealthPath    string            `json:"health_path"`
	OpenAPIURL    string            `json:"openapi_url,omitempty"`
	Routes        []PluginRouteSpec `json:"routes" binding:"required"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type PluginRouteSpec struct {
	Method            string            `json:"method" binding:"required"`
	MatchType         string            `json:"match_type" binding:"required"`
	Path              string            `json:"path" binding:"required"`
	UpstreamPath      string            `json:"upstream_path" binding:"required"`
	AuthPolicy        string            `json:"auth_policy"`
	TimeoutMS         int               `json:"timeout_ms"`
	ForwardAuthHeader bool              `json:"forward_auth_header"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

type PluginRegistrationResponse struct {
	PluginKey    string    `json:"plugin_key"`
	InstanceID   string    `json:"instance_id"`
	ManifestHash string    `json:"manifest_hash"`
	LeaseUntil   time.Time `json:"lease_until"`
}

type PluginHeartbeatResponse struct {
	PluginKey  string    `json:"plugin_key"`
	InstanceID string    `json:"instance_id"`
	LeaseUntil time.Time `json:"lease_until"`
}

type PluginServiceResponse struct {
	PluginKey           string            `json:"plugin_key"`
	Name                string            `json:"name"`
	Protocol            string            `json:"protocol"`
	CurrentManifestHash string            `json:"current_manifest_hash"`
	Status              string            `json:"status"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	DisabledAt          *time.Time        `json:"disabled_at,omitempty"`
}

type PluginInstanceResponse struct {
	InstanceID     string    `json:"instance_id"`
	Version        string    `json:"version"`
	BaseURL        string    `json:"base_url"`
	HealthPath     string    `json:"health_path"`
	ManifestHash   string    `json:"manifest_hash"`
	Status         string    `json:"status"`
	LastSeenAt     time.Time `json:"last_seen_at"`
	LeaseExpiresAt time.Time `json:"lease_expires_at"`
	LastError      string    `json:"last_error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PluginRouteResponse struct {
	Method            string            `json:"method"`
	MatchType         string            `json:"match_type"`
	Path              string            `json:"path"`
	UpstreamPath      string            `json:"upstream_path"`
	AuthPolicy        string            `json:"auth_policy"`
	TimeoutMS         int               `json:"timeout_ms"`
	ForwardAuthHeader bool              `json:"forward_auth_header"`
	Enabled           bool              `json:"enabled"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type PluginDetailResponse struct {
	Service   PluginServiceResponse    `json:"service"`
	Instances []PluginInstanceResponse `json:"instances"`
	Routes    []PluginRouteResponse    `json:"routes"`
}

func (r PluginRegistrationRequest) ToCommand(token string) appplugin.RegisterCommand {
	cmd := appplugin.RegisterCommand{
		Token:         token,
		SchemaVersion: r.SchemaVersion,
		PluginKey:     r.PluginKey,
		Name:          r.Name,
		Version:       r.Version,
		InstanceID:    r.InstanceID,
		Protocol:      r.Protocol,
		BaseURL:       r.BaseURL,
		HealthPath:    r.HealthPath,
		OpenAPIURL:    r.OpenAPIURL,
		Metadata:      r.Metadata,
	}
	for _, route := range r.Routes {
		cmd.Routes = append(cmd.Routes, appplugin.RouteCommand{
			Method:            route.Method,
			MatchType:         route.MatchType,
			Path:              route.Path,
			UpstreamPath:      route.UpstreamPath,
			AuthPolicy:        route.AuthPolicy,
			TimeoutMS:         route.TimeoutMS,
			ForwardAuthHeader: route.ForwardAuthHeader,
			Metadata:          route.Metadata,
		})
	}
	return cmd
}

func NewPluginRegistrationResponse(result *appplugin.RegisterResult) PluginRegistrationResponse {
	if result == nil {
		return PluginRegistrationResponse{}
	}
	return PluginRegistrationResponse{
		PluginKey:    result.PluginKey,
		InstanceID:   result.InstanceID,
		ManifestHash: result.ManifestHash,
		LeaseUntil:   result.LeaseUntil,
	}
}

func NewPluginHeartbeatResponse(result *appplugin.HeartbeatResult) PluginHeartbeatResponse {
	if result == nil {
		return PluginHeartbeatResponse{}
	}
	return PluginHeartbeatResponse{
		PluginKey:  result.PluginKey,
		InstanceID: result.InstanceID,
		LeaseUntil: result.LeaseUntil,
	}
}

func NewPluginServiceResponse(service *domainplugin.Service) PluginServiceResponse {
	if service == nil {
		return PluginServiceResponse{}
	}
	return PluginServiceResponse{
		PluginKey:           service.PluginKey,
		Name:                service.Name,
		Protocol:            string(service.Protocol),
		CurrentManifestHash: service.CurrentManifestHash,
		Status:              string(service.Status),
		Metadata:            service.Metadata,
		CreatedAt:           service.CreatedAt,
		UpdatedAt:           service.UpdatedAt,
		DisabledAt:          service.DisabledAt,
	}
}

func NewPluginInstanceResponse(instance *domainplugin.Instance) PluginInstanceResponse {
	if instance == nil {
		return PluginInstanceResponse{}
	}
	return PluginInstanceResponse{
		InstanceID:     instance.InstanceID,
		Version:        instance.Version,
		BaseURL:        instance.BaseURL,
		HealthPath:     instance.HealthPath,
		ManifestHash:   instance.ManifestHash,
		Status:         string(instance.Status),
		LastSeenAt:     instance.LastSeenAt,
		LeaseExpiresAt: instance.LeaseExpiresAt,
		LastError:      instance.LastError,
		CreatedAt:      instance.CreatedAt,
		UpdatedAt:      instance.UpdatedAt,
	}
}

func NewPluginRouteResponse(route *domainplugin.Route) PluginRouteResponse {
	if route == nil {
		return PluginRouteResponse{}
	}
	return PluginRouteResponse{
		Method:            string(route.Method),
		MatchType:         string(route.MatchType),
		Path:              route.Path,
		UpstreamPath:      route.UpstreamPath,
		AuthPolicy:        string(route.AuthPolicy),
		TimeoutMS:         int(route.Timeout / time.Millisecond),
		ForwardAuthHeader: route.ForwardAuthHeader,
		Enabled:           route.Enabled,
		Metadata:          route.Metadata,
		CreatedAt:         route.CreatedAt,
		UpdatedAt:         route.UpdatedAt,
	}
}
