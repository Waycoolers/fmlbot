package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Waycoolers/fmlbot/internal/app"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	store, err := storage.New(cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer store.Close()

	err = store.Migrate()
	if err != nil {
		log.Fatalf("Ошибка при запуске миграций: %v", err)
	}

	rdb, err := redis_store.New(cfg.RDB)
	if err != nil {
		log.Fatalf("Ошибка подключения к redis_store: %v", err)
	}
	defer func(rdb *redis.Client) {
		er := rdb.Close()
		if er != nil {
			log.Printf("Ошибка при закрытии подключения к redis_store: %v", er)
		}
	}(rdb)

	b, err := app.New(cfg, store, rdb)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	b.Run(ctx)
}
