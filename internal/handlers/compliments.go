package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowComplimentsMenu(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	err := h.ui.ComplimentsMenu(chatID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню комплиментов", err)
		return
	}
}

func (h *Handler) AddCompliment(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	err := h.Store.SetUserState(ctx, userID, domain.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	h.Reply(chatID, "Введи комплимент\n(Напиши "+string(domain.Cancel)+" чтобы отменить это действие)")
	h.ui.RemoveButtons(chatID, messageID)
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

func (h *Handler) GetCompliments(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID
	var reply string

	compliments, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) комплиментов. Добавь комплимент с помощью "+string(domain.AddCompliment))
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	var activeCompliments string
	var sentCompliments string
	for _, compliment := range compliments {
		if !compliment.IsSent {
			activeCompliments += "👉 " + compliment.Text + "\n\n"
		} else {
			sentCompliments += "👉 " + compliment.Text + "\n\n"
		}
	}

	if sentCompliments != "" {
		reply += "<b>Отправленные комплименты:</b>\n\n" + sentCompliments + "\n"
	}
	if activeCompliments != "" {
		reply += "<b>Заготовленные комплименты:</b>\n\n" + activeCompliments
	}

	h.Reply(chatID, reply)
	h.ui.RemoveButtons(chatID, messageID)
}

func truncateText(text string, maxLength int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

func (h *Handler) DeleteCompliment(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	compliments, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "У тебя пока нет запланированных комплиментов 😔")
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, compliment := range compliments {
		if compliment.IsSent {
			continue
		}

		buttonText := truncateText(compliment.Text, 30)
		callbackData := fmt.Sprintf("compliments:delete:confirm:%d", compliment.ID)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		}
		keyboard = append(keyboard, row)
	}

	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "compliments:delete:cancel"),
	})

	text := "🗑 <b>Выбери комплимент для удаления</b>"
	markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, markup)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}
}

func (h *Handler) HandleDeleteCompliment(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if strings.HasPrefix(data, "compliments:delete:confirm:") {
		complimentIDStr := strings.TrimPrefix(data, "compliments:delete:confirm:")
		complimentID, _ := strconv.Atoi(complimentIDStr)

		err := h.Store.DeleteCompliment(ctx, cb.From.ID, int64(complimentID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке удалить комплимент", err)
			return
		}

		h.Reply(chatID, "Комплимент успешно удален! ✅")
	} else if data == "compliments:delete:cancel" {
		h.Reply(chatID, "Удаление комплимента отменено")
	}
	h.ui.RemoveButtons(chatID, messageID)
}

func (h *Handler) ReceiveCompliment(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "Ты не можешь получить комплимент так как у тебя не добавлен партнёр. "+
			"Сначала добавь партнёра с помощью "+string(domain.SetPartner))
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	allCompliments, err := h.Store.GetCompliments(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	// Выбираем только активные комплименты
	var compliments []domain.Compliment
	for _, compliment := range allCompliments {
		if !compliment.IsSent {
			compliments = append(compliments, compliment)
		}
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Тебе не отправили комплимент (((")
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	compliment := compliments[0]
	err = h.Store.MarkComplimentSent(ctx, compliment.ID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отметить комплимент как отправленный", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	var complimentMessages = []string{
		"🌙 <b>Твоя половинка оставила для тебя нежное послание:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова согреют твоё сердце сегодня 💖",
		"✨ <b>Твой светлый лучик прислал тебе маленькое чудо:</b>\n\n«" + compliment.Text + "»\n\nУлыбнись! Этот комплимент специально для тебя 😄💛",
		"💛 <b>Твой дорогой человек хочет поднять тебе настроение:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова дадут тебе силы и радость сегодня 🌼",
		"🌹 <b>Твоя нежная половинка отправила тебе тёплые слова:</b>\n\n«" + compliment.Text + "»\n\nПусть этот маленький знак внимания согреет твоё сердце 💖",
		"🌸 <b>Твой любимый человек оставил для тебя послание:</b>\n\n«" + compliment.Text + "»\n\nПусть эти слова принесут тебе немного тепла и улыбок 💛",
	}

	randomIndex := rand.Intn(len(complimentMessages))
	h.Reply(chatID, complimentMessages[randomIndex])
	h.Reply(partnerID,
		"🌷 <b>Твой комплимент нашёл своего адресата!</b>\n"+
			"Ты только что сделал своего партнёра чуточку счастливее 😊\n\n"+
			"<i>Ты отправил:</i>\n"+"«"+compliment.Text+"»",
	)
	h.ui.RemoveButtons(chatID, messageID)
}
