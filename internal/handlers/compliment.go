package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func truncateText(text string, maxLength int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

func (h *Handler) AddCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, models.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
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
			h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		h.Reply(chatID, "Некорректный ввод")
		return
	}

	err := h.Store.SetUserState(ctx, userID, models.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.AddCompliment(ctx, userID, complimentText)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении комплимента", err)
		return
	}

	h.Reply(chatID, "Комплимент успешно добавлен")
}

func (h *Handler) GetCompliments(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	var reply string

	compliments, isSentList, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) комплиментов. Добавь комплимент с помощью "+string(models.AddCompliment))
		return
	}

	var activeCompliments string
	var sentCompliments string
	for i, compliment := range compliments {
		if !isSentList[i] {
			activeCompliments += "👉 " + compliment + "\n\n"
		} else {
			sentCompliments += "👉 " + compliment + "\n\n"
		}
	}

	if sentCompliments != "" {
		reply += "<b>Отправленные комплименты:</b>\n\n" + sentCompliments + "\n"
	}
	if activeCompliments != "" {
		reply += "<b>Заготовленные комплименты:</b>\n\n" + activeCompliments
	}

	h.Reply(chatID, reply)
}

func (h *Handler) DeleteCompliment(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	compliments, err := h.Store.GetActiveCompliments(ctx, userID)
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

	for i, compliment := range compliments {
		buttonText := truncateText(compliment, 30)
		callbackData := fmt.Sprintf("delete_compliment:%d", i)

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

func (h *Handler) HandleDeleteComplimentCallback(cb *tgbotapi.CallbackQuery) error {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if strings.HasPrefix(data, "delete_compliment:") {
		indexStr := strings.TrimPrefix(data, "delete_compliment:")
		index, _ := strconv.Atoi(indexStr)

		err := h.Store.DeleteCompliment(context.Background(), cb.From.ID, index)
		if err != nil {
			return err
		}

		h.Reply(chatID, "Комплимент успешно удален! ✅")
	} else if data == "cancel_deletion" {
		h.Reply(chatID, "Удаление комплимента отменено")
	}

	emptyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, emptyMarkup)
	_, err := h.api.Request(edit)
	if err != nil {
		return err
	}
	return nil
}
