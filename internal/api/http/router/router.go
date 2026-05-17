package router

import (
	"net/http"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/handler"
	"github.com/rin721/keiyaku-go/internal/api/http/middleware"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/port"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	"github.com/rin721/keiyaku-go/types"
	"go.uber.org/zap"
)

type Deps struct {
	Config   *config.Config
	Logger   *zap.Logger
	Tokens   port.TokenIssuer
	Enforcer *casbinv3.Enforcer

	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	ArticleHandler *handler.ArticleHandler
}

func New(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		middleware.I18N(),
		middleware.TraceID(),
		middleware.Recovery(deps.Logger),
		middleware.Logging(deps.Logger),
		cors.Default(),
		middleware.RateLimit(deps.Config.Security.RateLimit.RequestsPerSecond, deps.Config.Security.RateLimit.Burst),
		middleware.CircuitBreaker("http-api", deps.Config.Security.CircuitBreaker.FailureThreshold, deps.Config.Security.CircuitBreaker.OpenTimeout),
	)
	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")
	{
		v1.POST("/auth/register", deps.AuthHandler.Register)
		v1.POST("/auth/login", deps.AuthHandler.Login)
		v1.GET("/articles", deps.ArticleHandler.List)
		v1.GET("/articles/:id", deps.ArticleHandler.Get)

		protected := v1.Group("")
		protected.Use(middleware.Auth(deps.Tokens), middleware.Casbin(deps.Enforcer))
		protected.GET("/users/me", deps.UserHandler.Me)
		protected.POST("/articles", deps.ArticleHandler.Create)
	}

	engine.NoRoute(func(c *gin.Context) {
		response.JSON(c, http.StatusNotFound, types.CodeNotFound, types.MessageRouteNotFound, nil)
	})
	return engine
}
