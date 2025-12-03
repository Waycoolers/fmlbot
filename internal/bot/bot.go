package bot

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/models"
	"github.com/Waycoolers/fmlbot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func clearOldUpdates(api *tgbotapi.BotAPI) error {
	var lastID int

	for {
		updates, err := api.GetUpdates(tgbotapi.UpdateConfig{
			Offset:  0,
			Limit:   100,
			Timeout: 0,
		})
		if err != nil {
			return err
		}

		if len(updates) == 0 {
			break
		}

		lastID = updates[len(updates)-1].UpdateID
		_, _ = api.GetUpdates(tgbotapi.UpdateConfig{
			Offset: lastID + 1,
			Limit:  1,
		})
	}

	return nil
}

type Bot struct {
	Api    *tgbotapi.BotAPI
	store  *storage.Storage
	router *Router
}

func New(cfg *config.Config, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	commands := []tgbotapi.BotCommand{
		{Command: string(models.Start), Description: "Запустить бота и зарегистрироваться"},
		{Command: string(models.SetPartner), Description: "Добавить партнера"},
		{Command: string(models.DeletePartner), Description: "Удалить партнера"},
		{Command: string(models.DeleteAccount), Description: "Удалить аккаунт"},
		{Command: string(models.AddCompliment), Description: "Добавить комплимент"},
		{Command: string(models.GetCompliments), Description: "Получить свои комплименты"},
		{Command: string(models.DeleteCompliment), Description: "Удалить комплимент"},
		{Command: string(models.ReceiveCompliment), Description: "Получить комплимент"},
		{Command: string(models.Cancel), Description: "Отмена"},
	}

	_, err = api.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Ошибка установки команд: %v", err)
	}

	handler := handlers.New(api, store)
	router := NewRouter(handler)

	return &Bot{Api: api, store: store, router: router}, nil
}

func (b *Bot) Run(ctx context.Context) {
	log.Printf("Бот %s запущен", b.Api.Self.UserName)

	if err := clearOldUpdates(b.Api); err != nil {
		log.Printf("Ошибка очистки старых апдейтов: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Api.GetUpdatesChan(u)

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
