package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rin721/keiyaku-go/internal/api/http/handler"
	httpi18n "github.com/rin721/keiyaku-go/internal/api/http/i18n"
	httprouter "github.com/rin721/keiyaku-go/internal/api/http/router"
	apparticle "github.com/rin721/keiyaku-go/internal/application/article"
	"github.com/rin721/keiyaku-go/internal/application/auth"
	appplugin "github.com/rin721/keiyaku-go/internal/application/plugin"
	"github.com/rin721/keiyaku-go/internal/application/port"
	appuser "github.com/rin721/keiyaku-go/internal/application/user"
	authcasbin "github.com/rin721/keiyaku-go/internal/infrastructure/auth/casbin"
	authjwt "github.com/rin721/keiyaku-go/internal/infrastructure/auth/jwt"
	"github.com/rin721/keiyaku-go/internal/infrastructure/auth/password"
	rediscache "github.com/rin721/keiyaku-go/internal/infrastructure/cache/redis"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	idsnowflake "github.com/rin721/keiyaku-go/internal/infrastructure/id/snowflake"
	zaplogger "github.com/rin721/keiyaku-go/internal/infrastructure/logger/zap"
	"github.com/rin721/keiyaku-go/internal/infrastructure/persistence/mysql"
	infraplugin "github.com/rin721/keiyaku-go/internal/infrastructure/plugin"
	"github.com/rin721/keiyaku-go/internal/observability/metrics"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger
	Server *http.Server
	DB     *gorm.DB
	Redis  *redis.Client

	pluginHealthCancel context.CancelFunc
	syncLogger         func() error
}

func New(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err := httpi18n.Init(cfg.I18N.Default, cfg.I18N.Supported, cfg.I18N.Files); err != nil {
		return nil, err
	}
	logBundle, syncLogger, err := zaplogger.New(cfg.Log)
	if err != nil {
		return nil, err
	}
	logger := logBundle.Logger

	db, err := mysql.Open(cfg.MySQL, logger)
	if err != nil {
		_ = syncLogger()
		return nil, err
	}
	redisClient, err := rediscache.NewClient(ctx, cfg.Redis)
	if err != nil {
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}

	idGenerator, err := idsnowflake.New(cfg.Snowflake.Node)
	if err != nil {
		_ = redisClient.Close()
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}
	tokenService := authjwt.NewService(cfg.JWT)
	hasher := password.NewBcryptHasher(cfg.Security.BcryptCost)
	authorizer, err := authcasbin.NewAuthorizer(cfg.RBAC)
	if err != nil {
		_ = redisClient.Close()
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}

	userRepo := mysql.NewUserRepository(db)
	articleRepo := mysql.NewArticleRepository(db)
	pluginRepo := mysql.NewPluginRegistryRepository(db)
	authService := auth.NewService(userRepo, idGenerator, hasher, tokenService)
	userService := appuser.NewService(userRepo)
	articleService := apparticle.NewService(articleRepo, idGenerator)
	pluginService, err := appplugin.NewService(pluginRepo, appplugin.Config{
		Enabled:              cfg.Plugins.Enabled,
		RegistrationTokens:   cfg.Plugins.RegistrationTokens,
		AllowedPluginKeys:    cfg.Plugins.AllowedPluginKeys,
		PublicPrefix:         cfg.Plugins.PublicPrefix,
		HeartbeatTTL:         cfg.Plugins.HeartbeatTTL,
		RequestTimeout:       cfg.Plugins.RequestTimeout,
		HealthCheckInterval:  cfg.Plugins.HealthCheckInterval,
		HealthCheckTimeout:   cfg.Plugins.HealthCheckTimeout,
		UnhealthyThreshold:   cfg.Plugins.UnhealthyThreshold,
		RouteCacheTTL:        cfg.Plugins.RouteCacheTTL,
		AuditRetentionDays:   cfg.Plugins.AuditRetentionDays,
		MaxAuditQueryLimit:   cfg.Plugins.MaxAuditQueryLimit,
		AllowedHosts:         cfg.Plugins.AllowedHosts,
		AllowedCIDRs:         cfg.Plugins.AllowedCIDRs,
		AllowLoopback:        cfg.Plugins.AllowLoopback,
		AllowPublicRoutes:    cfg.Plugins.AllowPublicRoutes,
		GatewaySigningSecret: cfg.Plugins.GatewaySigningSecret,
	}, appplugin.WithAuditRepository(pluginRepo), appplugin.WithMetrics(metrics.NoopPluginMetrics{}))
	if err != nil {
		_ = redisClient.Close()
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}
	pluginProbe := infraplugin.NewHTTPHealthProbe(cfg.Plugins.HealthCheckTimeout)
	pluginHealthCancel := runPluginHealthChecks(ctx, logger, pluginService, pluginProbe, cfg.Plugins.HealthCheckInterval)

	router := httprouter.New(httprouter.Deps{
		Options: httprouter.Options{
			RateLimit: httprouter.RateLimitOptions{
				RequestsPerSecond: cfg.Security.RateLimit.RequestsPerSecond,
				Burst:             cfg.Security.RateLimit.Burst,
			},
			CircuitBreaker: httprouter.CircuitBreakerOptions{
				Name:             "http-api",
				FailureThreshold: cfg.Security.CircuitBreaker.FailureThreshold,
				OpenTimeout:      cfg.Security.CircuitBreaker.OpenTimeout,
			},
		},
		Logger:         logger,
		Tokens:         tokenService,
		Authorizer:     authorizer,
		AuthHandler:    handler.NewAuthHandler(authService),
		UserHandler:    handler.NewUserHandler(userService),
		ArticleHandler: handler.NewArticleHandler(articleService),
		PluginHandler:  handler.NewPluginHandler(pluginService, tokenService, authorizer, handler.WithPluginLogger(logger)),
	})
	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	return &App{
		Config:             cfg,
		Logger:             logger,
		Server:             server,
		DB:                 db,
		Redis:              redisClient,
		pluginHealthCancel: pluginHealthCancel,
		syncLogger:         syncLogger,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	var shutdownErr error
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("shutdown http server: %w", err)
		}
	}
	if a.pluginHealthCancel != nil {
		a.pluginHealthCancel()
	}
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil && shutdownErr == nil {
			shutdownErr = fmt.Errorf("close redis: %w", err)
		}
	}
	if err := mysql.Close(a.DB); err != nil && shutdownErr == nil {
		shutdownErr = err
	}
	if a.syncLogger != nil {
		if err := a.syncLogger(); err != nil && shutdownErr == nil {
			shutdownErr = fmt.Errorf("sync logger: %w", err)
		}
	}
	return shutdownErr
}

func (a *App) ShutdownTimeout() time.Duration {
	if a == nil || a.Config == nil || a.Config.Server.ShutdownTimeout <= 0 {
		return 10 * time.Second
	}
	return a.Config.Server.ShutdownTimeout
}

func runPluginHealthChecks(ctx context.Context, logger *zap.Logger, service *appplugin.Service, probe port.PluginHealthProbe, interval time.Duration) context.CancelFunc {
	if service == nil || probe == nil || interval <= 0 {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	healthCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-healthCtx.Done():
				return
			case <-ticker.C:
				if err := service.CheckHealth(healthCtx, probe); err != nil {
					logger.Warn("plugin health check failed", zap.Error(err))
				}
			}
		}
	}()
	return cancel
}
