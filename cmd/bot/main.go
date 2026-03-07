package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sre-learning-bot/internal/bootstrap"
	"sre-learning-bot/internal/bot"
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

	botSvc := bot.New(logger, appState.Config, appState.Store, appState.BotAPI)
	if err := botSvc.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("bot run failed", "err", err)
		os.Exit(1)
	}
}
