package api

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Waycoolers/fmlbot/common/logger"
	"github.com/Waycoolers/fmlbot/services/api/internal/config"
	"github.com/Waycoolers/fmlbot/services/api/internal/handlers"
	"github.com/Waycoolers/fmlbot/services/api/internal/scheduler"
	"github.com/Waycoolers/fmlbot/services/api/internal/sender"
	"github.com/Waycoolers/fmlbot/services/api/internal/server"
	"github.com/Waycoolers/fmlbot/services/api/internal/storage"
	"github.com/Waycoolers/fmlbot/services/api/internal/usecases"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env", "../../.env")

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

	uc := usecases.New(store.Repos)
	h := handlers.New(uc)

	srv := server.New(cfg.Server, h)

	snd := sender.NewHTTPSender(cfg.BotURL)
	sched := scheduler.New(h, snd)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sched.Run(ctx)
	srv.Start()

	<-ctx.Done()
	sched.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Stop(shutdownCtx)
	if err != nil {
		slog.Error("Error stopping server", "error", err)
	}
}
