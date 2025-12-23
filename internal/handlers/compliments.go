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
	text := "❤️ Комплименты"
	count := 0
	maxCount := 1
	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		text = "🤍 Добавь партнёра, и здесь появится магия комплиментов ✨"
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

func (h *Handler) AddCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, domain.AwaitingCompliment)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_compliment", err)
		return
	}

	h.Reply(chatID, "💌 Напиши комплимент")
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
		h.Reply(chatID, "Кажется, тут пусто 🙈 Попробуй ещё раз")
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

	h.Reply(chatID, "✨ Готово! Комплимент сохранён и ждёт своего часа 💛")
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

func (h *Handler) DeleteCompliment(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	compliments, err := h.Store.GetCompliments(ctx, userID)
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

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, compliment := range compliments {
		buttonText := truncateText(compliment.Text, 30)
		callbackData := fmt.Sprintf("compliments:delete:confirm:%d", compliment.ID)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		}
		keyboard = append(keyboard, row)
	}

	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("↩️ Передумал(а)", "compliments:delete:cancel"),
	})

	text := "🗑 <b>Выбери комплимент, который хочешь убрать</b>"
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

		h.Reply(chatID, "🧹 Готово. Комплимент удалён")
	} else if data == "compliments:delete:cancel" {
		h.Reply(chatID, "🌸 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
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
		h.Reply(chatID, "🤍 Чтобы получать комплименты, сначала добавь партнёра")
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
		h.Reply(chatID, "🌙 На сегодня лимит исчерпан. Завтра будет продолжение 💛")
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

		h.Reply(chatID, fmt.Sprintf("⏳ Немного терпения\nСледующий комплимент будет доступен через %d мин.", mins))
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
		h.Reply(chatID, "📭 Пока для тебя нет новых комплиментов")
		return
	}

	compliment := compliments[0]
	err = h.Store.MarkComplimentSent(ctx, compliment.ID)
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
	text := "💛 Сегодня твой партнёр получил <b>" + countStr + "/" + actualFreqStr +
		"</b> комплимент(ов).\n\n" +
		"Хочешь изменить лимит?\n" +
		"• отправь число\n" +
		"• или «-», чтобы убрать лимит"

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
			h.Reply(chatID, "🤔 Я не понял. Отправь число или «-»")
			return
		}
	}

	err := h.Store.SetComplimentMaxCount(ctx, userID, freqInt)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при изменении лимита", err)
		return
	}

	h.Reply(chatID, "✨ Лимит обновлён")
}
