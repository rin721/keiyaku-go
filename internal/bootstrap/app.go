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
	appplugin "github.com/rin721/keiyaku-go/internal/application/plugin"
	"github.com/rin721/keiyaku-go/internal/application/port"
	rediscache "github.com/rin721/keiyaku-go/internal/infrastructure/cache/redis"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	infraiam "github.com/rin721/keiyaku-go/internal/infrastructure/iam"
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

	pluginHealthCancel      context.CancelFunc
	pluginMaintenanceCancel context.CancelFunc
	syncLogger              func() error
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

	identityClient := infraiam.NewClient(cfg.IAM)
	pluginRepo := mysql.NewPluginRegistryRepository(db)
	pluginService, err := appplugin.NewService(pluginRepo, appplugin.Config{
		Enabled:                  cfg.Plugins.Enabled,
		TrustedPlugins:           appPluginTrust(cfg.Plugins.TrustedPlugins),
		PublicPrefix:             cfg.Plugins.PublicPrefix,
		HeartbeatTTL:             cfg.Plugins.HeartbeatTTL,
		RequestTimeout:           cfg.Plugins.RequestTimeout,
		MaxRegistrationBodyBytes: cfg.Plugins.MaxRegistrationBodyBytes,
		MaxGatewayBodyBytes:      cfg.Plugins.MaxGatewayBodyBytes,
		MaxRouteTimeout:          cfg.Plugins.MaxRouteTimeout,
		HealthCheckInterval:      cfg.Plugins.HealthCheckInterval,
		HealthCheckTimeout:       cfg.Plugins.HealthCheckTimeout,
		UnhealthyThreshold:       cfg.Plugins.UnhealthyThreshold,
		RouteCacheTTL:            cfg.Plugins.RouteCacheTTL,
		MaintenanceInterval:      cfg.Plugins.MaintenanceInterval,
		AuditRetentionDays:       cfg.Plugins.AuditRetentionDays,
		MaxAuditQueryLimit:       cfg.Plugins.MaxAuditQueryLimit,
		AllowPublicRoutes:        cfg.Plugins.AllowPublicRoutes,
	}, appplugin.WithAuditRepository(pluginRepo), appplugin.WithMetrics(metrics.NoopPluginMetrics{}))
	if err != nil {
		_ = redisClient.Close()
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}
	pluginProbe := infraplugin.NewHTTPHealthProbe(cfg.Plugins.HealthCheckTimeout)
	pluginHealthCancel := runPluginHealthChecks(ctx, logger, pluginService, pluginProbe, cfg.Plugins.HealthCheckInterval)
	pluginMaintenanceCancel := runPluginMaintenance(ctx, logger, pluginService, cfg.Plugins.MaintenanceInterval)

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
			PluginPublicPrefix: cfg.Plugins.PublicPrefix,
		},
		Logger:        logger,
		Tokens:        identityClient,
		Authorizer:    identityClient,
		Readiness:     readinessCheck(db, redisClient, identityClient),
		PluginHandler: handler.NewPluginHandler(pluginService, identityClient, identityClient, handler.WithPluginLogger(logger), handler.WithPluginHTTPClient(&http.Client{Timeout: cfg.Plugins.RequestTimeout})),
	})
	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	return &App{
		Config:                  cfg,
		Logger:                  logger,
		Server:                  server,
		DB:                      db,
		Redis:                   redisClient,
		pluginHealthCancel:      pluginHealthCancel,
		pluginMaintenanceCancel: pluginMaintenanceCancel,
		syncLogger:              syncLogger,
	}, nil
}

func readinessCheck(db *gorm.DB, redisClient *redis.Client, identityClient *infraiam.Client) func(context.Context) error {
	return func(ctx context.Context) error {
		if db == nil {
			return fmt.Errorf("mysql is not ready")
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("mysql handle: %w", err)
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("mysql ping: %w", err)
		}
		if redisClient == nil {
			return fmt.Errorf("redis is not ready")
		}
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis ping: %w", err)
		}
		if identityClient != nil {
			if err := identityClient.Health(ctx); err != nil {
				return fmt.Errorf("iam health: %w", err)
			}
		}
		return nil
	}
}

func appPluginTrust(input map[string]config.TrustedPluginConfig) map[string]appplugin.TrustedPluginConfig {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]appplugin.TrustedPluginConfig, len(input))
	for pluginKey, trust := range input {
		output[pluginKey] = appplugin.TrustedPluginConfig{
			RegistrationSecret: trust.RegistrationSecret,
			GatewaySecret:      trust.GatewaySecret,
			AllowedHosts:       append([]string(nil), trust.AllowedHosts...),
			AllowedCIDRs:       append([]string(nil), trust.AllowedCIDRs...),
			AllowLoopback:      trust.AllowLoopback,
		}
	}
	return output
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
	if a.pluginMaintenanceCancel != nil {
		a.pluginMaintenanceCancel()
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

func runPluginMaintenance(ctx context.Context, logger *zap.Logger, service *appplugin.Service, interval time.Duration) context.CancelFunc {
	if service == nil || interval <= 0 {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	maintenanceCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-maintenanceCtx.Done():
				return
			case <-ticker.C:
				result, err := service.Maintain(maintenanceCtx)
				if err != nil {
					logger.Warn("plugin maintenance failed", zap.Error(err))
					continue
				}
				if result != nil && (result.PrunedSignatureNonces > 0 || result.PrunedAuditEvents > 0 || result.DisabledStaleInstances > 0) {
					logger.Info("plugin maintenance completed",
						zap.Int64("pruned_signature_nonces", result.PrunedSignatureNonces),
						zap.Int64("pruned_audit_events", result.PrunedAuditEvents),
						zap.Int64("disabled_stale_instances", result.DisabledStaleInstances),
					)
				}
			}
		}
	}()
	return cancel
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
