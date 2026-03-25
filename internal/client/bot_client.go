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

func convertUpdate(upd tgbotapi.Update) domain.Update {
	if upd.Message != nil {
		return domain.Update{
			Message: &domain.Message{
				ChatID:    upd.Message.Chat.ID,
				UserID:    upd.Message.From.ID,
				UserName:  upd.Message.From.UserName,
				FirstName: upd.Message.From.FirstName,
				Text:      upd.Message.Text,
			},
			CallbackQuery: nil,
		}
	} else if upd.CallbackQuery != nil {
		return domain.Update{
			Message: nil,
			CallbackQuery: &domain.CallbackQuery{
				ChatID:    upd.CallbackQuery.Message.Chat.ID,
				UserID:    upd.CallbackQuery.From.ID,
				MessageID: upd.CallbackQuery.Message.MessageID,
				Data:      upd.CallbackQuery.Data,
				UserName:  upd.CallbackQuery.From.UserName,
				Message:   upd.CallbackQuery.Message.Text,
			},
		}
	}
	return domain.Update{}
}

type TelegramClient struct {
	api            *tgbotapi.BotAPI
	updatesTimeout int
}

func NewTelegramClient(cfg *config.Config) domain.BotClient {
	api, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	return &TelegramClient{api: api, updatesTimeout: cfg.Bot.UpdatesTimeout}
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

func convertInlineKeyboard(keyboard domain.InlineKeyboard) tgbotapi.InlineKeyboardMarkup {
	var tgRaws [][]tgbotapi.InlineKeyboardButton
	for _, raw := range keyboard.Rows {
		var tgRaw []tgbotapi.InlineKeyboardButton
		for _, btn := range raw.Buttons {
			tgBtn := tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Data)
			tgRaw = append(tgRaw, tgBtn)
		}
		tgRaws = append(tgRaws, tgRaw)
	}
	tgKeyboard := tgbotapi.NewInlineKeyboardMarkup(tgRaws...)
	return tgKeyboard
}

func (c *TelegramClient) SendWithInlineKeyboard(chatID int64, text string, keyboard domain.InlineKeyboard) error {
	markup := convertInlineKeyboard(keyboard)

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

func (c *TelegramClient) EditMessageReplyMarkup(chatID int64, messageID int, keyboard domain.InlineKeyboard) error {
	markup := convertInlineKeyboard(keyboard)

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, markup)
	_, err := c.api.Request(edit)
	if err != nil {
		log.Printf("Ошибка при редактировании кнопок: %v", err)
	}
	return err
}

func (c *TelegramClient) GetUpdatesChan() <-chan domain.Update {
	if err := clearOldUpdates(c.api); err != nil {
		log.Printf("Ошибка очистки старых апдейтов: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = c.updatesTimeout
	tgChan := c.api.GetUpdatesChan(u)
	domainChan := make(chan domain.Update)

	go func() {
		defer close(domainChan)
		for update := range tgChan {
			domainChan <- convertUpdate(update)
		}
	}()

	return domainChan
}

func (c *TelegramClient) StopReceivingUpdates() {
	c.api.StopReceivingUpdates()
}

func (c *TelegramClient) SendKeyboard(chatID int64, text string, keyboard domain.Keyboard) (domain.Message, error) {
	markup := convertKeyboard(keyboard)

	markup.ResizeKeyboard = true
	markup.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = markup

	botMessage, err := c.api.Send(msg)
	if err != nil {
		return domain.Message{}, err
	}
	message := domain.Message{
		ChatID:    botMessage.Chat.ID,
		UserID:    botMessage.From.ID,
		UserName:  botMessage.From.UserName,
		FirstName: botMessage.From.FirstName,
		Text:      botMessage.Text,
	}
	return message, nil
}

func (c *TelegramClient) DeleteMessage(chatID int64, messageID int) error {
	req := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := c.api.Request(req)
	return err
}

func convertKeyboard(keyboard domain.Keyboard) tgbotapi.ReplyKeyboardMarkup {
	var tgRaws [][]tgbotapi.KeyboardButton
	for _, raw := range keyboard.Rows {
		var tgRaw []tgbotapi.KeyboardButton
		for _, btn := range raw.Buttons {
			tgBtn := tgbotapi.NewKeyboardButton(string(btn.Command))
			tgRaw = append(tgRaw, tgBtn)
		}
		tgRaws = append(tgRaws, tgRaw)
	}
	tgKeyboard := tgbotapi.NewReplyKeyboard(tgRaws...)
	return tgKeyboard
}
