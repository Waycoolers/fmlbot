package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func truncateText(text string, maxLength int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

func (h *Handler) DeleteCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	compliments, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "У тебя пока нет запланированных комплиментов 😔")
		return
	}

	message := tgbotapi.NewMessage(chatID, "🗑 <b>Выбери комплимент для удаления</b>")
	message.ParseMode = "HTML"

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, compliment := range compliments {
		if compliment.IsSent {
			continue
		}

		buttonText := truncateText(compliment.Text, 30)
		callbackData := fmt.Sprintf("delete_compliment:%d", compliment.ID)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		}
		keyboard = append(keyboard, row)
	}

	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_deletion"),
	})

	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	_, err = h.api.Send(message)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
	log.Printf("Бот ответил: %v", message.Text)
}

func (h *Handler) HandleDeleteComplimentCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if strings.HasPrefix(data, "delete_compliment:") {
		complimentIDStr := strings.TrimPrefix(data, "delete_compliment:")
		complimentID, _ := strconv.Atoi(complimentIDStr)

		err := h.Store.DeleteCompliment(ctx, cb.From.ID, int64(complimentID))
		if err != nil {
			h.RemoveButtons(chatID, messageID)
			return err
		}

		h.Reply(chatID, "Комплимент успешно удален! ✅")
	} else if data == "cancel_deletion" {
		h.Reply(chatID, "Удаление комплимента отменено")
	}
	h.RemoveButtons(chatID, messageID)
	return nil
}
