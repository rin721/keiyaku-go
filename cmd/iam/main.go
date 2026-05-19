package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/keiyaku-go/internal/api/http/dto"
	"github.com/rin721/keiyaku-go/internal/api/http/handler"
	httpi18n "github.com/rin721/keiyaku-go/internal/api/http/i18n"
	"github.com/rin721/keiyaku-go/internal/api/http/middleware"
	"github.com/rin721/keiyaku-go/internal/api/http/response"
	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/auth"
	"github.com/rin721/keiyaku-go/internal/application/port"
	appuser "github.com/rin721/keiyaku-go/internal/application/user"
	domainiam "github.com/rin721/keiyaku-go/internal/domain/iam"
	domainuser "github.com/rin721/keiyaku-go/internal/domain/user"
	authcasbin "github.com/rin721/keiyaku-go/internal/infrastructure/auth/casbin"
	authjwt "github.com/rin721/keiyaku-go/internal/infrastructure/auth/jwt"
	"github.com/rin721/keiyaku-go/internal/infrastructure/auth/password"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	idsnowflake "github.com/rin721/keiyaku-go/internal/infrastructure/id/snowflake"
	zaplogger "github.com/rin721/keiyaku-go/internal/infrastructure/logger/zap"
	"github.com/rin721/keiyaku-go/internal/infrastructure/persistence/mysql"
	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
	"go.uber.org/zap"
)

const (
	appName    cmdcli.AppName  = "keiyaku-iam"
	flagConfig cmdcli.FlagName = "config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cmdcli.RunAndExit(ctx, newAppSpec(), os.Args)
}

func newAppSpec() cmdcli.AppSpec {
	return cmdcli.AppSpec{
		Name:                   appName,
		Usage:                  "Start Keiyaku-Go IAM service",
		UsageText:              "keiyaku-iam [global options]",
		UseShortOptionHandling: true,
		Flags: []cmdcli.Flag{
			cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagConfig, Aliases: []string{"c"}, Usage: "Config file path"}),
		},
		Action: runServer,
	}
}

func runServer(ctx context.Context, cliCtx *cmdcli.Context) error {
	cfg, err := config.Load(cliCtx.String(flagConfig))
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "load config", err)
	}
	if err := httpi18n.Init(cfg.I18N.Default, cfg.I18N.Supported, cfg.I18N.Files); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "init i18n", err)
	}
	logBundle, syncLogger, err := zaplogger.New(cfg.Log)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "init logger", err)
	}
	defer func() { _ = syncLogger() }()
	db, err := mysql.Open(cfg.MySQL, logBundle.Logger)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "open mysql", err)
	}
	defer func() { _ = mysql.Close(db) }()
	ids, err := idsnowflake.New(cfg.Snowflake.Node)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "init ids", err)
	}
	tokenService := authjwt.NewService(cfg.JWT)
	sessionRepo := mysql.NewIAMRefreshSessionRepository(db)
	sessionTokens := newSessionTokenIssuer(tokenService, sessionRepo, ids)
	hasher := password.NewBcryptHasher(cfg.Security.BcryptCost)
	authorizer, err := authcasbin.NewAuthorizer(cfg.RBAC)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "init authorizer", err)
	}
	userRepo := mysql.NewUserRepository(db)
	authService := auth.NewService(userRepo, ids, hasher, sessionTokens)
	userService := appuser.NewService(userRepo)

	engine := newRouter(routerDeps{
		serviceToken: cfg.IAM.ServiceToken,
		tokens:       sessionTokens,
		refresh:      sessionTokens,
		sessions:     sessionRepo,
		authorizer:   authorizer,
		authHandler:  handler.NewAuthHandler(authService),
		userHandler:  handler.NewUserHandler(userService),
		userRepo:     userRepo,
	})
	server := &http.Server{
		Addr:         cfg.IAM.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	errCh := make(chan error, 1)
	go func() {
		logBundle.Logger.Info("iam server starting", zap.String("addr", cfg.IAM.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "run iam server", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "shutdown iam server", err)
	}
	return nil
}

type routerDeps struct {
	serviceToken string
	tokens       port.TokenIssuer
	refresh      *sessionTokenIssuer
	sessions     *mysql.IAMRefreshSessionRepository
	authorizer   interface {
		Allow(role string, object string, action string) (bool, error)
	}
	authHandler *handler.AuthHandler
	userHandler *handler.UserHandler
	userRepo    interface {
		FindByID(ctx context.Context, id int64) (*domainuser.User, error)
	}
}

func newRouter(deps routerDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(middleware.I18N(), gin.Recovery())
	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	v1 := engine.Group("/api/v1")
	v1.POST("/auth/register", deps.authHandler.Register)
	v1.POST("/auth/login", deps.authHandler.Login)
	v1.POST("/auth/refresh", refresh(deps.refresh, deps.userRepo))
	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.tokens))
	protected.POST("/auth/logout", logout(deps.sessions))
	protected.GET("/users/me", deps.userHandler.Me)

	internal := engine.Group("/internal/v1")
	internal.Use(internalAuth(deps.serviceToken))
	internal.POST("/tokens/introspect", introspect(deps.tokens, deps.userRepo))
	internal.POST("/authorize", authorize(deps.authorizer))
	return engine
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func refresh(tokens *sessionTokenIssuer, users interface {
	FindByID(ctx context.Context, id int64) (*domainuser.User, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req refreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
			return
		}
		claims, err := tokens.ParseRefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil || claims.TokenID == "" {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken))
			return
		}
		entity, err := users.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Error(c, err)
			return
		}
		if err := entity.EnsureActive(); err != nil {
			response.Error(c, err)
			return
		}
		token, err := tokens.RotateRefreshToken(c.Request.Context(), port.TokenUser{ID: entity.ID, Username: entity.Username, Roles: entity.Roles}, claims.TokenID)
		if err != nil {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken))
			return
		}
		response.OK(c, dto.AuthResponse{User: dto.NewUserResponse(entity), AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, ExpiresAt: token.ExpiresAt})
	}
}

func logout(sessions *mysql.IAMRefreshSessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := middleware.Claims(c)
		if !ok {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken))
			return
		}
		if sessions != nil {
			if err := sessions.RevokeActiveRefreshSessionsByUser(c.Request.Context(), claims.UserID, time.Now().UTC()); err != nil {
				response.Error(c, apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err))
				return
			}
		}
		response.NoContent(c)
	}
}

type introspectRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

func introspect(tokens port.TokenIssuer, users interface {
	FindByID(ctx context.Context, id int64) (*domainuser.User, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req introspectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
			return
		}
		claims, err := tokens.ParseAccessToken(c.Request.Context(), req.AccessToken)
		if err != nil {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidAccessToken))
			return
		}
		entity, err := users.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Error(c, err)
			return
		}
		if err := entity.EnsureActive(); err != nil {
			response.Error(c, err)
			return
		}
		response.OK(c, gin.H{
			"user_id":    claims.UserID,
			"username":   claims.Username,
			"roles":      claims.Roles,
			"expires_at": claims.ExpiresAt,
		})
	}
}

type sessionTokenIssuer struct {
	tokens   *authjwt.Service
	sessions *mysql.IAMRefreshSessionRepository
	ids      port.IDGenerator
	now      func() time.Time
}

func newSessionTokenIssuer(tokens *authjwt.Service, sessions *mysql.IAMRefreshSessionRepository, ids port.IDGenerator) *sessionTokenIssuer {
	return &sessionTokenIssuer{
		tokens:   tokens,
		sessions: sessions,
		ids:      ids,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *sessionTokenIssuer) IssueToken(ctx context.Context, subject port.TokenUser) (port.TokenPair, error) {
	pair, session, err := s.newTokenPair(ctx, subject)
	if err != nil {
		return port.TokenPair{}, err
	}
	if err := s.sessions.CreateRefreshSession(ctx, session); err != nil {
		return port.TokenPair{}, err
	}
	return pair, nil
}

func (s *sessionTokenIssuer) ParseAccessToken(ctx context.Context, raw string) (port.TokenClaims, error) {
	return s.tokens.ParseAccessToken(ctx, raw)
}

func (s *sessionTokenIssuer) ParseRefreshToken(ctx context.Context, raw string) (port.TokenClaims, error) {
	return s.tokens.ParseRefreshToken(ctx, raw)
}

func (s *sessionTokenIssuer) RotateRefreshToken(ctx context.Context, subject port.TokenUser, currentTokenID string) (port.TokenPair, error) {
	pair, session, err := s.newTokenPair(ctx, subject)
	if err != nil {
		return port.TokenPair{}, err
	}
	if err := s.sessions.RotateRefreshSession(ctx, currentTokenID, subject.ID, session, s.now()); err != nil {
		return port.TokenPair{}, err
	}
	return pair, nil
}

func (s *sessionTokenIssuer) newTokenPair(ctx context.Context, subject port.TokenUser) (port.TokenPair, *domainiam.RefreshSession, error) {
	if s == nil || s.tokens == nil || s.sessions == nil || s.ids == nil {
		return port.TokenPair{}, nil, errors.New("iam session token issuer is not ready")
	}
	sessionID, err := s.ids.NewID(ctx)
	if err != nil {
		return port.TokenPair{}, nil, err
	}
	refreshTokenID, err := s.ids.NewID(ctx)
	if err != nil {
		return port.TokenPair{}, nil, err
	}
	refreshTokenIDText := strconv.FormatInt(refreshTokenID, 10)
	pair, err := s.tokens.IssueTokenWithRefreshID(ctx, subject, refreshTokenIDText)
	if err != nil {
		return port.TokenPair{}, nil, err
	}
	session, err := domainiam.NewRefreshSession(sessionID, subject.ID, refreshTokenIDText, pair.RefreshExpiresAt, s.now())
	if err != nil {
		return port.TokenPair{}, nil, err
	}
	return pair, session, nil
}

type authorizeRequest struct {
	Role   string `json:"role" binding:"required"`
	Object string `json:"object" binding:"required"`
	Action string `json:"action" binding:"required"`
}

func authorize(authorizer interface {
	Allow(role string, object string, action string) (bool, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authorizeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeInvalidArgument, apperror.MessageInvalidRequestBody, err))
			return
		}
		allowed, err := authorizer.Allow(req.Role, req.Object, req.Action)
		if err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeDependency, apperror.MessagePermissionCheckFail, err))
			return
		}
		response.OK(c, gin.H{"allowed": allowed})
	}
}

func internalAuth(serviceToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if serviceToken == "" {
			c.Next()
			return
		}
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(serviceToken)) != 1 {
			response.Error(c, apperror.New(apperror.CodeUnauthorized, apperror.MessageUnauthorized))
			c.Abort()
			return
		}
		c.Next()
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
