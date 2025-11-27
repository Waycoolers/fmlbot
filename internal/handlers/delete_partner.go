package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) DeletePartner(msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	partnerUsername, err := h.Store.GetPartnerUsername(context.Background(), userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении юзернейма партнера", err)
		return
	}

	if partnerUsername == "" {
		h.Reply(userID, "У тебя ещё не добавлен партнер")
		return
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "delete_partner_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "delete_partner_cancel"),
		),
	)

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

	switch cb.Data {
	case "delete_partner_confirm":
		ctx := context.Background()
		partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
		if err != nil {
			break
		}

		partnerID, _ := h.Store.GetUserIDByUsername(ctx, partnerUsername)

		err = h.Store.SetPartners(ctx, userID, partnerID, "", "")
		if err != nil {
			break
		}

		h.Reply(chatID, "Партнёр успешно удалён 💔")
		h.Reply(partnerID, "Твой партнёр отписался от тебя 💔")

	case "delete_partner_cancel":
		h.Reply(chatID, "Удаление партнёра отменено")
	}

	emptyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, cb.Message.MessageID, emptyMarkup)
	_, err := h.api.Request(edit)
	if err != nil {
		return err
	}
	return err
}
