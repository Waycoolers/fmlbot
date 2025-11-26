package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) AddCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, models.AwaitingCompliment)
	if err != nil {
		h.handleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

	h.Reply(chatID, "Введи комплимент\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
}

func (h *Handler) ProcessCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	complimentText := msg.Text

	if complimentText == "" {
		err := h.Store.SetUserState(ctx, userID, models.Empty)
		if err != nil {
			h.handleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		h.Reply(chatID, "Некорректный ввод")
		return
	}

	err := h.Store.SetUserState(ctx, userID, models.Empty)
	if err != nil {
		h.handleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.AddCompliment(ctx, userID, complimentText)
	if err != nil {
		h.handleErr(chatID, "Ошибка при добавлении комплимента", err)
		return
	}

	h.Reply(chatID, "Комплимент успешно добавлен")
}
