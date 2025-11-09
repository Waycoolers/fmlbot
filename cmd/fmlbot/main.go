package main

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/bot"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	store, err := storage.New(&cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer store.DB.Close()

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	b.Run()
}
