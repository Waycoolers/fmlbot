package bot

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/models"
	"github.com/Waycoolers/fmlbot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api    *tgbotapi.BotAPI
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
		{Command: string(models.Delete), Description: "Удалить аккаунт"},
		{Command: string(models.AddCompliment), Description: "Добавить комплимент"},
		{Command: string(models.GetCompliments), Description: "Получить свои комплименты"},
		{Command: string(models.DeleteCompliment), Description: "Удалить комплимент"},
		{Command: string(models.Cancel), Description: "Отмена"},
	}

	_, err = api.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Ошибка установки команд: %v", err)
	}

	handler := handlers.New(api, store)
	router := NewRouter(handler)

	return &Bot{api: api, store: store, router: router}, nil
}

func (b *Bot) Run() {
	log.Printf("Бот %s запущен", b.api.Self.UserName)

	for {
		updates, err := b.api.GetUpdates(tgbotapi.UpdateConfig{
			Offset:  0,
			Limit:   100,
			Timeout: 0,
		})
		if err != nil || len(updates) == 0 {
			break
		}

		lastUpdateID := updates[len(updates)-1].UpdateID
		_, err = b.api.GetUpdates(tgbotapi.UpdateConfig{Offset: lastUpdateID + 1})
		if err != nil {
			log.Printf("Ошибка при запуске бота: %v", err)
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			b.router.HandleUpdate(update)
			continue
		}

		if update.Message != nil {
			b.router.HandleUpdate(update)
		}
	}
}
