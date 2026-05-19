package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	apispec "github.com/rin721/keiyaku-go/api"
	"github.com/rin721/keiyaku-go/internal/api/http/apidocs"
	"github.com/rin721/keiyaku-go/internal/api/http/handler"
	"github.com/rin721/keiyaku-go/internal/api/http/middleware"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
	"go.uber.org/zap"
)

type Options struct {
	RateLimit          RateLimitOptions
	CircuitBreaker     CircuitBreakerOptions
	APIDocs            APIDocsOptions
	PluginPublicPrefix string
}

type RateLimitOptions struct {
	RequestsPerSecond float64
	Burst             int
}

type CircuitBreakerOptions struct {
	Name             string
	FailureThreshold uint32
	OpenTimeout      time.Duration
}

type APIDocsOptions struct {
	Disabled bool
	Path     string
	SpecPath string
	Title    string
}

type Deps struct {
	Options    Options
	Logger     *zap.Logger
	Tokens     port.TokenIssuer
	Authorizer port.Authorizer
	Readiness  func(context.Context) error

	PluginHandler *handler.PluginHandler
}

func New(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	options := normalizeOptions(deps.Options)
	engine := gin.New()
	engine.Use(
		middleware.I18N(),
		middleware.TraceID(),
		middleware.Recovery(deps.Logger),
		middleware.Logging(deps.Logger),
		cors.Default(),
		middleware.RateLimit(options.RateLimit.RequestsPerSecond, options.RateLimit.Burst),
		middleware.CircuitBreaker(options.CircuitBreaker.Name, options.CircuitBreaker.FailureThreshold, options.CircuitBreaker.OpenTimeout),
	)
	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if deps.Readiness != nil {
			if err := deps.Readiness(c.Request.Context()); err != nil {
				response.Error(c, apperror.Wrap(apperror.CodeServiceUnavailable, apperror.MessageServiceUnavailable, err))
				return
			}
		}
		response.OK(c, gin.H{"status": "ready"})
	})
	apidocs.Inject(engine, apidocs.Options{
		Disabled: options.APIDocs.Disabled,
		Path:     options.APIDocs.Path,
		SpecPath: options.APIDocs.SpecPath,
		Title:    options.APIDocs.Title,
		Spec:     apispec.OpenAPIYAML(),
	})

	if deps.PluginHandler != nil {
		registerPluginGateway(engine, options.PluginPublicPrefix, deps.PluginHandler)
	}

	v1 := engine.Group("/api/v1")
	{
		if deps.PluginHandler != nil {
			v1.POST("/plugins/registrations", deps.PluginHandler.Register)
			v1.POST("/plugins/:plugin_key/instances/:instance_id/heartbeat", deps.PluginHandler.Heartbeat)
			v1.DELETE("/plugins/:plugin_key/instances/:instance_id", deps.PluginHandler.Unregister)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(deps.Tokens), middleware.Casbin(deps.Authorizer))
		if deps.PluginHandler != nil {
			protected.GET("/plugins", deps.PluginHandler.List)
			protected.GET("/plugins/:plugin_key", deps.PluginHandler.Get)
			protected.GET("/plugins/:plugin_key/diagnostics", deps.PluginHandler.Diagnostics)
			protected.GET("/plugins/:plugin_key/instances", deps.PluginHandler.ListInstances)
			protected.POST("/plugins/:plugin_key/disable", deps.PluginHandler.Disable)
			protected.POST("/plugins/:plugin_key/enable", deps.PluginHandler.Enable)
			protected.POST("/plugins/:plugin_key/instances/:instance_id/disable", deps.PluginHandler.DisableInstance)
			protected.POST("/plugins/:plugin_key/instances/:instance_id/enable", deps.PluginHandler.EnableInstance)
			protected.GET("/plugins/:plugin_key/audit-events", deps.PluginHandler.ListAuditEvents)
		}
	}

	engine.NoRoute(func(c *gin.Context) {
		response.JSON(c, http.StatusNotFound, apperror.CodeNotFound, apperror.MessageRouteNotFound, nil)
	})
	return engine
}

func registerPluginGateway(engine *gin.Engine, publicPrefix string, handler *handler.PluginHandler) {
	prefix := pkgplugin.NormalizePublicPrefix(publicPrefix)
	if prefix == "/" {
		engine.Any("/*proxy_path", handler.Gateway)
		return
	}
	engine.Any(prefix+"/*proxy_path", handler.Gateway)
}

func normalizeOptions(options Options) Options {
	if options.RateLimit.RequestsPerSecond <= 0 {
		options.RateLimit.RequestsPerSecond = 100
	}
	if options.RateLimit.Burst <= 0 {
		options.RateLimit.Burst = int(options.RateLimit.RequestsPerSecond)
	}
	if options.CircuitBreaker.Name == "" {
		options.CircuitBreaker.Name = "http-api"
	}
	if options.CircuitBreaker.FailureThreshold == 0 {
		options.CircuitBreaker.FailureThreshold = 5
	}
	if options.CircuitBreaker.OpenTimeout <= 0 {
		options.CircuitBreaker.OpenTimeout = 5 * time.Second
	}
	options.PluginPublicPrefix = pkgplugin.NormalizePublicPrefix(options.PluginPublicPrefix)
	return options
}
