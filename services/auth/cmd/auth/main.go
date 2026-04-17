package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Waycoolers/fmlbot/common/logger"
	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
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

	h, err := handlers.New(store.Tokens, cfg)
	if err != nil {
		slog.Error("Error creating handler", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg.Server, h)
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
