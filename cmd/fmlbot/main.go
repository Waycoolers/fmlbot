package main

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/bot"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/jmoiron/sqlx"
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
	defer func(DB *sqlx.DB) {
		err := DB.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии подключения к БД: %v", err)
		}
	}(store.DB)

	err = store.Migrate()
	if err != nil {
		log.Fatalf("Ошибка при запуске миграций: %v", err)
	}

	b, err := bot.New(cfg, store)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	b.Run()
}
