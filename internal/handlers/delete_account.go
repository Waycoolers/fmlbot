package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeleteAccount(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "delete_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "delete_cancel"),
		),
	)

	text := "Ты уверен, что хочешь удалить аккаунт? Все твои пользовательские данные тоже будут удалены."

	err := h.UI.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
	log.Printf("Бот ответил: %v", text)
}

func (h *Handler) HandleDeleteCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	switch cb.Data {
	case "delete_confirm":
		partnerID, err := h.Store.GetPartnerID(ctx, userID)
		if err != nil {
			h.UI.RemoveButtons(chatID, messageID)
			return err
		}

		if partnerID != 0 {
			err = h.Store.RemovePartners(ctx, userID, partnerID)
			if err != nil {
				h.UI.RemoveButtons(chatID, messageID)
				return err
			}

			err = h.Store.DeleteUser(ctx, userID)
			if err != nil {
				h.UI.RemoveButtons(chatID, messageID)
				return err
			}
			h.Reply(partnerID, "Твой партнёр удалил свой аккаунт 💔")
		} else {
			err = h.Store.DeleteUser(ctx, userID)
			if err != nil {
				h.UI.RemoveButtons(chatID, messageID)
				return err
			}
		}

		h.Reply(chatID, "Твой аккаунт успешно удалён 💔")

	case "delete_cancel":
		h.Reply(chatID, "Удаление аккаунта отменено ✅")
	}
	h.UI.RemoveButtons(chatID, messageID)
	return nil
}
