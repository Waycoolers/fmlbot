package bot

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	return &Bot{api: api}, nil
}

func (b *Bot) Run() {
	log.Printf("Бот %s запущен", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой fmlbot 💖")
			_, err := b.api.Send(msg)
			if err != nil {
				log.Fatalf("Ошибка при отправке сообщения: %v", err)
			}
		}
	}
}
