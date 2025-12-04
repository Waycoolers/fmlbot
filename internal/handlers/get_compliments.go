package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) GetCompliments(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	var reply string

	compliments, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) комплиментов. Добавь комплимент с помощью "+string(domain.AddCompliment))
		return
	}

	var activeCompliments string
	var sentCompliments string
	for _, compliment := range compliments {
		if !compliment.IsSent {
			activeCompliments += "👉 " + compliment.Text + "\n\n"
		} else {
			sentCompliments += "👉 " + compliment.Text + "\n\n"
		}
	}

	if sentCompliments != "" {
		reply += "<b>Отправленные комплименты:</b>\n\n" + sentCompliments + "\n"
	}
	if activeCompliments != "" {
		reply += "<b>Заготовленные комплименты:</b>\n\n" + activeCompliments
	}

	h.Reply(chatID, reply)
}
