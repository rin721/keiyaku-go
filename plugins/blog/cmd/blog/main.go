package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
	"github.com/rin721/keiyaku-go/plugins/blog/internal/article"
	"github.com/rin721/keiyaku-go/plugins/blog/internal/platform"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := platform.LoadConfig()
	if err != nil {
		return err
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer func() { _ = logger.Sync() }()

	db, err := platform.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		return err
	}
	defer func() { _ = platform.CloseMySQL(db) }()

	ids, err := platform.NewSnowflakeGenerator(cfg.SnowflakeNode)
	if err != nil {
		return err
	}
	repo := article.NewMySQLRepository(db)
	service := article.NewService(repo, ids)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	platform.RegisterHealth(engine, db)
	article.NewHandler(service).RegisterRoutes(engine.Group("", platform.GatewaySignature("blog", cfg.GatewaySecret, pluginsdk.NewMemoryNonceStore())))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	manifest := blogManifest(cfg)
	if cfg.RegistrationSecret != "" {
		client := pluginsdk.NewClient(cfg.KeiyakuHost, manifest.PluginKey, cfg.RegistrationSecret)
		go func() {
			runner := pluginsdk.LifecycleRunner{
				Client:            client,
				Manifest:          manifest,
				HeartbeatInterval: cfg.HeartbeatInterval,
				RegisterTimeout:   cfg.RegisterTimeout,
				UnregisterTimeout: cfg.ShutdownTimeout,
				OnError:           func(err error) { logger.Warn("blog plugin lifecycle failed", zap.Error(err)) },
			}
			_ = runner.Run(ctx)
		}()
	}

	server := &http.Server{Addr: cfg.Addr, Handler: engine, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	logger.Info("blog plugin listening", zap.String("addr", cfg.Addr), zap.String("base_url", cfg.BaseURL))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func blogManifest(cfg platform.Config) pluginsdk.Manifest {
	return pluginsdk.Manifest{
		SchemaVersion: pluginsdk.DefaultSchemaVersion,
		PluginKey:     "blog",
		Name:          "Blog",
		Version:       "0.1.0",
		InstanceID:    cfg.InstanceID,
		Protocol:      pluginsdk.ProtocolHTTP,
		BaseURL:       cfg.BaseURL,
		HealthPath:    "/healthz",
		Metadata: map[string]string{
			"domain":  "blog",
			"service": "blog",
		},
		Routes: []pluginsdk.Route{
			{
				RouteID:           "articles-create",
				Method:            pluginsdk.MethodPost,
				MatchType:         pluginsdk.MatchTypeExact,
				GatewayPath:       "/api/v1/extensions/blog/articles",
				UpstreamPath:      "/articles",
				AuthPolicy:        pluginsdk.AuthPolicyRBAC,
				Timeout:           "5s",
				ForwardAuthHeader: false,
			},
			{
				RouteID:           "articles-list",
				Method:            pluginsdk.MethodGet,
				MatchType:         pluginsdk.MatchTypeExact,
				GatewayPath:       "/api/v1/extensions/blog/articles",
				UpstreamPath:      "/articles",
				AuthPolicy:        pluginsdk.AuthPolicyRBAC,
				Timeout:           "5s",
				ForwardAuthHeader: false,
			},
			{
				RouteID:           "articles-detail",
				Method:            pluginsdk.MethodGet,
				MatchType:         pluginsdk.MatchTypePrefix,
				GatewayPath:       "/api/v1/extensions/blog/articles",
				UpstreamPath:      "/articles",
				AuthPolicy:        pluginsdk.AuthPolicyRBAC,
				Timeout:           "5s",
				ForwardAuthHeader: false,
			},
		},
	}
}
