package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Cancel(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.Chat.ID
	userState, err := h.Store.GetUserState(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении состояния пользователя: %v", err)
		return
	}

	if userState == "" {
		return
	}

	err = h.Store.SetUserState(ctx, userID, "")
	if err != nil {
		log.Printf("Ошибка при сбросе состояния пользователя: %v", err)
		return
	}
	h.Reply(msg.Chat.ID, "Действие отменено")
}
