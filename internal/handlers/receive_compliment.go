package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ReceiveCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "Ты не можешь получить комплимент так как у тебя не добавлен партнёр. "+
			"Сначала добавь партнёра с помощью "+string(models.SetPartner))
		return
	}

	allCompliments, err := h.Store.GetCompliments(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	// Выбираем только активные комплименты
	var compliments []models.Compliment
	for _, compliment := range allCompliments {
		if !compliment.IsSent {
			compliments = append(compliments, compliment)
		}
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Тебе не отправили комплимент (((")
		return
	}

	compliment := compliments[0]
	err = h.Store.MarkComplimentSent(ctx, compliment.ID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отметить комплимент как отправленный", err)
		return
	}

	h.Reply(chatID, compliment.Text)
	h.Reply(partnerID, "<b>Твой партнер получил комплимент:</b>\n"+compliment.Text)
}
