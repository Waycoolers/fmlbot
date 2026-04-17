package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Waycoolers/fmlbot/common/logger"
	"github.com/Waycoolers/fmlbot/services/bot/internal/app"
	"github.com/Waycoolers/fmlbot/services/bot/internal/clients/api"
	"github.com/Waycoolers/fmlbot/services/bot/internal/clients/telegram"
	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/handlers"
	"github.com/Waycoolers/fmlbot/services/bot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/services/bot/internal/server"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load("../../.env")

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.Loglevel)

	rdb, err := redis_store.New(cfg.RDB)
	if err != nil {
		slog.Error("Redis store init failed", "error", err)
		os.Exit(1)
	}
	defer func(rdb *redis.Client) {
		er := rdb.Close()
		if er != nil {
			slog.Error("Redis close failed", "error", er)
		}
	}(rdb)

	tgClient := telegram.NewTelegramClient(cfg)
	menuUI := ui.New(tgClient)
	importantDateDrafts := redis_store.NewImportantDateDraftStore(rdb, 15*time.Minute)
	importantDateEditDrafts := redis_store.NewImportantDateEditDraftStore(rdb, 15*time.Minute)
	client := api.New(cfg)
	machine := state.New()
	handler := handlers.New(menuUI, importantDateDrafts, importantDateEditDrafts, client, machine)
	router := app.NewRouter(handler)

	b, err := app.New(tgClient, router)
	if err != nil {
		slog.Error("Error creating bot", "error", err)
		os.Exit(1)
	}

	s := server.NewHTTPServer(cfg.Server, handler)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	s.Run()
	b.Run(ctx)
	<-ctx.Done()
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.Stop(stopCtx)
	if err != nil {
		slog.Error("Error stopping server", "error", err)
	}
	b.Stop()
}
