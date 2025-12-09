package handlers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowComplimentsMenu(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := "Комплименты"
	count := 0
	maxCount := 1
	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		text = "Добавь партнёра, чтобы получить возможность получать и отправлять комплименты."
	} else {
		count, err = h.Store.GetComplimentCount(ctx, partnerID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при получении количества полученных комплиментов", err)
			return
		}
		maxCount, err = h.Store.GetComplimentMaxCount(ctx, partnerID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при получении максимального количества комплиментов", err)
			return
		}

		if maxCount == -1 {
			text = "Сегодня ты можешь получить еще ♾️ комплиментов."
		} else {
			delta := maxCount - count
			if delta > 0 {
				deltaStr := strconv.Itoa(delta)
				text = "Сегодня ты можешь получить еще <b>" + deltaStr + "</b> комплимент(ов)."
			} else {
				text = "Сегодня ты больше не можешь получать комплименты ("
			}
		}
	}

	err = h.ui.ComplimentsMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню комплиментов", err)
		return
	}
}

func (h *Handler) AddCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, domain.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

	h.Reply(chatID, "Введи комплимент\n(Напиши чтобы отменить это действие)")
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

func (h *Handler) GetCompliments(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	var reply string

	compliments, err := h.Store.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) комплиментов. Добавь комплимент")
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
}

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

func (h *Handler) ReceiveCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "Ты не можешь получить комплимент так как у тебя не добавлен партнёр. "+
			"Сначала добавь партнёра с помощью")
		return
	}

	count, err := h.Store.GetComplimentCount(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении количества полученных комплиментов", err)
		return
	}
	maxCount, err := h.Store.GetComplimentMaxCount(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении максимального количества комплиментов", err)
		return
	}

	if count >= maxCount && maxCount != -1 {
		h.Reply(chatID, "Комплименты на сегодня закончились (")
		return
	}
	count++

	last, err := h.Store.GetComplimentTime(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении времени последнего комплимента", err)
		return
	}
	now := time.Now().UTC()
	log.Print(now)
	log.Print(last)
	if last.Add(1 * time.Hour).After(now) {
		remaining := last.Add(time.Hour).Sub(now)
		mins := int(remaining.Minutes())

		h.Reply(chatID, fmt.Sprintf("Ты уже получал комплимент недавно ❤️\nПопробуй снова через %d минут.", mins))
		return
	}

	allCompliments, err := h.Store.GetCompliments(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
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
		return
	}

	compliment := compliments[0]
	err = h.Store.MarkComplimentSent(ctx, compliment.ID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отметить комплимент как отправленный", err)
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

	err = h.Store.SetComplimentTime(ctx, partnerID)
	if err != nil {
		log.Printf("Ошибка при попытке установить время получения комплимента: %v", err)
	}

	err = h.Store.SetComplimentCount(ctx, partnerID, count)
	if err != nil {
		log.Printf("Ошибка при попытке изменить количество полученных комплиментов: %v", err)
	}
}

func (h *Handler) EditComplimentFrequency(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	actualFreq, err := h.Store.GetComplimentMaxCount(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить частоту комплиментов", err)
		return
	}
	count, err := h.Store.GetComplimentCount(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении количества комплиментов", err)
		return
	}

	actualFreqStr := strconv.Itoa(actualFreq)
	countStr := strconv.Itoa(count)
	if actualFreq == -1 {
		actualFreqStr = "♾️"
	}
	text := "Твой партнёр сегодня получил <b>" + countStr + "/" + actualFreqStr + "</b> комплимент(ов). " +
		"Хочешь изменить лимит? Просто отправь новое значение в чат. Чтобы убрать лимит, отправь «-»."

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingComplimentFrequency)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке установить состояние", err)
		return
	}

	err = h.ui.EditComplimentFrequencyMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню для изменения частоты комплиментов", err)
		return
	}
}

func (h *Handler) ProcessComplimentFrequency(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	freq := msg.Text
	freqInt := 1

	// Валидация
	if freq == "-" {
		freqInt = -1
	} else {
		var err error
		freqInt, err = strconv.Atoi(freq)
		if err != nil || freqInt <= 0 {
			h.Reply(chatID, "Некорректный ввод")
			return
		}
	}

	err := h.Store.SetComplimentMaxCount(ctx, userID, freqInt)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при изменении частоты комплиментов", err)
		return
	}

	h.Reply(chatID, "Лимит изменен")
}
