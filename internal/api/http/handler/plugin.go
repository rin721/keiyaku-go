package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
)

type PluginHandler struct {
	service    *appplugin.Service
	tokens     port.TokenIssuer
	authorizer port.Authorizer
	client     *http.Client
}

func NewPluginHandler(service *appplugin.Service, tokens port.TokenIssuer, authorizer port.Authorizer) *PluginHandler {
	return &PluginHandler{service: service, tokens: tokens, authorizer: authorizer, client: &http.Client{}}
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
// @Failure 401 {object} response.Body "Invalid plugin registration token"
// @Failure 403 {object} response.Body "Plugin key not allowed"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/registrations [post]
func (h *PluginHandler) Register(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	var req dto.PluginRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
		return
	}
	result, err := h.service.Register(c.Request.Context(), req.ToCommand(bearerToken(c)))
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
// @Failure 401 {object} response.Body "Invalid plugin registration token"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id}/heartbeat [post]
func (h *PluginHandler) Heartbeat(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	result, err := h.service.Heartbeat(c.Request.Context(), bearerToken(c), c.Param("plugin_key"), c.Param("instance_id"))
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
// @Failure 401 {object} response.Body "Invalid plugin registration token"
// @Failure 404 {object} response.Body "Plugin instance not found"
// @Failure 500 {object} response.Body "Internal server error"
// @Router /plugins/{plugin_key}/instances/{instance_id} [delete]
func (h *PluginHandler) Unregister(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	if err := h.service.Unregister(c.Request.Context(), bearerToken(c), c.Param("plugin_key"), c.Param("instance_id")); err != nil {
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

func (h *PluginHandler) Gateway(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, apperror.New(apperror.CodeInternal, apperror.MessagePluginHandlerNotReady))
		return
	}
	proxyPath := c.Param("proxy_path")
	if proxyPath == "" {
		proxyPath = "/"
	}
	resolved, err := h.service.ResolveRoute(c.Request.Context(), appplugin.ResolveRouteQuery{
		PluginKey: c.Param("plugin_key"),
		Method:    c.Request.Method,
		Path:      proxyPath,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	claims, ok, err := h.authorizeGateway(c, resolved.Route)
	if err != nil {
		response.Error(c, err)
		return
	}
	upstreamURL, err := domainplugin.BuildUpstreamURL(resolved.Instance.BaseURL, resolved.Route.UpstreamPath, resolved.Suffix, c.Request.URL.RawQuery)
	if err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return
	}
	h.forward(c, upstreamURL, resolved, claims, ok)
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

func (h *PluginHandler) forward(c *gin.Context, upstreamURL string, resolved *domainplugin.ResolvedRoute, claims port.TokenClaims, hasClaims bool) {
	timeout := resolved.Route.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, c.Request.Method, upstreamURL, c.Request.Body)
	if err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return
	}
	copyPluginHeaders(req.Header, c.Request.Header, resolved.Route.ForwardAuthHeader)
	traceID := trace.IDFromContext(c.Request.Context())
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
	if secret := h.service.GatewaySigningSecret(); secret != "" {
		addGatewaySignature(req, secret, traceID)
	}
	client := h.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		if isTimeout(err) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			response.Error(c, apperror.Wrap(apperror.CodeGatewayTimeout, apperror.MessagePluginUpstreamTimeout, err))
			return
		}
		response.Error(c, apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err))
		return
	}
	defer resp.Body.Close()
	copyResponseHeaders(c.Writer.Header(), resp.Header)
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
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

func addGatewaySignature(req *http.Request, secret string, traceID string) {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	req.Header.Set("X-Keiyaku-Timestamp", timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(req.Method))
	mac.Write([]byte("\n"))
	mac.Write([]byte(req.URL.Path))
	mac.Write([]byte("\n"))
	mac.Write([]byte(traceID))
	mac.Write([]byte("\n"))
	mac.Write([]byte(timestamp))
	req.Header.Set("X-Keiyaku-Signature", hex.EncodeToString(mac.Sum(nil)))
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
