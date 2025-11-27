package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Cancel(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	userState, err := h.Store.GetUserState(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении состояния пользователя", err)
		return
	}

	if userState == models.Empty {
		return
	}

	err = h.Store.SetUserState(ctx, userID, models.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния пользователя", err)
		return
	}
	h.Reply(chatID, "Действие отменено")
}
