package router

import (
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
	"go.uber.org/zap"
)

type Options struct {
	RateLimit      RateLimitOptions
	CircuitBreaker CircuitBreakerOptions
	APIDocs        APIDocsOptions
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

	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	ArticleHandler *handler.ArticleHandler
	PluginHandler  *handler.PluginHandler
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
	apidocs.Inject(engine, apidocs.Options{
		Disabled: options.APIDocs.Disabled,
		Path:     options.APIDocs.Path,
		SpecPath: options.APIDocs.SpecPath,
		Title:    options.APIDocs.Title,
		Spec:     apispec.OpenAPIYAML(),
	})

	v1 := engine.Group("/api/v1")
	{
		v1.POST("/auth/register", deps.AuthHandler.Register)
		v1.POST("/auth/login", deps.AuthHandler.Login)
		v1.GET("/articles", deps.ArticleHandler.List)
		v1.GET("/articles/:id", deps.ArticleHandler.Get)
		if deps.PluginHandler != nil {
			v1.POST("/plugins/registrations", deps.PluginHandler.Register)
			v1.POST("/plugins/:plugin_key/instances/:instance_id/heartbeat", deps.PluginHandler.Heartbeat)
			v1.DELETE("/plugins/:plugin_key/instances/:instance_id", deps.PluginHandler.Unregister)
			v1.Any("/extensions/:plugin_key/*proxy_path", deps.PluginHandler.Gateway)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(deps.Tokens), middleware.Casbin(deps.Authorizer))
		protected.GET("/users/me", deps.UserHandler.Me)
		protected.POST("/articles", deps.ArticleHandler.Create)
		if deps.PluginHandler != nil {
			protected.GET("/plugins", deps.PluginHandler.List)
			protected.GET("/plugins/:plugin_key", deps.PluginHandler.Get)
		}
	}

	engine.NoRoute(func(c *gin.Context) {
		response.JSON(c, http.StatusNotFound, apperror.CodeNotFound, apperror.MessageRouteNotFound, nil)
	})
	return engine
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
	return options
}
