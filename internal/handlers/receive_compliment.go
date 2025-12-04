package handlers

import (
	"context"
	"math/rand"

	"github.com/Waycoolers/fmlbot/internal/domain"
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
			"Сначала добавь партнёра с помощью "+string(domain.SetPartner))
		return
	}

	allCompliments, err := h.Store.GetCompliments(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	// Выбираем только активные комплименты
	var compliments []domain.Compliment
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

	var complimentMessages = []string{
		"🌙 <b>Твоя половинка оставила для тебя нежное послание:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова согреют твоё сердце сегодня 💖",
		"✨ <b>Твой светлый лучик прислал тебе маленькое чудо:</b>\n\n«" + compliment.Text + "»\n\nУлыбнись! Этот комплимент специально для тебя 😄💛",
		"💛 <b>Твой дорогой человек хочет поднять тебе настроение:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова дадут тебе силы и радость сегодня 🌼",
		"🌹 <b>Твоя нежная половинка отправила тебе тёплые слова:</b>\n\n«" + compliment.Text + "»\n\nПусть этот маленький знак внимания согреет твоё сердце 💖",
		"🌸 <b>Твой любимый человек оставил для тебя послание:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова принесут тебе немного тепла и улыбок 💛",
	}

	randomIndex := rand.Intn(len(complimentMessages))
	h.Reply(chatID, complimentMessages[randomIndex])
	h.Reply(partnerID,
		"🌷 <b>Твой комплимент нашёл своего адресата!</b>\n"+
			"Ты только что сделал своего партнёра чуточку счастливее 😊\n\n"+
			"<i>Ты отправил:</i>\n"+"«"+compliment.Text+"»",
	)
}
