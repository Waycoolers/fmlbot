package config

import (
	"os"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	Token string
}

func Load() (*Config, error) {
	_ = dotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, ErrMissingToken
	}

	return &Config{Token: token}, nil
}

var ErrMissingToken = os.ErrNotExist
