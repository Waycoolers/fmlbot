package client

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/Waycoolers/fmlbot/internal/domain"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

type TelegramClient struct {
	api *tgbotapi.BotAPI
}

func NewTelegramClient(cfg *config.Config) domain.BotClient {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	return &TelegramClient{api: api}
}

func (c *TelegramClient) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	_, err := c.api.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке ответа: %v", err)
	}
	log.Printf("Бот ответил: %v", msg.Text)
	return err
}

func (c *TelegramClient) SendWithInlineKeyboard(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	msg.ParseMode = "HTML"
	_, err := c.api.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения с кнопками: %v", err)
	}
	log.Printf("Бот ответил: %v", msg.Text)
	return err
}

func (c *TelegramClient) EditMessageReplyMarkup(chatID int64, messageID int, markup tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, markup)
	_, err := c.api.Request(edit)
	if err != nil {
		log.Printf("Ошибка при редактировании кнопок: %v", err)
	}
	return err
}

func (c *TelegramClient) GetUpdatesChan() <-chan tgbotapi.Update {
	if err := clearOldUpdates(c.api); err != nil {
		log.Printf("Ошибка очистки старых апдейтов: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return c.api.GetUpdatesChan(u)
}

func (c *TelegramClient) StopReceivingUpdates() {
	c.api.StopReceivingUpdates()
}

func (c *TelegramClient) Send(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	return c.api.Send(msg)
}

func (c *TelegramClient) DeleteMessage(chatID int64, messageID int) error {
	req := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := c.api.Request(req)
	return err
}
