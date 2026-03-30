package api

import (
	"log/slog"
	"os"

	"github.com/Waycoolers/fmlbot/services/api/internal/config"
	"github.com/Waycoolers/fmlbot/services/api/internal/logger"
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

}
