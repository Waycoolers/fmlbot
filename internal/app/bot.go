package app

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/client"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/scheduler"
	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
)

type Bot struct {
	Client    client.BotClient
	router    *Router
	scheduler *scheduler.Scheduler
}

func New(cfg *config.Config, store *storage.Storage) (*Bot, error) {
	telegramClient := client.NewTelegramClient(cfg)
	menuUI := ui.New(telegramClient)
	handler := handlers.New(menuUI, store)
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
