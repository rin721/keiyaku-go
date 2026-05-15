package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/rin721/keiyaku-go/internal/bootstrap"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "", "path to config yaml")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(ctx, *configPath)
	if err != nil {
		panic(err)
	}

	errCh := make(chan error, 1)
	go func() {
		app.Logger.Info("http server starting", zap.String("addr", app.Server.Addr))
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		app.Logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			app.Logger.Error("http server failed", zap.Error(err))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout())
	defer cancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		panic(err)
	}
}
