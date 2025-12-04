package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Cancel(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	userState, err := h.Store.GetUserState(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении состояния пользователя", err)
		return
	}

	if userState == domain.Empty {
		return
	}

	err = h.Store.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния пользователя", err)
		return
	}
	h.Reply(chatID, "Действие отменено")
}
