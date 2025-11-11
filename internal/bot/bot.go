package bot

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/handlers"
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

	handler := handlers.New(api, store)
	router := NewRouter(handler)

	return &Bot{api: api, store: store, router: router}, nil
}

func (b *Bot) Run() {
	log.Printf("Бот %s запущен", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case <-updates:
				// просто прогоняем все старые updates
			default:
				return
			}
		}
	}()

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
