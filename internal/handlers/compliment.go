package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) AddCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, "awaiting_compliment")
	if err != nil {
		h.handleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

}
