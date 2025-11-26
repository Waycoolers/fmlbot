package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeleteAccount(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "delete_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "delete_cancel"),
		),
	)

	message := tgbotapi.NewMessage(chatID, "Вы уверены, что хотите удалить аккаунт? Это действие нельзя отменить.")
	message.ReplyMarkup = buttons

	_, err := h.api.Send(message)
	if err != nil {
		h.handleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
	log.Printf("Бот ответил: %v", message.Text)
}

func (h *Handler) HandleDeleteCallback(cb *tgbotapi.CallbackQuery) error {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID

	switch cb.Data {
	case "delete_confirm":
		ctx := context.Background()

		partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
		log.Print(partnerUsername)
		if err != nil {
			h.handleErr(chatID, "Ошибка при попытке получить username партнера", err)
			break
		}

		_ = h.Store.ClearPartnerReferences(ctx, userID)

		_ = h.Store.DeleteUser(ctx, userID)

		if partnerUsername != "" {
			partnerID, er := h.Store.GetUserIDByUsername(ctx, partnerUsername)
			_ = h.Store.SetPartner(ctx, partnerID, "")
			if er == nil {
				h.Reply(partnerID, "Твой партнёр удалил свой аккаунт 💔")
			}
		}

		h.Reply(chatID, "Твой аккаунт успешно удалён 💔")

	case "delete_cancel":
		h.Reply(chatID, "Удаление аккаунта отменено ✅")
	}

	emptyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, cb.Message.MessageID, emptyMarkup)
	_, err := h.api.Request(edit)
	if err != nil {
		log.Printf("Ошибка при убирании кнопок: %v", err)
	}
	return err
}
