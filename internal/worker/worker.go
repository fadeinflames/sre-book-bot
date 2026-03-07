package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sre-learning-bot/internal/app"
	"sre-learning-bot/internal/bot"
)

type Service struct {
	logger *slog.Logger
	store  *app.Store
	botSvc *bot.Service
}

func New(logger *slog.Logger, store *app.Store, botSvc *bot.Service) *Service {
	return &Service{
		logger: logger,
		store:  store,
		botSvc: botSvc,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("worker started")
	for {
		if err := s.runIteration(ctx); err != nil {
			s.logger.Error("worker iteration failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Service) runIteration(ctx context.Context) error {
	counts, err := s.store.DueReviewCounts(ctx)
	if err != nil {
		return err
	}
	dayKey := time.Now().UTC().Format("2006-01-02")

	for telegramID, count := range counts {
		ok, err := s.store.ShouldSendReminder(ctx, telegramID, dayKey)
		if err != nil {
			s.logger.Error("redis check failed", "err", err, "telegram_id", telegramID)
			continue
		}
		if !ok {
			continue
		}
		if err := s.botSvc.SendReminder(ctx, telegramID, count); err != nil {
			s.logger.Error("send reminder failed", "err", err, "telegram_id", telegramID)
			continue
		}
		s.logger.Info("sent reminder", "telegram_id", telegramID, "count", count)
	}

	return nil
}

func FormatWorkerHealth() string {
	return fmt.Sprintf("worker_ok %d\n", time.Now().UTC().Unix())
}
