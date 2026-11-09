package telegram

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
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
				ID:        upd.CallbackQuery.ID,
				ChatID:    upd.CallbackQuery.Message.Chat.ID,
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
	// Таймаут больше long polling, иначе getUpdates обрывался бы раньше времени
	httpClient := &http.Client{Timeout: time.Duration(cfg.Bot.UpdatesTimeout+30) * time.Second}
	api, err := tgbotapi.NewBotAPIWithClient(cfg.Bot.Token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	return &TelegramClient{api: api, updatesTimeout: cfg.Bot.UpdatesTimeout}
}

// Лимит Telegram - 4096 символов, берём с запасом
const maxMessageLength = 4000

// splitMessage делит длинный текст на части, стараясь резать по границе строки.
func splitMessage(text string, limit int) []string {
	runes := []rune(text)
	var parts []string
	for len(runes) > limit {
		cut := limit
		for i := limit - 1; i > 0; i-- {
			if runes[i] == '\n' {
				cut = i
				break
			}
		}
		if part := strings.TrimSpace(string(runes[:cut])); part != "" {
			parts = append(parts, part)
		}
		runes = runes[cut:]
	}
	if part := strings.TrimSpace(string(runes)); part != "" {
		parts = append(parts, part)
	}
	return parts
}

// classifySendError отделяет окончательные отказы Telegram от временных сбоев.
func classifySendError(err error) error {
	var tgErr *tgbotapi.Error
	if errors.As(err, &tgErr) && (tgErr.Code == 400 || tgErr.Code == 403) {
		return fmt.Errorf("%w: %s", domain.ErrRecipientUnavailable, tgErr.Message)
	}
	return err
}

func (c *TelegramClient) SendMessage(chatID int64, text string) error {
	for _, part := range splitMessage(text, maxMessageLength) {
		msg := tgbotapi.NewMessage(chatID, part)
		msg.ParseMode = "HTML"
		if _, err := c.api.Send(msg); err != nil {
			log.Printf("Ошибка при отправке ответа: %v", err)
			return classifySendError(err)
		}
	}
	return nil
}

func (c *TelegramClient) AnswerCallback(callbackID string) error {
	_, err := c.api.Request(tgbotapi.NewCallback(callbackID, ""))
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
	return err
}

func (c *TelegramClient) EditMessageReplyMarkup(chatID int64, messageID int, keyboard domain.InlineKeyboard) error {
	markup := convertInlineKeyboard(keyboard)

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, markup)
	_, err := c.api.Request(edit)
	return err
}

func (c *TelegramClient) DeleteMessageReplyMarkup(chatID int64, messageID int) error {
	empty := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	del := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, empty)
	_, err := c.api.Request(del)
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
