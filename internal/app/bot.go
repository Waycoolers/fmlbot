package app

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/client"
	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/domain"
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

	commands := []tgbotapi.BotCommand{
		{Command: string(domain.Start), Description: "Запустить бота и зарегистрироваться"},
		{Command: string(domain.SetPartner), Description: "Добавить партнера"},
		{Command: string(domain.DeletePartner), Description: "Удалить партнера"},
		{Command: string(domain.DeleteAccount), Description: "Удалить аккаунт"},
		{Command: string(domain.AddCompliment), Description: "Добавить комплимент"},
		{Command: string(domain.GetCompliments), Description: "Получить свои комплименты"},
		{Command: string(domain.DeleteCompliment), Description: "Удалить комплимент"},
		{Command: string(domain.ReceiveCompliment), Description: "Получить комплимент"},
		{Command: string(domain.Cancel), Description: "Отмена"},
	}

	_, err = api.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Ошибка установки команд: %v", err)
	}

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
