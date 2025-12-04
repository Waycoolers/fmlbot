package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) AddCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, domain.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

	h.Reply(chatID, "Введи комплимент\n(Напиши "+string(domain.Cancel)+" чтобы отменить это действие)")
}

func (h *Handler) ProcessCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	complimentText := msg.Text

	if complimentText == "" {
		err := h.Store.SetUserState(ctx, userID, domain.Empty)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		h.Reply(chatID, "Некорректный ввод")
		return
	}

	err := h.Store.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	_, err = h.Store.AddCompliment(ctx, userID, complimentText)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении комплимента", err)
		return
	}

	h.Reply(chatID, "Комплимент успешно добавлен")
}
