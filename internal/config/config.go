package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BotToken           string
	DatabaseURL        string
	RedisURL           string
	WebhookBaseURL     string
	WebhookSecret      string
	BotMode            string
	APIAddr            string
	AdminTelegramIDs   map[int64]struct{}
	MentorTelegramIDs  map[int64]struct{}
	AllowedTelegramIDs map[int64]struct{}
	PollTimeout        time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		BotToken:           strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:           strings.TrimSpace(os.Getenv("REDIS_URL")),
		WebhookBaseURL:     strings.TrimSpace(os.Getenv("WEBHOOK_BASE_URL")),
		WebhookSecret:      strings.TrimSpace(os.Getenv("WEBHOOK_SECRET")),
		BotMode:            defaultString(os.Getenv("BOT_MODE"), "polling"),
		APIAddr:            defaultString(os.Getenv("API_ADDR"), ":8080"),
		AdminTelegramIDs:   map[int64]struct{}{},
		MentorTelegramIDs:  map[int64]struct{}{},
		AllowedTelegramIDs: map[int64]struct{}{},
		PollTimeout:        60 * time.Second,
	}

	if timeout := strings.TrimSpace(os.Getenv("POLL_TIMEOUT_SECONDS")); timeout != "" {
		seconds, err := strconv.Atoi(timeout)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POLL_TIMEOUT_SECONDS: %w", err)
		}
		cfg.PollTimeout = time.Duration(seconds) * time.Second
	}

	if admins := strings.TrimSpace(os.Getenv("ADMIN_TELEGRAM_IDS")); admins != "" {
		for _, raw := range strings.Split(admins, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return Config{}, fmt.Errorf("invalid admin id %q: %w", raw, err)
			}
			cfg.AdminTelegramIDs[id] = struct{}{}
		}
	}
	if mentors := strings.TrimSpace(os.Getenv("MENTOR_TELEGRAM_IDS")); mentors != "" {
		for _, raw := range strings.Split(mentors, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return Config{}, fmt.Errorf("invalid mentor id %q: %w", raw, err)
			}
			cfg.MentorTelegramIDs[id] = struct{}{}
		}
	}
	if allowed := strings.TrimSpace(os.Getenv("ALLOWED_TELEGRAM_IDS")); allowed != "" {
		for _, raw := range strings.Split(allowed, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return Config{}, fmt.Errorf("invalid allowed id %q: %w", raw, err)
			}
			cfg.AllowedTelegramIDs[id] = struct{}{}
		}
	}
	for adminID := range cfg.AdminTelegramIDs {
		cfg.AllowedTelegramIDs[adminID] = struct{}{}
	}
	for mentorID := range cfg.MentorTelegramIDs {
		cfg.AllowedTelegramIDs[mentorID] = struct{}{}
	}

	if cfg.BotToken == "" {
		return Config{}, fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}
	return cfg, nil
}

func defaultString(v, fallback string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
