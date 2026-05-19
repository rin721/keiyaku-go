package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/dto"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	appplugin "github.com/rin721/keiyaku-go/internal/application/plugin"
	"github.com/rin721/keiyaku-go/internal/application/port"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	"github.com/rin721/keiyaku-go/internal/observability/trace"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
	"go.uber.org/zap"
)

type PluginHandler struct {
	service    *appplugin.Service
	tokens     port.TokenIssuer
	authorizer port.Authorizer
	client     *http.Client
	logger     *zap.Logger
}

type PluginHandlerOption func(*PluginHandler)

func WithPluginLogger(logger *zap.Logger) PluginHandlerOption {
	return func(handler *PluginHandler) {
		handler.logger = logger
	}
}

func WithPluginHTTPClient(client *http.Client) PluginHandlerOption {
	return func(handler *PluginHandler) {
		if client != nil {
			handler.client = client
		}
	}
}

func NewPluginHandler(service *appplugin.Service, tokens port.TokenIssuer, authorizer port.Authorizer, options ...PluginHandlerOption) *PluginHandler {
	handler := &PluginHandler{service: service, tokens: tokens, authorizer: authorizer, client: &http.Client{}, logger: zap.NewNop()}
	for _, option := range options {
		if option != nil {
			option(handler)
		}
	}
	if handler.logger == nil {
		handler.logger = zap.NewNop()
	}
	return handler
}

// Register handles remote plugin registration.
// @Summary Register remote plugin
// @Description Register a remote HTTP plugin instance and its routes.
// @Tags Plugin
// @Accept json
// @Produce json
// @Param request body dto.PluginRegistrationRequest true "Plugin registration payload"
// @Success 200 {object} dto.PluginRegistrationResponse "OK"
// @Failure 400 {object} response.Body "Invalid request"
// @Failure 401 {object} response.Body "Invalid plugin signature"
// @Failure 403 {object} response.Body "Plugin key not allowed"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/registrations [post]
func (h *PluginHandler) Register(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	var req dto.PluginRegistrationRequest
	body, signature, err := h.signedBody(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
		return
	}
	result, err := h.service.Register(c.Request.Context(), req.ToCommand(signature))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewPluginRegistrationResponse(result))
}

// Heartbeat refreshes a plugin instance lease.
// @Summary Heartbeat plugin instance
// @Description Refresh the lease of a registered plugin instance.
// @Tags Plugin
// @Produce json
// @Param plugin_key path string true "Plugin key"
// @Param instance_id path string true "Instance ID"
// @Success 200 {object} dto.PluginHeartbeatResponse "OK"
// @Failure 401 {object} response.Body "Invalid plugin signature"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id}/heartbeat [post]
func (h *PluginHandler) Heartbeat(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	signature := h.signature(c, nil)
	result, err := h.service.Heartbeat(c.Request.Context(), signature, c.Param("plugin_key"), c.Param("instance_id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewPluginHeartbeatResponse(result))
}

// Unregister disables a plugin instance.
// @Summary Unregister plugin instance
// @Description Disable a registered plugin instance.
// @Tags Plugin
// @Produce json
// @Param plugin_key path string true "Plugin key"
// @Param instance_id path string true "Instance ID"
// @Success 200 {object} response.Body "OK"
// @Failure 401 {object} response.Body "Invalid plugin signature"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id} [delete]
func (h *PluginHandler) Unregister(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	signature := h.signature(c, nil)
	if err := h.service.Unregister(c.Request.Context(), signature, c.Param("plugin_key"), c.Param("instance_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// List returns registered plugins.
// @Summary List plugins
// @Description List registered remote plugin services.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Success 200 {object} []dto.PluginServiceResponse "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins [get]
func (h *PluginHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	services, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	items := make([]dto.PluginServiceResponse, 0, len(services))
	for _, service := range services {
		items = append(items, dto.NewPluginServiceResponse(service))
	}
	response.OK(c, items)
}

// Get returns one plugin with instances and routes.
// @Summary Get plugin
// @Description Get a registered remote plugin service.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Success 200 {object} dto.PluginDetailResponse "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key} [get]
func (h *PluginHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	detail, err := h.service.Get(c.Request.Context(), c.Param("plugin_key"))
	if err != nil {
		response.Error(c, err)
		return
	}
	result := dto.PluginDetailResponse{Service: dto.NewPluginServiceResponse(detail.Service)}
	for _, instance := range detail.Instances {
		result.Instances = append(result.Instances, dto.NewPluginInstanceResponse(instance))
	}
	for _, route := range detail.Routes {
		result.Routes = append(result.Routes, dto.NewPluginRouteResponse(route))
	}
	response.OK(c, result)
}

// ListInstances returns plugin instance states.
// @Summary List plugin instances
// @Description List runtime instances for a registered remote plugin service.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Success 200 {object} []dto.PluginInstanceResponse "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances [get]
func (h *PluginHandler) ListInstances(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	instances, err := h.service.ListInstances(c.Request.Context(), c.Param("plugin_key"))
	if err != nil {
		response.Error(c, err)
		return
	}
	items := make([]dto.PluginInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		items = append(items, dto.NewPluginInstanceResponse(instance))
	}
	response.OK(c, items)
}

// Disable disables a plugin service.
// @Summary Disable plugin
// @Description Disable a registered remote plugin service so gateway routing stops selecting it.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Success 200 {object} response.Body "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/disable [post]
func (h *PluginHandler) Disable(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	if err := h.service.DisableService(c.Request.Context(), c.Param("plugin_key")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// Enable enables a plugin service.
// @Summary Enable plugin
// @Description Enable a registered remote plugin service after an administrative disable.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Success 200 {object} response.Body "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/enable [post]
func (h *PluginHandler) Enable(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	if err := h.service.EnableService(c.Request.Context(), c.Param("plugin_key")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// DisableInstance disables one plugin instance.
// @Summary Disable plugin instance
// @Description Disable one registered plugin instance.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Param instance_id path string true "Instance ID"
// @Success 200 {object} response.Body "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id}/disable [post]
func (h *PluginHandler) DisableInstance(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	if err := h.service.DisableInstance(c.Request.Context(), c.Param("plugin_key"), c.Param("instance_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// EnableInstance enables one plugin instance.
// @Summary Enable plugin instance
// @Description Enable one registered plugin instance after an administrative disable.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Param instance_id path string true "Instance ID"
// @Success 200 {object} response.Body "OK"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id}/enable [post]
func (h *PluginHandler) EnableInstance(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	if err := h.service.EnableInstance(c.Request.Context(), c.Param("plugin_key"), c.Param("instance_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// Diagnostics returns plugin routability diagnostics.
// @Summary Diagnose plugin routability
// @Description Return route matching and instance routability diagnostics for one plugin.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Param method query string false "HTTP method to match"
// @Param path query string false "Gateway path to match"
// @Success 200 {object} dto.PluginDiagnosticsResponse "OK"
// @Failure 400 {object} response.Body "Invalid request"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/diagnostics [get]
func (h *PluginHandler) Diagnostics(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	result, err := h.service.Diagnose(c.Request.Context(), c.Param("plugin_key"), appplugin.ResolveRouteQuery{
		Method: c.Query("method"),
		Path:   c.Query("path"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewPluginDiagnosticsResponse(result))
}

// ListAuditEvents returns plugin audit events.
// @Summary List plugin audit events
// @Description List recent audit events for a registered remote plugin service.
// @Tags Plugin
// @Produce json
// @Security bearerAuth
// @Param plugin_key path string true "Plugin key"
// @Param limit query int false "Maximum number of events"
// @Success 200 {object} []dto.PluginAuditEventResponse "OK"
// @Failure 400 {object} response.Body "Invalid request"
// @Failure 401 {object} response.Body "Unauthorized"
// @Failure 403 {object} response.Body "Forbidden"
// @Failure 404 {object} response.Body "Plugin not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/audit-events [get]
func (h *PluginHandler) ListAuditEvents(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	limit, err := parseOptionalPositiveInt(c.Query("limit"))
	if err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidArgument, err))
		return
	}
	events, err := h.service.ListAuditEvents(c.Request.Context(), c.Param("plugin_key"), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	items := make([]dto.PluginAuditEventResponse, 0, len(events))
	for _, event := range events {
		items = append(items, dto.NewPluginAuditEventResponse(event))
	}
	response.OK(c, items)
}

func (h *PluginHandler) signedBody(c *gin.Context) ([]byte, appplugin.SignatureCommand, error) {
	body, err := pkgplugin.ReadLimitedBody(c.Request.Body, h.service.MaxRegistrationBodyBytes())
	if err != nil {
		return nil, appplugin.SignatureCommand{}, bodyReadError(err)
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, h.signature(c, body), nil
}

func (h *PluginHandler) signature(c *gin.Context, body []byte) appplugin.SignatureCommand {
	parts := pkgplugin.SignatureFromHeader(c.Request.Header)
	return appplugin.SignatureCommand{
		PluginKey: parts.PluginKey,
		Method:    c.Request.Method,
		Path:      c.Request.URL.EscapedPath(),
		Timestamp: parts.Timestamp,
		Nonce:     parts.Nonce,
		Signature: parts.Signature,
		Body:      body,
	}
}

func (h *PluginHandler) Gateway(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	start := time.Now()
	traceID := trace.IDFromContext(c.Request.Context())
	if traceID == "" {
		traceID = trace.NewID()
	}
	var metric domainplugin.GatewayMetric
	metric.TraceID = traceID
	defer func() {
		metric.Duration = time.Since(start)
		if metric.UpstreamStatus == 0 {
			metric.UpstreamStatus = c.Writer.Status()
		}
		h.logGateway(metric)
		h.service.RecordGatewayRequest(c.Request.Context(), metric)
		h.service.RecordGatewayFailure(c.Request.Context(), metric)
	}()
	resolved, err := h.service.ResolveRoute(c.Request.Context(), appplugin.ResolveRouteQuery{
		Method: c.Request.Method,
		Path:   c.Request.URL.Path,
	})
	if err != nil {
		metric.GatewayError = gatewayErrorName(err)
		response.Error(c, err)
		return
	}
	metric.PluginKey = resolved.Service.PluginKey
	metric.InstanceID = resolved.Instance.InstanceID
	metric.RouteID = resolved.Route.RouteID
	metric.GatewayPath = resolved.Route.GatewayPath
	claims, ok, err := h.authorizeGateway(c, resolved.Route)
	if err != nil {
		metric.GatewayError = "auth"
		response.Error(c, err)
		return
	}
	upstreamURL, err := domainplugin.BuildUpstreamURL(resolved.Instance.BaseURL, resolved.Route.UpstreamPath, resolved.Suffix, c.Request.URL.RawQuery)
	if err != nil {
		metric.GatewayError = "upstream_request"
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return
	}
	metric.GatewayError = h.forward(c, upstreamURL, resolved, claims, ok, traceID)
	metric.UpstreamStatus = c.Writer.Status()
}

func (h *PluginHandler) authorizeGateway(c *gin.Context, route domainplugin.Route) (port.TokenClaims, bool, error) {
	if route.AuthPolicy == domainplugin.AuthPolicyPublic {
		if !h.service.AllowPublicRoutes() {
			return port.TokenClaims{}, false, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionDenied)
		}
		return port.TokenClaims{}, false, nil
	}
	claims, err := h.parseClaims(c)
	if err != nil {
		return port.TokenClaims{}, false, err
	}
	switch route.AuthPolicy {
	case domainplugin.AuthPolicyAdmin:
		if !hasRole(claims, "admin") {
			return port.TokenClaims{}, false, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionDenied)
		}
	case domainplugin.AuthPolicyRBAC:
		if h.authorizer == nil {
			return port.TokenClaims{}, false, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionNotReady)
		}
		allowed := false
		for _, role := range claims.Roles {
			ok, err := h.authorizer.Allow(role, c.Request.URL.Path, c.Request.Method)
			if err != nil {
				return port.TokenClaims{}, false, apperror.Wrap(apperror.CodeDependency, apperror.MessagePermissionCheckFail, err)
			}
			if ok {
				allowed = true
				break
			}
		}
		if !allowed {
			return port.TokenClaims{}, false, apperror.New(apperror.CodeForbidden, apperror.MessagePermissionDenied)
		}
	}
	return claims, true, nil
}

func (h *PluginHandler) parseClaims(c *gin.Context) (port.TokenClaims, error) {
	if h.tokens == nil {
		return port.TokenClaims{}, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken)
	}
	token := bearerToken(c)
	if token == "" {
		return port.TokenClaims{}, apperror.New(apperror.CodeUnauthorized, apperror.MessageMissingAuthHeader)
	}
	claims, err := h.tokens.ParseAccessToken(c.Request.Context(), token)
	if err != nil {
		return port.TokenClaims{}, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken)
	}
	return claims, nil
}

func (h *PluginHandler) forward(c *gin.Context, upstreamURL string, resolved *domainplugin.ResolvedRoute, claims port.TokenClaims, hasClaims bool, traceID string) string {
	timeout := h.service.EffectiveRouteTimeout(resolved.Route.Timeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	body, err := pkgplugin.ReadLimitedBody(c.Request.Body, h.service.MaxGatewayBodyBytes())
	if err != nil {
		response.Error(c, bodyReadError(err))
		return "upstream_request"
	}
	req, err := http.NewRequestWithContext(ctx, c.Request.Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return "upstream_request"
	}
	copyPluginHeaders(req.Header, c.Request.Header, resolved.Route.ForwardAuthHeader)
	if traceID == "" {
		traceID = trace.NewID()
	}
	req.Header.Set(trace.HeaderName, traceID)
	req.Header.Set("X-Keiyaku-Plugin-Key", resolved.Service.PluginKey)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)
	req.Header.Set("X-Forwarded-Proto", forwardedProto(c.Request))
	req.Header.Set("X-Forwarded-Method", c.Request.Method)
	if hasClaims {
		req.Header.Set("X-Keiyaku-User-ID", int64String(claims.UserID))
		req.Header.Set("X-Keiyaku-Username", claims.Username)
		req.Header.Set("X-Keiyaku-User-Roles", strings.Join(claims.Roles, ","))
	}
	if secret := h.service.GatewaySigningSecret(resolved.Service.PluginKey); secret != "" {
		if err := pkgplugin.SignRequest(req, resolved.Service.PluginKey, secret, body, time.Time{}, ""); err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
			return "upstream_request"
		}
	}
	client := h.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		if isTimeout(err) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			response.Error(c, apperror.Wrap(apperror.CodeGatewayTimeout, apperror.MessagePluginUpstreamTimeout, err))
			return "upstream_timeout"
		}
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return "upstream_connect"
	}
	defer resp.Body.Close()
	copyResponseHeaders(c.Writer.Header(), resp.Header)
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
	return ""
}

func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func parseOptionalPositiveInt(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func bodyReadError(err error) error {
	if errors.Is(err, pkgplugin.ErrBodyTooLarge) {
		return apperror.Wrap(apperror.CodePayloadTooLarge, apperror.MessageRequestBodyTooLarge, err)
	}
	return apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err)
}

func gatewayErrorName(err error) string {
	appErr := apperror.From(err)
	switch appErr.Code {
	case apperror.CodeNotFound:
		return "route_not_found"
	case apperror.CodeServiceUnavailable:
		return "plugin_unavailable"
	default:
		return "resolve"
	}
}

func (h *PluginHandler) logGateway(metric domainplugin.GatewayMetric) {
	logger := h.logger
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("plugin gateway request",
		zap.String("plugin_key", metric.PluginKey),
		zap.String("instance_id", metric.InstanceID),
		zap.String("route_id", metric.RouteID),
		zap.String("gateway_path", metric.GatewayPath),
		zap.Int("upstream_status", metric.UpstreamStatus),
		zap.Int64("duration_ms", metric.Duration.Milliseconds()),
		zap.String("gateway_error", metric.GatewayError),
		zap.String("trace_id", metric.TraceID),
	)
}

func copyPluginHeaders(dst http.Header, src http.Header, forwardAuth bool) {
	for key, values := range src {
		if skipRequestHeader(key, forwardAuth) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func copyResponseHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		if hopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func skipRequestHeader(key string, forwardAuth bool) bool {
	lower := strings.ToLower(key)
	if hopByHopHeader(key) {
		return true
	}
	if strings.HasPrefix(lower, "x-keiyaku-") || strings.HasPrefix(lower, "x-forwarded-") {
		return true
	}
	switch lower {
	case "host", "cookie":
		return true
	case "authorization":
		return !forwardAuth
	default:
		return false
	}
}

func hopByHopHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func forwardedProto(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

func hasRole(claims port.TokenClaims, role string) bool {
	for _, item := range claims.Roles {
		if item == role {
			return true
		}
	}
	return false
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
