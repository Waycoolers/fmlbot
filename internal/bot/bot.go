package bot

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Bot struct {
	api   *tgbotapi.BotAPI
	store *storage.Storage
}

func New(cfg *config.Config, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	return &Bot{api: api, store: store}, nil
}

func (b *Bot) Run() {
	log.Printf("Бот %s запущен", b.api.Self.UserName)

	_ = godotenv.Load()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			log.Print("Клиент вызвал: /start")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой fmlbot 💖")
			log.Printf("Бот ответил: %s", msg.Text)
			_, err := b.api.Send(msg)
			if err != nil {
				log.Fatalf("Ошибка при отправке приветствия: %v", err)
			}

		case "/compliment":
			log.Print("Клиент вызвал: /compliment")
			ctx := context.Background()

			// преобразуем string значение из .env в int
			limitStr := os.Getenv("LIMIT_COMPLIMENTS_PER_DAY")
			dailyLimit, err := strconv.Atoi(limitStr)
			if err != nil {
				dailyLimit = 3 // дефолтное значение
			}

			userID := update.Message.Chat.ID

			canSend, err := b.store.CanSendCompliment(ctx, userID, dailyLimit)
			if err != nil {
				log.Println(err)
				break
			}

			if !canSend {
				msg := tgbotapi.NewMessage(userID, "Комплименты на сегодня закончились.")
				log.Printf("Бот ответил: %s", msg.Text)
				_, err := b.api.Send(msg)
				if err != nil {
					log.Fatalf("Ошибка при отправке сообщения о том, что лимит комплиментов исчерпан: %v", err)
				}
				break
			}

			complimentID, text, err := b.store.GetNextCompliment(ctx)
			if err != nil {
				text = "😅 У меня сейчас нет комплиментов, но ты всё равно чудесная!"
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			log.Printf("Бот ответил: %s", msg.Text)

			err = b.store.RecordCompliment(ctx, userID, complimentID)
			if err != nil {
				log.Fatalf("Ошибка при записи комплимента в таблицу с историей: %v", err)
			}

			_, err = b.api.Send(msg)
			if err != nil {
				log.Fatalf("Ошибка при отправке комплимента: %v", err)
			}
		}
	}
}
