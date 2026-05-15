package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rin721/keiyaku-go/internal/api/http/handler"
	httprouter "github.com/rin721/keiyaku-go/internal/api/http/router"
	apparticle "github.com/rin721/keiyaku-go/internal/application/article"
	"github.com/rin721/keiyaku-go/internal/application/auth"
	appuser "github.com/rin721/keiyaku-go/internal/application/user"
	authcasbin "github.com/rin721/keiyaku-go/internal/infrastructure/auth/casbin"
	authjwt "github.com/rin721/keiyaku-go/internal/infrastructure/auth/jwt"
	"github.com/rin721/keiyaku-go/internal/infrastructure/auth/password"
	rediscache "github.com/rin721/keiyaku-go/internal/infrastructure/cache/redis"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	idsnowflake "github.com/rin721/keiyaku-go/internal/infrastructure/id/snowflake"
	zaplogger "github.com/rin721/keiyaku-go/internal/infrastructure/logger/zap"
	"github.com/rin721/keiyaku-go/internal/infrastructure/persistence/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger
	Server *http.Server
	DB     *gorm.DB
	Redis  *redis.Client

	syncLogger func() error
}

func New(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
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
	enforcer, err := authcasbin.NewEnforcer()
	if err != nil {
		_ = redisClient.Close()
		_ = mysql.Close(db)
		_ = syncLogger()
		return nil, err
	}

	userRepo := mysql.NewUserRepository(db)
	articleRepo := mysql.NewArticleRepository(db)
	authService := auth.NewService(userRepo, idGenerator, hasher, tokenService)
	userService := appuser.NewService(userRepo)
	articleService := apparticle.NewService(articleRepo, idGenerator)

	router := httprouter.New(httprouter.Deps{
		Config:         cfg,
		Logger:         logger,
		Tokens:         tokenService,
		Enforcer:       enforcer,
		AuthHandler:    handler.NewAuthHandler(authService),
		UserHandler:    handler.NewUserHandler(userService),
		ArticleHandler: handler.NewArticleHandler(articleService),
	})
	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	return &App{
		Config:     cfg,
		Logger:     logger,
		Server:     server,
		DB:         db,
		Redis:      redisClient,
		syncLogger: syncLogger,
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
