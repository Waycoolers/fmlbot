package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeletePartner(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "delete_partner_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "delete_partner_cancel"),
		),
	)

	message := tgbotapi.NewMessage(chatID, "Вы уверены, что хотите удалить партнёра?")
	message.ReplyMarkup = buttons

	_, err := h.api.Send(message)
	if err != nil {
		log.Printf("Ошибка при отправке подтверждения: %v", err)
	}
	log.Printf("Бот ответил: %v", message.Text)
}

func (h *Handler) HandleDeletePartnerCallback(cb *tgbotapi.CallbackQuery) error {
	userID := cb.From.ID

	switch cb.Data {
	case "delete_partner_confirm":
		ctx := context.Background()
		partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
		if err != nil {
			log.Printf("Ошибка при попытке получить username партнера: %v", err)
		}

		if partnerUsername == "" {
			h.Reply(userID, "У тебя и так не добавлен партнер")
			return nil
		}
		partnerID, _ := h.Store.GetUserIDByUsername(ctx, partnerUsername)

		err = h.Store.SetPartner(ctx, userID, "")
		if err != nil {
			log.Printf("Ошибка при удалении партнера у юзера: %v", err)
		}

		err = h.Store.SetPartner(ctx, partnerID, "")
		if err != nil {
			log.Printf("Ошибка при удалении партнера у партнера: %v", err)
		}

		h.Reply(userID, "Партнёр успешно удалён 💔")
		h.Reply(partnerID, "Твой партнёр отписался от тебя 💔")

	case "delete_partner_cancel":
		h.Reply(userID, "Удаление партнёра отменено ✅")
	}

	emptyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(userID, cb.Message.MessageID, emptyMarkup)
	_, err := h.api.Request(edit)
	if err != nil {
		log.Printf("Ошибка при убирании кнопок: %v", err)
	}
	return err
}
