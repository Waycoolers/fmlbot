package handlers

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	ui    *ui.MenuUI
	Store *storage.Storage
}

func New(ui *ui.MenuUI, store *storage.Storage) *Handler {
	return &Handler{ui: ui, Store: store}
}

func (h *Handler) ShowMainMenu(_ context.Context, chatID int64) {
	err := h.ui.MainMenu(chatID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить главное меню", err)
		return
	}
}

func (h *Handler) Reply(chatID int64, text string) {
	err := h.ui.Client.SendMessage(chatID, text)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (h *Handler) ReplyUnknownCallback(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	h.Reply(chatID, "Используй кнопки")
}

func (h *Handler) HandleErr(chatID int64, msg string, err error) {
	h.Reply(chatID, "Произошла ошибка 😔")
	log.Printf("%s: %v", msg, err)
}
