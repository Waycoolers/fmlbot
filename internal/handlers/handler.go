package handlers

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	api   *tgbotapi.BotAPI
	Store *storage.Storage
}

func New(api *tgbotapi.BotAPI, store *storage.Storage) *Handler {
	return &Handler{api: api, Store: store}
}

func (h *Handler) Reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	_, err := h.api.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке ответа: %v", err)
	}
	log.Printf("Бот ответил: %v", msg.Text)
}
