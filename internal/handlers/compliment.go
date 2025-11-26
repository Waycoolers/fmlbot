package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) AddCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, "awaiting_compliment")
	if err != nil {
		h.Reply(userID, "Произошла ошибка 😔")
		log.Printf("Ошибка при установке состояния awaiting_compliment: %v", err)
		return
	}

}
