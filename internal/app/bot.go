package app

import (
	"context"
	"log"
	"time"

	"github.com/Waycoolers/fmlbot/internal/client"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/domain"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/internal/scheduler"
	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
	"github.com/redis/go-redis/v9"
)

type Bot struct {
	Client    domain.BotClient
	router    *Router
	scheduler *scheduler.Scheduler
}

func New(cfg *config.Config, store *storage.Storage, rdb *redis.Client) (*Bot, error) {
	telegramClient := client.NewTelegramClient(cfg)
	menuUI := ui.New(telegramClient)
	importantDateDrafts := redis_store.NewImportantDateDraftStore(rdb, 15*time.Minute)
	handler := handlers.New(menuUI, store, importantDateDrafts)
	s := scheduler.New(handler)
	router := NewRouter(handler)

	return &Bot{Client: telegramClient, router: router, scheduler: s}, nil
}

func (b *Bot) Run(ctx context.Context) {
	log.Printf("Бот запущен")

	b.scheduler.Run(ctx)

	updates := b.Client.GetUpdatesChan()

	for {
		select {
		case <-ctx.Done():
			log.Println("Контекст завершён, останавливаем бота...")
			return

		case update, ok := <-updates:
			if !ok {
				log.Println("Канал updates закрыт, бот остановлен.")
				return
			}

			b.router.HandleUpdate(ctx, update)
		}
	}
}
