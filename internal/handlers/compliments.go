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
)

func (h *Handler) ShowComplimentsMenu(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	text := "❤️ Комплименты"
	count := 0
	maxCount := 1
	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		text = "🤍 Добавь партнёра, и здесь появится магия комплиментов ✨"
	} else {
		count, err = h.Store.UserConfig.GetComplimentCount(ctx, partnerID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при получении количества полученных комплиментов", err)
			return
		}
		maxCount, err = h.Store.UserConfig.GetComplimentMaxCount(ctx, partnerID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при получении максимального количества комплиментов", err)
			return
		}

		if maxCount == -1 {
			text = "💫 Сегодня ты можешь получить ещё ♾️ комплиментов"
		} else {
			delta := maxCount - count
			if delta > 0 {
				text = "💛 Сегодня для тебя доступно ещё <b>" + strconv.Itoa(delta) + "</b> комплимент(ов)"
			} else {
				text = "🌙 На сегодня комплименты закончились. Завтра будет больше тепла 💛"
			}
		}
	}

	err = h.ui.ComplimentsMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню комплиментов", err)
		return
	}
}

func (h *Handler) AddCompliment(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID

	err := h.Store.Users.SetUserState(ctx, userID, domain.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

	h.Reply(chatID, "💌 Напиши комплимент")
}

func (h *Handler) ProcessCompliment(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	complimentText := msg.Text

	if complimentText == "" {
		err := h.Store.Users.SetUserState(ctx, userID, domain.Empty)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		h.Reply(chatID, "Кажется, тут пусто 🙈 Попробуй ещё раз")
		return
	}

	err := h.Store.Users.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	_, err = h.Store.Compliments.AddCompliment(ctx, userID, complimentText)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении комплимента", err)
		return
	}

	h.Reply(chatID, "✨ Готово! Комплимент сохранён и ждёт своего часа 💛")
}

func (h *Handler) GetCompliments(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	var reply string

	compliments, err := h.Store.Compliments.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	if len(compliments) == 0 {
		h.Reply(chatID, "📭 Здесь пока пусто. Добавь первый комплимент — пусть он согревает 🤍")
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

	if activeCompliments != "" {
		reply += "<b>Заготовленные комплименты:</b>\n\n" + activeCompliments
	}
	if sentCompliments != "" {
		reply += "<b>Отправленные комплименты:</b>\n\n" + sentCompliments + "\n"
	}

	h.Reply(chatID, reply)
}

func truncateText(text string, maxLength int) string {
	text = strings.TrimSpace(text)
	runes := []rune(text) // конвертируем в руны
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength-3]) + "..."
}

func (h *Handler) DeleteCompliment(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID

	compliments, err := h.Store.Compliments.GetCompliments(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка комплиментов", err)
		return
	}

	var filtered []domain.Compliment
	for _, c := range compliments {
		if !c.IsSent {
			filtered = append(filtered, c)
		}
	}
	compliments = filtered

	if len(compliments) == 0 {
		h.Reply(chatID, "🌿 У тебя нет комплиментов, которые можно удалить")
		return
	}

	var rows []domain.InlineKeyboardRow

	for _, compliment := range compliments {
		buttonText := truncateText(compliment.Text, 30)
		callbackData := fmt.Sprintf("compliments:delete:confirm:%d", compliment.ID)

		row := domain.InlineKeyboardRow{
			Buttons: []domain.InlineKeyboardButton{
				{Text: buttonText, Data: callbackData},
			},
		}
		rows = append(rows, row)
	}

	rows = append(rows, domain.InlineKeyboardRow{
		Buttons: []domain.InlineKeyboardButton{
			{Text: "↩️ Передумал(а)", Data: "compliments:delete:cancel"},
		},
	})

	text := "🗑 <b>Выбери комплимент, который хочешь убрать</b>"
	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeleteCompliment(ctx context.Context, cb *domain.CallbackQuery) {
	data := cb.Data
	chatID := cb.ChatID
	userID := cb.UserID
	messageID := cb.MessageID

	if strings.HasPrefix(data, "compliments:delete:confirm:") {
		complimentIDStr := strings.TrimPrefix(data, "compliments:delete:confirm:")
		complimentID, _ := strconv.Atoi(complimentIDStr)

		err := h.Store.Compliments.DeleteCompliment(ctx, userID, int64(complimentID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке удалить комплимент", err)
			return
		}

		h.Reply(chatID, "🧹 Готово. Комплимент удалён")
	} else if data == "compliments:delete:cancel" {
		h.Reply(chatID, "🌸 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}

func (h *Handler) ReceiveCompliment(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID

	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "🤍 Чтобы получать комплименты, сначала добавь партнёра")
		return
	}

	count, err := h.Store.UserConfig.GetComplimentCount(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении количества полученных комплиментов", err)
		return
	}
	maxCount, err := h.Store.UserConfig.GetComplimentMaxCount(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении максимального количества комплиментов", err)
		return
	}

	if count >= maxCount && maxCount != -1 {
		h.Reply(chatID, "🌙 На сегодня лимит исчерпан. Завтра будет продолжение 💛")
		return
	}
	count++

	bucket, err := h.Store.UserConfig.GetComplimentsBucket(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении ведра комплиментов", err)
		return
	}
	lastBucketUpdate, err := h.Store.UserConfig.GetLastBucketUpdate(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении последнего обновления ведра комплиментов", err)
		return
	}
	now := time.Now().UTC()
	hoursPassed := int(now.Sub(lastBucketUpdate).Hours())
	updateTime := lastBucketUpdate
	if hoursPassed > 0 {
		newBucket := hoursPassed + bucket
		if newBucket > 2 {
			newBucket = 2
		}
		if newBucket != bucket {
			err = h.Store.UserConfig.UpdateComplimentBucket(ctx, partnerID, newBucket, now)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при обновлении ведра комплиментов", err)
				return
			}
			bucket = newBucket
			updateTime = now
		}
	}

	if bucket == 0 {
		hoursSinceUpdate := now.Sub(updateTime).Hours()
		remaining := updateTime.Add(time.Duration(hoursSinceUpdate+1) * time.Hour).Sub(now)
		mins := int(remaining.Minutes())
		if mins < 0 {
			mins = 0
		}
		h.Reply(chatID, fmt.Sprintf("⏳ Немного терпения\nСледующий комплимент будет доступен через %d мин.", mins))
		return
	}

	allCompliments, err := h.Store.Compliments.GetCompliments(ctx, partnerID)
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
		h.Reply(chatID, "📭 Пока для тебя нет новых комплиментов")
		return
	}

	compliment := compliments[0]
	err = h.Store.Compliments.MarkComplimentSent(ctx, compliment.ID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отметить комплимент как отправленный", err)
		return
	}

	var complimentMessages = []string{
		"💖 <b>Для тебя есть тёплые слова:</b>\n\n«" + compliment.Text + "»",
		"✨ <b>Небольшое послание от твоего человека:</b>\n\n«" + compliment.Text + "»",
		"🌷 <b>Тебе отправили комплимент:</b>\n\n«" + compliment.Text + "»",
	}

	randomIndex := rand.Intn(len(complimentMessages))
	h.Reply(chatID, complimentMessages[randomIndex])
	h.Reply(partnerID,
		"💌 <b>Комплимент доставлен</b>\n\nТы только что порадовал(а) своего партнёра ✨\n\n«"+compliment.Text+"»",
	)

	err = h.Store.UserConfig.SetComplimentTime(ctx, partnerID)
	if err != nil {
		log.Printf("Ошибка при попытке установить время получения комплимента: %v", err)
	}

	err = h.Store.UserConfig.TakeComplimentFromBucket(ctx, partnerID)
	if err != nil {
		log.Printf("Ошибка при взятии комплимента из ведра: %v", err)
	}

	err = h.Store.UserConfig.SetComplimentCount(ctx, partnerID, count)
	if err != nil {
		log.Printf("Ошибка при попытке изменить количество полученных комплиментов: %v", err)
	}
}

func (h *Handler) EditComplimentFrequency(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID

	actualFreq, err := h.Store.UserConfig.GetComplimentMaxCount(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить частоту комплиментов", err)
		return
	}
	count, err := h.Store.UserConfig.GetComplimentCount(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении количества комплиментов", err)
		return
	}

	actualFreqStr := strconv.Itoa(actualFreq)
	countStr := strconv.Itoa(count)
	if actualFreq == -1 {
		actualFreqStr = "♾️"
	}
	text := "💛 Сегодня твой партнёр получил <b>" + countStr + "/" + actualFreqStr +
		"</b> комплимент(ов).\n\n" +
		"Хочешь изменить лимит?\n" +
		"• отправь число\n" +
		"• или «-», чтобы убрать лимит"

	err = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingComplimentFrequency)
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

func (h *Handler) ProcessComplimentFrequency(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	freq := msg.Text
	freqInt := 1

	// Валидация
	if freq == "-" {
		freqInt = -1
	} else {
		var err error
		freqInt, err = strconv.Atoi(freq)
		if err != nil || freqInt <= 0 {
			h.Reply(chatID, "🤔 Я не понял. Отправь число или «-»")
			return
		}
	}

	err := h.Store.UserConfig.SetComplimentMaxCount(ctx, userID, freqInt)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при изменении лимита", err)
		return
	}

	h.Reply(chatID, "✨ Лимит обновлён")
}
