package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/database"
	"github.com/nivas/server/internal/notification"
	"github.com/nivas/server/internal/repository/postgres"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/logger"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{
		Level:  envOr("LOG_LEVEL", "info"),
		Format: envOr("LOG_FORMAT", "json"),
	})
	slog.SetDefault(log)

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.Database, log)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repos := postgres.NewStoreBundle(pool)
	emailSender := notification.NewEmailSender(cfg.Email, cfg.App.Env, log)
	reminderSvc := service.NewReminderService(repos.Reminders, emailSender, cfg.Reminder)

	c := cron.New()
	_, err = c.AddFunc(cfg.Reminder.CronSpec, func() {
		jobCtx := logger.WithContext(context.Background(), log.With("job", "rent_reminders"))
		if err := reminderSvc.Run(jobCtx); err != nil {
			log.Error("rent reminder job failed", "error", err)
		}
	})
	if err != nil {
		log.Error("schedule cron", "error", err)
		os.Exit(1)
	}

	c.Start()
	log.Info("cron started", "spec", cfg.Reminder.CronSpec)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	c.Stop()
	log.Info("cron stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
