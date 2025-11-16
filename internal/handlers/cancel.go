package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Cancel(msg *tgbotapi.Message) {
	err := h.Store.SetUserState(context.Background(), msg.Chat.ID, "")
	if err != nil {
		log.Printf("Ошибка при сброса состояния пользователя: %v", err)
		return
	}
	h.Reply(msg.Chat.ID, "Действие отменено")
}
