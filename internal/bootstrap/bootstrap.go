package bootstrap

import (
	"context"
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"sre-learning-bot/internal/app"
	"sre-learning-bot/internal/config"
	"sre-learning-bot/internal/platform"
)

type App struct {
	Config config.Config
	Store  *app.Store
	BotAPI *tgbotapi.BotAPI
	Close  func()
}

func Init(ctx context.Context, logger *slog.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	pool, err := platform.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	redisClient, err := platform.NewRedisClient(cfg.RedisURL)
	if err != nil {
		pool.Close()
		return nil, err
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		pool.Close()
		_ = redisClient.Close()
		return nil, err
	}

	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		pool.Close()
		_ = redisClient.Close()
		return nil, err
	}
	botAPI.Debug = os.Getenv("BOT_DEBUG") == "1"
	logger.Info("telegram authorized", "username", botAPI.Self.UserName)

	return &App{
		Config: cfg,
		Store:  app.NewStore(pool, redisClient),
		BotAPI: botAPI,
		Close: func() {
			pool.Close()
			_ = redisClient.Close()
		},
	}, nil
}
