package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/logger"
	"github.com/Waycoolers/fmlbot/services/auth/internal/clients/api"
	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
	"github.com/Waycoolers/fmlbot/services/auth/internal/handlers"
	"github.com/Waycoolers/fmlbot/services/auth/internal/server"
	"github.com/Waycoolers/fmlbot/services/auth/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../../.env")

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.Loglevel)

	store, err := storage.New(cfg.DB)
	if err != nil {
		slog.Error("Error connecting to database", "error", err)
		os.Exit(1)
	}

	err = store.Migrate()
	if err != nil {
		slog.Error("Error migrating database", "error", err)
		os.Exit(1)
	}

	c := api.New(cfg.API, cfg.InternalSecret)

	h, err := handlers.New(store.Tokens, cfg, c)
	if err != nil {
		slog.Error("Error creating handler", "error", err)
		os.Exit(1)
	}

	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	go cleanupExpiredTokens(cleanupCtx, store.Tokens)

	srv := server.New(cfg.Server, h, cfg.InternalSecret)
	srv.Start()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	slog.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Stop(shutdownCtx)
	if err != nil {
		slog.Error("Shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}

func cleanupExpiredTokens(ctx context.Context, tokens domain.TokensRepo) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := tokens.DeleteExpired(ctx)
			if err != nil {
				slog.Error("Error deleting expired refresh tokens", "error", err)
				continue
			}
			slog.Info("Expired refresh tokens deleted", "count", deleted)
		}
	}
}
