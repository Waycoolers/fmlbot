package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeletePartner(msg *tgbotapi.Message) {
	ctx := context.Background()
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

	message := tgbotapi.NewMessage(chatID, "Вы уверены, что хотите удалить партнёра @"+partnerUsername+"?")
	message.ReplyMarkup = buttons

	_, err = h.api.Send(message)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
	log.Printf("Бот ответил: %v", message.Text)
}

func (h *Handler) HandleDeletePartnerCallback(cb *tgbotapi.CallbackQuery) error {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	switch cb.Data {
	case "delete_partner_confirm":
		ctx := context.Background()
		partnerID, err := h.Store.GetPartnerID(ctx, userID)
		if err != nil {
			h.RemoveButtons(chatID, messageID)
			return err
		}

		err = h.Store.RemovePartners(ctx, userID, partnerID)
		if err != nil {
			h.RemoveButtons(chatID, messageID)
			return err
		}

		h.Reply(chatID, "Партнёр успешно удалён 💔")
		h.Reply(partnerID, "Твой партнёр отписался от тебя 💔")

	case "delete_partner_cancel":
		h.Reply(chatID, "Удаление партнёра отменено")
	}
	h.RemoveButtons(chatID, messageID)
	return nil
}
