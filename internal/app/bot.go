package app

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/client"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Client client.BotClient
	router *Router
}

func New(cfg *config.Config, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	telegramClient := client.NewTelegramClient(api)
	menuUI := ui.New(telegramClient)
	handler := handlers.New(menuUI, store)
	router := NewRouter(handler)

	return &Bot{Client: telegramClient, router: router}, nil
}

func (b *Bot) Run(ctx context.Context) {
	log.Printf("Бот запущен")

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
