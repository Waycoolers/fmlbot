package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Cancel(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	userState, err := h.Store.GetUserState(ctx, userID)
	if err != nil {
		h.handleErr(chatID, "Ошибка при получении состояния пользователя", err)
		return
	}

	if userState == "" {
		return
	}

	err = h.Store.SetUserState(ctx, userID, "")
	if err != nil {
		h.handleErr(chatID, "Ошибка при сбросе состояния пользователя", err)
		return
	}
	h.Reply(chatID, "Действие отменено")
}
