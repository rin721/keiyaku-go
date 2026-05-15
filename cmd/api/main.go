package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rin721/keiyaku-go/internal/bootstrap"
	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
	"go.uber.org/zap"
)

const (
	appName    cmdcli.AppName  = "keiyaku-api"
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
		Usage:                  "启动 Keiyaku-Go HTTP API 服务",
		UsageText:              "keiyaku-api [global options]",
		Description:            "读取配置、装配依赖并启动 Gin HTTP Server。",
		UseShortOptionHandling: true,
		Flags: []cmdcli.Flag{
			cmdcli.StringFlag(cmdcli.StringFlagSpec{
				Name:    flagConfig,
				Aliases: []string{"c"},
				Usage:   "配置文件路径",
			}),
		},
		Action: runServer,
	}
}

func runServer(ctx context.Context, cliCtx *cmdcli.Context) error {
	ui := cliCtx.UI()
	configPath := cliCtx.String(flagConfig)

	app, err := bootstrap.New(ctx, configPath)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "初始化应用失败", err)
	}

	errCh := make(chan error, 1)
	go func() {
		app.Logger.Info("http server starting", zap.String("addr", app.Server.Addr))
		ui.Infof("HTTP 服务启动中：%s", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		app.Logger.Info("shutdown signal received")
		ui.Info("收到退出信号，准备关闭服务")
	case err := <-errCh:
		if err != nil {
			app.Logger.Error("http server failed", zap.Error(err))
			return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "HTTP 服务运行失败", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout())
	defer cancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "关闭应用失败", err)
	}
	ui.Success("HTTP 服务已关闭")
	return nil
}
