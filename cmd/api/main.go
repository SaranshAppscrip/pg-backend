package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/database"
	"github.com/nivas/server/internal/repository/postgres"
	"github.com/nivas/server/internal/router"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/logger"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.Database, log)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repos := postgres.NewStoreBundle(pool)
	tokens := auth.NewTokenService(cfg.JWT)

	authSvc := service.NewAuthService(repos, tokens)
	deps := router.Deps{
		Config:   cfg,
		Log:      log,
		Tokens:   tokens,
		Auth:     authSvc,
		Settings: service.NewSettingsService(repos.Settings),
		Rooms:    service.NewRoomService(repos),
		Tenants:  service.NewTenantService(repos),
		Payments: service.NewPaymentService(repos.Payments),
		Expenses: service.NewExpenseService(repos.Expenses),
		Kitchen:  service.NewKitchenService(repos.Kitchen),
		Staff:    service.NewStaffService(repos.Staff, authSvc),
	}

	engine := router.New(deps)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("server starting", "addr", cfg.Addr(), "env", cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
