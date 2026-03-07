package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sre-learning-bot/internal/api"
	"sre-learning-bot/internal/bootstrap"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appState, err := bootstrap.Init(ctx, logger)
	if err != nil {
		logger.Error("bootstrap failed", "err", err)
		os.Exit(1)
	}
	defer appState.Close()

	srv := api.New(logger, appState.Config.APIAddr, appState.Store)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("api server failed", "err", err)
		os.Exit(1)
	}
}
