package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeletePartner(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(userID, "У тебя ещё не добавлен партнер")
		return
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "delete_partner_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "delete_partner_cancel"),
		),
	)

	partnerUsername, err := h.Store.GetUsername(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить username партнера", err)
		return
	}

	text := "Вы уверены, что хотите удалить партнёра @" + partnerUsername + "?"

	err = h.UI.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
	log.Printf("Бот ответил: %v", text)
}

func (h *Handler) HandleDeletePartnerCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	switch cb.Data {
	case "delete_partner_confirm":
		partnerID, err := h.Store.GetPartnerID(ctx, userID)
		if err != nil {
			h.UI.RemoveButtons(chatID, messageID)
			return err
		}

		err = h.Store.RemovePartners(ctx, userID, partnerID)
		if err != nil {
			h.UI.RemoveButtons(chatID, messageID)
			return err
		}

		h.Reply(chatID, "Партнёр успешно удалён 💔")
		h.Reply(partnerID, "Твой партнёр отписался от тебя 💔")

	case "delete_partner_cancel":
		h.Reply(chatID, "Удаление партнёра отменено")
	}
	h.UI.RemoveButtons(chatID, messageID)
	return nil
}
