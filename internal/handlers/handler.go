package handlers

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
)

type Handler struct {
	UI    *ui.MenuUI
	Store *storage.Storage
}

func New(ui *ui.MenuUI, store *storage.Storage) *Handler {
	return &Handler{UI: ui, Store: store}
}

func (h *Handler) Reply(chatID int64, text string) {
	err := h.UI.Client.SendMessage(chatID, text)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (h *Handler) HandleErr(chatID int64, msg string, err error) {
	h.Reply(chatID, "Произошла ошибка 😔")
	log.Printf("%s: %v", msg, err)
}
