package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
)

func (h *Handler) ShowComplimentsMenu(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := "❤️ Комплименты"

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			h.HandleErr(chatID, "Error while trying to get user", err)
			return
		}
		h.HandleUnknownError(chatID, err)
		return
	}

	if user.PartnerID == 0 {
		text = "🤍 Добавь партнёра, и здесь появится магия комплиментов ✨"
	} else {
		userConfig, err := h.api.GetPartnerUserConfig(ctx, chatID)
		if err != nil || userConfig == nil {
			h.HandleErr(chatID, "Error while trying to get user config", err)
			return
		}

		count := userConfig.ComplimentCount
		maxCount := userConfig.MaxComplimentCount

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
		h.HandleErr(chatID, "An error occurred while trying to display the compliments menu.", err)
		return
	}
}

func (h *Handler) AddCompliment(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	h.sm.SetStep(state.AwaitingCompliment)
	h.Reply(chatID, "💌 Напиши комплимент")
}

func (h *Handler) ProcessCompliment(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	complimentText := msg.Text

	if complimentText == "" {
		h.sm.SetStep(state.Empty)
		h.Reply(chatID, "Кажется, тут пусто 🙈 Попробуй ещё раз")
		return
	}

	h.sm.SetStep(state.Empty)

	_, err := h.api.AddCompliment(ctx, chatID, complimentText)
	if err != nil {
		h.HandleErr(chatID, "Error adding compliment", err)
		return
	}

	h.Reply(chatID, "✨ Готово! Комплимент сохранён и ждёт своего часа 💛")
}

func (h *Handler) GetCompliments(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	var reply string

	compliments, err := h.api.GetAllCompliments(ctx, chatID)
	if err != nil || compliments == nil {
		h.HandleErr(chatID, "Error getting list of compliments", err)
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
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength-3]) + "..."
}

func (h *Handler) DeleteCompliment(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	compliments, err := h.api.GetAllCompliments(ctx, chatID)
	if err != nil || compliments == nil {
		h.HandleErr(chatID, "Error getting list of compliments", err)
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
	messageID := cb.MessageID

	if strings.HasPrefix(data, "compliments:delete:confirm:") {
		complimentIDStr := strings.TrimPrefix(data, "compliments:delete:confirm:")
		complimentID, _ := strconv.Atoi(complimentIDStr)

		err := h.api.DeleteCompliment(ctx, chatID, int64(complimentID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error deleting compliment", err)
			return
		}

		h.Reply(chatID, "🧹 Готово. Комплимент удалён")
	} else if data == "compliments:delete:cancel" {
		h.Reply(chatID, "🌸 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}

func (h *Handler) ReceiveCompliment(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	if user.PartnerID == 0 {
		h.Reply(chatID, "🤍 Чтобы получать комплименты, сначала добавь партнёра")
		return
	}

	compliment, err := h.api.ReceiveNextCompliment(ctx, chatID)
	if err != nil {
		var e *errs.ErrBucketEmpty
		switch {
		case errors.As(err, &e):
			h.Reply(chatID, fmt.Sprintf("⏳ Немного терпения\nСледующий комплимент будет доступен через %d мин.", e.Minutes))
			return
		default:
			if errors.Is(err, errs.ErrNoCompliments) {
				h.Reply(chatID, "📭 Пока для тебя нет новых комплиментов")
				return
			}
			if errors.Is(err, errs.ErrLimitExceeded) {
				h.Reply(chatID, "🌙 На сегодня лимит исчерпан. Завтра будет продолжение 💛")
				return
			}
			h.HandleErr(chatID, "Error when receiving a compliment", err)
			return
		}
	}

	var complimentMessages = []string{
		"💖 <b>Для тебя есть тёплые слова:</b>\n\n«" + compliment.Text + "»",
		"✨ <b>Небольшое послание от твоего человека:</b>\n\n«" + compliment.Text + "»",
		"🌷 <b>Тебе отправили комплимент:</b>\n\n«" + compliment.Text + "»",
	}

	randomIndex := rand.Intn(len(complimentMessages))
	h.Reply(chatID, complimentMessages[randomIndex])
	h.Reply(user.PartnerID,
		"💌 <b>Комплимент доставлен</b>\n\nТы только что порадовал(а) своего партнёра ✨\n\n«"+compliment.Text+"»",
	)
}

func (h *Handler) EditComplimentFrequency(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	userConfig, err := h.api.GetMyUserConfig(ctx, chatID)
	if err != nil || userConfig == nil {
		h.HandleErr(chatID, "Error getting user config", err)
		return
	}
	actualFreq := userConfig.MaxComplimentCount
	count := userConfig.ComplimentCount

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

	h.sm.SetStep(state.AwaitingComplimentFrequency)

	err = h.ui.EditComplimentFrequencyMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню для изменения частоты комплиментов", err)
		return
	}
}

func (h *Handler) ProcessComplimentFrequency(ctx context.Context, msg *domain.Message) {
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

	err := h.api.UpdateUserConfig(ctx, chatID, &freqInt)
	if err != nil {
		h.HandleErr(chatID, "Error changing limit", err)
		return
	}

	h.Reply(chatID, "✨ Лимит обновлён")
}
