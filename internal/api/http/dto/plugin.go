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
	RouteID           string            `json:"route_id" binding:"required"`
	Method            string            `json:"method" binding:"required"`
	MatchType         string            `json:"match_type" binding:"required"`
	GatewayPath       string            `json:"gateway_path" binding:"required"`
	UpstreamPath      string            `json:"upstream_path" binding:"required"`
	AuthPolicy        string            `json:"auth_policy"`
	Timeout           string            `json:"timeout" binding:"required"`
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
	OpenAPIURL          string            `json:"openapi_url,omitempty"`
	Status              string            `json:"status"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	DisabledAt          *time.Time        `json:"disabled_at,omitempty"`
}

type PluginInstanceResponse struct {
	InstanceID          string     `json:"instance_id"`
	Version             string     `json:"version"`
	BaseURL             string     `json:"base_url"`
	HealthPath          string     `json:"health_path"`
	ManifestHash        string     `json:"manifest_hash"`
	Status              string     `json:"status"`
	HealthStatus        string     `json:"health_status"`
	LastSeenAt          time.Time  `json:"last_seen_at"`
	LeaseExpiresAt      time.Time  `json:"lease_expires_at"`
	LastCheckedAt       *time.Time `json:"last_checked_at,omitempty"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	LastError           string     `json:"last_error,omitempty"`
	LastErrorAt         *time.Time `json:"last_error_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type PluginRouteResponse struct {
	RouteID           string            `json:"route_id"`
	Method            string            `json:"method"`
	MatchType         string            `json:"match_type"`
	GatewayPath       string            `json:"gateway_path"`
	UpstreamPath      string            `json:"upstream_path"`
	AuthPolicy        string            `json:"auth_policy"`
	Timeout           string            `json:"timeout"`
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

type PluginAuditEventResponse struct {
	ID         int64             `json:"id"`
	PluginKey  string            `json:"plugin_key"`
	InstanceID string            `json:"instance_id,omitempty"`
	Action     string            `json:"action"`
	Message    string            `json:"message"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type PluginDiagnosticsResponse struct {
	Service           PluginServiceResponse              `json:"service"`
	MatchedRoute      *PluginRouteResponse               `json:"matched_route,omitempty"`
	RouteMatched      bool                               `json:"route_matched"`
	RouteSuffix       string                             `json:"route_suffix,omitempty"`
	CheckedAt         time.Time                          `json:"checked_at"`
	RoutableInstances int                                `json:"routable_instances"`
	Instances         []PluginInstanceDiagnosticResponse `json:"instances"`
}

type PluginInstanceDiagnosticResponse struct {
	Instance PluginInstanceResponse `json:"instance"`
	Routable bool                   `json:"routable"`
	Reasons  []string               `json:"reasons,omitempty"`
}

func (r PluginRegistrationRequest) ToCommand(signature appplugin.SignatureCommand) appplugin.RegisterCommand {
	cmd := appplugin.RegisterCommand{
		Signature:     signature,
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
			RouteID:           route.RouteID,
			Method:            route.Method,
			MatchType:         route.MatchType,
			GatewayPath:       route.GatewayPath,
			UpstreamPath:      route.UpstreamPath,
			AuthPolicy:        route.AuthPolicy,
			Timeout:           route.Timeout,
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
		OpenAPIURL:          service.OpenAPIURL,
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
		InstanceID:          instance.InstanceID,
		Version:             instance.Version,
		BaseURL:             instance.BaseURL,
		HealthPath:          instance.HealthPath,
		ManifestHash:        instance.ManifestHash,
		Status:              string(instance.Status),
		HealthStatus:        string(instance.HealthStatus.Normalize()),
		LastSeenAt:          instance.LastSeenAt,
		LeaseExpiresAt:      instance.LeaseExpiresAt,
		LastCheckedAt:       instance.LastCheckedAt,
		ConsecutiveFailures: instance.ConsecutiveFailures,
		LastError:           instance.LastError,
		LastErrorAt:         instance.LastErrorAt,
		CreatedAt:           instance.CreatedAt,
		UpdatedAt:           instance.UpdatedAt,
	}
}

func NewPluginRouteResponse(route *domainplugin.Route) PluginRouteResponse {
	if route == nil {
		return PluginRouteResponse{}
	}
	return PluginRouteResponse{
		RouteID:           route.RouteID,
		Method:            string(route.Method),
		MatchType:         string(route.MatchType),
		GatewayPath:       route.GatewayPath,
		UpstreamPath:      route.UpstreamPath,
		AuthPolicy:        string(route.AuthPolicy),
		Timeout:           route.Timeout.String(),
		ForwardAuthHeader: route.ForwardAuthHeader,
		Enabled:           route.Enabled,
		Metadata:          route.Metadata,
		CreatedAt:         route.CreatedAt,
		UpdatedAt:         route.UpdatedAt,
	}
}

func NewPluginAuditEventResponse(event *domainplugin.AuditEvent) PluginAuditEventResponse {
	if event == nil {
		return PluginAuditEventResponse{}
	}
	return PluginAuditEventResponse{
		ID:         event.ID,
		PluginKey:  event.PluginKey,
		InstanceID: event.InstanceID,
		Action:     string(event.Action),
		Message:    event.Message,
		Metadata:   event.Metadata,
		CreatedAt:  event.CreatedAt,
	}
}

func NewPluginDiagnosticsResponse(result *appplugin.PluginDiagnostics) PluginDiagnosticsResponse {
	if result == nil {
		return PluginDiagnosticsResponse{}
	}
	response := PluginDiagnosticsResponse{
		Service:           NewPluginServiceResponse(result.Service),
		RouteMatched:      result.RouteMatched,
		RouteSuffix:       result.RouteSuffix,
		CheckedAt:         result.CheckedAt,
		RoutableInstances: result.RoutableInstances,
		Instances:         make([]PluginInstanceDiagnosticResponse, 0, len(result.InstanceDiagnostics)),
	}
	if result.MatchedRoute != nil {
		route := NewPluginRouteResponse(result.MatchedRoute)
		response.MatchedRoute = &route
	}
	for _, item := range result.InstanceDiagnostics {
		response.Instances = append(response.Instances, PluginInstanceDiagnosticResponse{
			Instance: NewPluginInstanceResponse(item.Instance),
			Routable: item.Routable,
			Reasons:  append([]string(nil), item.Reasons...),
		})
	}
	return response
}
