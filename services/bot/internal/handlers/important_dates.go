package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
)

func (h *Handler) beautifyImportantDates(importantDates []domain.ImportantDate, maxLength int) []domain.ImportantDate {
	var beautifiedImportantDates []domain.ImportantDate
	var otherDates []domain.ImportantDate

	for _, importantDate := range importantDates {
		dateText := strings.Split(importantDate.Date.Format("02.01.2006"), " ")[0]
		days := strconv.Itoa(importantDate.NotifyBeforeDays)
		if importantDate.PartnerID.Valid && importantDate.UserID.Valid {
			if importantDate.IsActive {
				importantDate.Title = "👩‍❤️‍👨 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, maxLength) + " | " + dateText + " | 🟢 | " + days
			} else {
				importantDate.Title = "👩‍❤️‍👨 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, maxLength) + " | " + dateText + " | ⚪ | " + days
			}
			beautifiedImportantDates = append(beautifiedImportantDates, importantDate)
		} else {
			if importantDate.IsActive {
				importantDate.Title = "👤 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, maxLength) + " | " + dateText + " | 🟢 | " + days
			} else {
				importantDate.Title = "👤 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, maxLength) + " | " + dateText + " | ⚪ | " + days
			}
			otherDates = append(otherDates, importantDate)
		}
	}
	beautifiedImportantDates = append(beautifiedImportantDates, otherDates...)
	return beautifiedImportantDates
}

func (h *Handler) detailImportantDate(importantDate *domain.ImportantDate, maxLength int) string {
	var title string
	dateText := strings.Split(importantDate.Date.Format("02.01.2006"), " ")[0]
	days := strconv.Itoa(importantDate.NotifyBeforeDays)

	if importantDate.PartnerID.Valid && importantDate.UserID.Valid {
		if importantDate.IsActive {
			title = "👩‍❤️‍👨 | " + importantDate.Title
			title = truncateText(title, maxLength) + " | " + dateText + " | 🟢 | " + days
		} else {
			title = "👩‍❤️‍👨 | " + importantDate.Title
			title = truncateText(title, maxLength) + " | " + dateText + " | ⚪ | " + days
		}
	} else {
		if importantDate.IsActive {
			title = "👤 | " + importantDate.Title
			title = truncateText(title, maxLength) + " | " + dateText + " | 🟢 | " + days
		} else {
			title = "👤 | " + importantDate.Title
			title = truncateText(title, maxLength) + " | " + dateText + " | ⚪ | " + days
		}
	}
	return title
}

func nextOccurrence(date time.Time, now time.Time) time.Time {
	next := time.Date(
		now.Year(),
		date.Month(),
		date.Day(),
		0, 0, 0, 0,
		now.Location(),
	)

	// если в этом году уже прошло — берём следующий год
	if next.Before(now) {
		next = next.AddDate(1, 0, 0)
	}

	return next
}

func (h *Handler) nearestImportantDate(dates []domain.ImportantDate, now time.Time) (*domain.ImportantDate, bool) {
	var nearest domain.ImportantDate
	found := false
	var nearestTime time.Time

	for _, d := range dates {
		if !d.IsActive {
			continue
		}

		next := nextOccurrence(d.Date, now)

		if !found || next.Before(nearestTime) {
			nearest = d
			nearestTime = next
			found = true
		}
	}

	return &nearest, found
}

func (h *Handler) ShowImportantDatesMenu(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	var text string

	importantDates, err := h.api.GetAllImportantDates(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error while getting list of important dates", err)
		return
	}

	if len(importantDates) == 0 {
		text = "📅 Ближайших дат нет"
	} else {
		nearest, found := h.nearestImportantDate(importantDates, time.Now())
		if !found {
			text = "📅 Ближайших дат нет"
		} else {
			title := h.detailImportantDate(nearest, 256)
			text = "📅 Ближайшая важная дата: \n\n" + title
		}
	}

	err = h.ui.ImportantDatesMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Error while trying to display important dates menu", err)
		return
	}
}

func (h *Handler) AddImportantDate(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	h.sm.SetStep(state.AwaitingTitleImportantDate)

	h.Reply(chatID,
		"✍️ Как назовём эту дату?\n\n"+
			"Примеры:\n"+
			"• <b>Годовщина</b>\n"+
			"• <b>День рождения</b>\n"+
			"• <b>Первое свидание</b> 💫",
	)
}

func (h *Handler) HandleTitleImportantDate(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	title := msg.Text
	draft := domain.ImportantDateDraft{}

	draft.Title = title
	err := h.importantDateDrafts.Save(ctx, chatID, &draft)
	if err != nil {
		h.HandleErr(chatID, "Error saving the name of an important date", err)
		return
	}

	h.sm.SetStep(state.AwaitingDateImportantDate)

	err = h.ui.SendYearKeyboard(chatID, time.Now().Year(), false)
	if err != nil {
		h.HandleErr(chatID, "Error sending keyboard for year selection", err)
		return
	}
}

func (h *Handler) HandlePartnerImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	draft, err := h.importantDateDrafts.Get(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error receiving draft", err)
		return
	}
	if draft == nil {
		h.HandleErr(chatID, "Draft is empty", err)
		return
	}

	h.ui.RemoveButtons(chatID, messageID)

	switch cq.Data {
	case "important_dates:add:partner:false":
		draft.PartnerID = sql.NullInt64{Valid: false}
		err = h.importantDateDrafts.Save(ctx, chatID, draft)
		if err != nil {
			h.HandleErr(chatID, "Error saving partner important date", err)
			return
		}
	case "important_dates:add:partner:true":
		user, err := h.api.GetMe(ctx, chatID)
		if err != nil || user == nil {
			h.HandleErr(chatID, "Error getting user", err)
			return
		}

		if user.PartnerID == 0 {
			h.Reply(chatID, "У тебя пока нет партнёра 💭\nСначала добавь его, и сможете делить даты вместе")
			return
		}

		draft.PartnerID = sql.NullInt64{Int64: user.PartnerID, Valid: true}
		err = h.importantDateDrafts.Save(ctx, chatID, draft)
		if err != nil {
			h.HandleErr(chatID, "Error saving partner important date", err)
			return
		}
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.sm.SetStep(state.AwaitingNotifyBeforeImportantDate)

	err = h.ui.SendNotifyBeforeKeyboard(chatID, false)
	if err != nil {
		h.HandleErr(chatID, "Error sending keyboard to select number of days", err)
		return
	}
}

func (h *Handler) HandleNotifyBeforeImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	h.ui.RemoveButtons(chatID, messageID)

	draft, err := h.importantDateDrafts.Get(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error receiving draft", err)
		return
	}
	if draft == nil {
		h.HandleErr(chatID, "Draft is empty", err)
		return
	}

	days, err := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:notify_before:"))
	if err != nil {
		h.HandleErr(chatID, "Error converting string to number", err)
		return
	}

	draft.NotifyBeforeDays = days
	err = h.importantDateDrafts.Save(ctx, chatID, draft)
	if err != nil {
		h.HandleErr(chatID, "Error saving number of days until important date", err)
		return
	}

	finalDraft, err := h.importantDateDrafts.Get(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error receiving draft", err)
		return
	}
	if finalDraft == nil {
		h.HandleErr(chatID, "Draft is empty", err)
		return
	}

	h.sm.SetStep(state.Empty)

	err = h.importantDateDrafts.Delete(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting draft from redis", err)
		return
	}

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	date := time.Date(
		draft.Year,
		time.Month(draft.Month),
		draft.Day,
		0, 0, 0, 0,
		time.Local,
	)

	var isShared bool
	if finalDraft.PartnerID.Valid {
		isShared = true
	} else {
		isShared = false
	}

	importantDate := domain.ImportantDateRequest{
		UserID:           sql.NullInt64{Int64: user.UserID, Valid: true},
		IsShared:         isShared,
		Title:            finalDraft.Title,
		Date:             date,
		IsActive:         true,
		NotifyBeforeDays: finalDraft.NotifyBeforeDays,
	}
	_, err = h.api.AddImportantDate(ctx, chatID, importantDate)
	if err != nil {
		h.HandleErr(chatID, "Error adding important date", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	stringDate := date.Format("02.01.2006")

	h.Reply(chatID, "🎉 Важная дата добавлена!")
	if user.PartnerID != 0 && draft.PartnerID.Valid && user.PartnerID == draft.PartnerID.Int64 {
		h.Reply(user.PartnerID, "🎉 Твой партнёр добавил важную дату:\n"+"<b>"+finalDraft.Title+"</b>"+"\n"+stringDate)
	}
}

func (h *Handler) GetImportantDates(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	importantDates, err := h.api.GetAllImportantDates(ctx, chatID)
	if err != nil || importantDates == nil {
		h.HandleErr(chatID, "Error while getting list of important dates", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) важных дат… Давай создадим первую вместе! 💖")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates, 256)

	var activeImportantDates string
	var unactiveImportantDates string
	var reply string
	for _, importantDate := range sortedImportantDates {
		if importantDate.IsActive {
			activeImportantDates += "👉 " + importantDate.Title + "\n\n"
		} else {
			unactiveImportantDates += "👉 " + importantDate.Title + "\n\n"
		}
	}

	if activeImportantDates != "" {
		reply += "<b>Активные важные даты:</b>\n\n" + activeImportantDates
	}
	if unactiveImportantDates != "" {
		reply += "<b>Неактивные важные даты:</b>\n\n" + unactiveImportantDates + "\n"
	}

	h.Reply(chatID, reply)
}

func (h *Handler) DeleteImportantDate(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	importantDates, err := h.api.GetAllImportantDates(ctx, chatID)
	if err != nil || importantDates == nil {
		h.HandleErr(chatID, "Error while getting list of important dates", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) важных дат… Давай создадим первую вместе! 💖")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates, 30)

	var rows []domain.InlineKeyboardRow

	for _, importantDate := range sortedImportantDates {
		callbackData := fmt.Sprintf("important_dates:delete:confirm:%d", importantDate.ID)

		row := domain.InlineKeyboardRow{
			Buttons: []domain.InlineKeyboardButton{
				{Text: importantDate.Title, Data: callbackData},
			},
		}
		rows = append(rows, row)
	}

	rows = append(rows, domain.InlineKeyboardRow{
		Buttons: []domain.InlineKeyboardButton{
			{Text: "❌ Ой, передумал(а)", Data: "important_dates:delete:cancel"},
		},
	})

	text := "🗑 Выбери, какую дату мы удалим"
	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
	if err != nil {
		h.HandleErr(chatID, "Error sending confirmation", err)
		return
	}
}

func (h *Handler) HandleDeleteImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	data := cq.Data
	chatID := cq.ChatID
	messageID := cq.MessageID

	if strings.HasPrefix(data, "important_dates:delete:confirm") {
		importantDateIDStr := strings.TrimPrefix(data, "important_dates:delete:confirm:")
		importantDateID, _ := strconv.Atoi(importantDateIDStr)

		user, err := h.api.GetMe(ctx, chatID)
		if err != nil || user == nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error getting user", err)
			return
		}

		importantDate, err := h.api.GetImportantDate(ctx, chatID, int64(importantDateID))
		if err != nil || importantDate == nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error getting important date", err)
			return
		}

		title := importantDate.Title
		date := importantDate.Date.Format("02.01.2006")

		err = h.api.DeleteImportantDate(ctx, chatID, importantDate.ID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error deleting important date", err)
			return
		}

		h.Reply(chatID, "✅ Готово! Важная дата удалена")

		partnerID := user.PartnerID
		if (partnerID != 0 && importantDate.PartnerID.Valid && importantDate.PartnerID.Int64 == partnerID) ||
			(partnerID != 0 && importantDate.UserID.Valid && importantDate.UserID.Int64 == partnerID) {
			h.Reply(partnerID, "💔 Твой партнёр удалил важную дату:\n"+"<b>"+title+"</b>"+"\n"+date)
		}
	} else if strings.HasPrefix(data, "important_dates:delete:cancel") {
		h.Reply(chatID, "😉 Удаление отменено")
	} else {
		h.Reply(chatID, "😢 Что-то пошло не так…")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}

func (h *Handler) EditImportantDate(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	importantDates, err := h.api.GetAllImportantDates(ctx, chatID)
	if err != nil || importantDates == nil {
		h.HandleErr(chatID, "Error while getting list of important dates", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) важных дат… Давай создадим первую вместе! 💖")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates, 30)

	var rows []domain.InlineKeyboardRow

	for _, importantDate := range sortedImportantDates {
		callbackData := fmt.Sprintf("important_dates:update_menu:%d", importantDate.ID)

		row := domain.InlineKeyboardRow{
			Buttons: []domain.InlineKeyboardButton{
				{Text: importantDate.Title, Data: callbackData},
			},
		}
		rows = append(rows, row)
	}

	rows = append(rows, domain.InlineKeyboardRow{
		Buttons: []domain.InlineKeyboardButton{
			{Text: "❌ Отмена", Data: "important_dates:update_menu:cancel"},
		},
	})

	text := "🌸 Выбери дату, которую хочешь изменить"
	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
	if err != nil {
		h.HandleErr(chatID, "Error sending confirmation", err)
		return
	}
}

func (h *Handler) HandleEditImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	data := cq.Data
	chatID := cq.ChatID
	messageID := cq.MessageID

	data = strings.TrimPrefix(data, "important_dates:update_menu:")
	if data == "cancel" {
		h.Reply(chatID, "😉 Редактирование отменено")
	} else {
		id, _ := strconv.Atoi(data)

		importantDate, err := h.api.GetImportantDate(ctx, chatID, int64(id))
		if err != nil || importantDate == nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error getting important date", err)
			return
		}

		var active string
		if importantDate.IsActive {
			active = "Деактивировать 💤"
		} else {
			active = "Активировать ✨"
		}

		keyboard := domain.InlineKeyboard{
			Rows: []domain.InlineKeyboardRow{
				{
					Buttons: []domain.InlineKeyboardButton{
						{Text: "Название 📝", Data: "important_dates:update:title:" + data},
						{Text: "Дата 📅", Data: "important_dates:update:date:" + data},
					},
				},
				{
					Buttons: []domain.InlineKeyboardButton{
						{Text: "Партнёр 💑", Data: "important_dates:update:partner:" + data},
						{Text: "Уведомлять за ⏰", Data: "important_dates:update:notify_before:" + data},
					},
				},
				{
					Buttons: []domain.InlineKeyboardButton{
						{Text: active, Data: "important_dates:update:is_active:" + data},
						{Text: "❌ Отмена", Data: "important_dates:update:cancel"},
					},
				},
			},
		}

		title := h.detailImportantDate(importantDate, 256)
		text := "💌 Что хочешь изменить?\n\n" + title

		err = h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
		if err != nil {
			h.HandleErr(chatID, "Error sending confirmation", err)
			return
		}
	}
	err := h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}
}

func (h *Handler) CancelCallbackImportantDate(_ context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	err := h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.Reply(chatID, "😉 Действие отменено")
}

func (h *Handler) HandleEditTitleImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:title:"))

	err := h.importantDateEditDrafts.Save(ctx, chatID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Error saving editing session", err)
		return
	}

	h.sm.SetStep(state.AwaitingEditTitleImportantDate)

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.Reply(chatID, "✍️ Введи новое название памятной даты")
}

func (h *Handler) HandleEditTitleImportantDateText(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Editing session expired", err)
		return
	}

	date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
	if err != nil || date == nil {
		h.HandleErr(chatID, "Error getting important date", err)
		return
	}

	var isShared bool
	if date.PartnerID.Valid {
		isShared = true
	} else {
		isShared = false
	}

	req := domain.ImportantDateRequest{
		UserID:           date.UserID,
		IsShared:         isShared,
		Title:            msg.Text,
		Date:             date.Date,
		IsActive:         date.IsActive,
		NotifyBeforeDays: date.NotifyBeforeDays,
	}
	err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
	if err != nil {
		h.HandleErr(chatID, "Error updating important date", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, chatID)

	h.sm.SetStep(state.Empty)

	h.Reply(chatID, "✅ Отлично! Название обновлено")
}

func (h *Handler) HandleEditDateImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:date:"))

	err := h.importantDateEditDrafts.Save(ctx, chatID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Error saving editing session", err)
		return
	}

	h.sm.SetStep(state.AwaitingEditDateImportantDate)

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	err = h.ui.SendYearKeyboard(chatID, time.Now().Year(), true)
	if err != nil {
		h.HandleErr(chatID, "Error sending keyboard for year selection", err)
		return
	}
}

func (h *Handler) HandleEditPartnerImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:partner:"))

	err := h.importantDateEditDrafts.Save(ctx, chatID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Error saving session", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Error deleting message", er)
	}

	err = h.ui.SendPartnerKeyboard(chatID, true)
	if err != nil {
		h.HandleErr(chatID, "Error sending keyboard for partner selection on important date", err)
		return
	}
}

func (h *Handler) HandleEditPartnerImportantDateSelect(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Session expired", err)
		return
	}

	date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
	if err != nil || date == nil {
		h.HandleErr(chatID, "Date not found", err)
		return
	}

	var isShared bool
	switch cq.Data {
	case "important_dates:edit:partner:false":
		isShared = false
	case "important_dates:edit:partner:true":
		isShared = true
	}

	req := domain.ImportantDateRequest{
		UserID:           date.UserID,
		IsShared:         isShared,
		Title:            date.Title,
		Date:             date.Date,
		IsActive:         date.IsActive,
		NotifyBeforeDays: date.NotifyBeforeDays,
	}
	err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
	if err != nil {
		h.HandleErr(chatID, "Error updating important date", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, chatID)

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.Reply(chatID, "👥 Партнёр успешно обновлён")
}

func (h *Handler) HandleEditNotifyBeforeImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:notify_before:"))

	err := h.importantDateEditDrafts.Save(ctx, chatID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Error saving session", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	err = h.ui.SendNotifyBeforeKeyboard(chatID, true)
	if err != nil {
		h.HandleErr(chatID, "Error sending keyboard to select number of days", err)
		return
	}
}

func (h *Handler) HandleEditNotifyBeforeImportantDateSelect(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Session expired", err)
		return
	}

	days, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:edit:notify_before:"))

	date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
	if err != nil || date == nil {
		h.HandleErr(chatID, "Date not found", err)
		return
	}

	var isShared bool
	if date.PartnerID.Valid {
		isShared = true
	} else {
		isShared = false
	}

	req := domain.ImportantDateRequest{
		UserID:           date.UserID,
		IsShared:         isShared,
		Title:            date.Title,
		Date:             date.Date,
		IsActive:         date.IsActive,
		NotifyBeforeDays: days,
	}
	err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
	if err != nil {
		h.HandleErr(chatID, "Error updating important date", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, chatID)

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.Reply(chatID, "⏰ Уведомления успешно обновлены")
}

func (h *Handler) HandleEditIsActiveImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:is_active:"))

	date, err := h.api.GetImportantDate(ctx, chatID, int64(id))
	if err != nil || date == nil {
		h.HandleErr(chatID, "Date not found", err)
		return
	}

	var isShared bool
	if date.PartnerID.Valid {
		isShared = true
	} else {
		isShared = false
	}

	req := domain.ImportantDateRequest{
		UserID:           date.UserID,
		IsShared:         isShared,
		Title:            date.Title,
		Date:             date.Date,
		IsActive:         !date.IsActive,
		NotifyBeforeDays: date.NotifyBeforeDays,
	}
	err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
	if err != nil {
		h.HandleErr(chatID, "Error updating important date", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	if !date.IsActive {
		h.Reply(chatID, "🟢 Дата теперь активна")
	} else {
		h.Reply(chatID, "⚪ Дата деактивирована")
	}
}

func (h *Handler) HandleYearImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID
	data := cq.Data

	// Определяем flow: add или edit
	var isEdit bool
	if strings.HasPrefix(data, "important_dates:edit:") {
		isEdit = true
		data = strings.TrimPrefix(data, "important_dates:edit:")
	} else {
		data = strings.TrimPrefix(data, "important_dates:add:")
	}

	// Пагинация
	if strings.HasPrefix(data, "year:page:") {
		startYear, _ := strconv.Atoi(strings.TrimPrefix(data, "year:page:"))
		keyboard := h.ui.BuildYearKeyboard(startYear, isEdit)
		err := h.ui.Client.EditMessageReplyMarkup(chatID, messageID, keyboard)
		if err != nil {
			h.HandleErr(chatID, "Error updating keyboard", err)
		}
		return
	}

	// Выбор конкретного года
	if strings.HasPrefix(data, "year:select:") {
		year, _ := strconv.Atoi(strings.TrimPrefix(data, "year:select:"))

		if isEdit {
			// Редактируем дату
			draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Сессия редактирования истекла", err)
				return
			}

			date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
			if err != nil || date == nil {
				h.HandleErr(chatID, "Date not found", err)
				return
			}

			newDate := time.Date(year, date.Date.Month(), date.Date.Day(), 0, 0, 0, 0, time.Local)

			var isShared bool
			if date.PartnerID.Valid {
				isShared = true
			} else {
				isShared = false
			}

			req := domain.ImportantDateRequest{
				UserID:           date.UserID,
				IsShared:         isShared,
				Title:            date.Title,
				Date:             newDate,
				IsActive:         date.IsActive,
				NotifyBeforeDays: date.NotifyBeforeDays,
			}
			err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
			if err != nil {
				h.HandleErr(chatID, "Error updating important date", err)
				return
			}

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			err = h.ui.SendMonthKeyboard(chatID, isEdit)
			if err != nil {
				h.HandleErr(chatID, "Error sending keyboard for month selection", err)
			}

		} else {
			// Добавляем дату
			draft, err := h.importantDateDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Draft is empty", err)
				return
			}

			draft.Year = year
			err = h.importantDateDrafts.Save(ctx, chatID, draft)
			if err != nil {
				h.HandleErr(chatID, "Error saving year", err)
				return
			}

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			err = h.ui.SendMonthKeyboard(chatID, isEdit)
			if err != nil {
				h.HandleErr(chatID, "Error sending keyboard for month selection", err)
			}
		}

		return
	}

	h.HandleErr(chatID, "Unknown callback for the year", nil)
}

func (h *Handler) HandleMonthImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID
	data := cq.Data

	var isEdit bool
	if strings.HasPrefix(data, "important_dates:edit:") {
		isEdit = true
		data = strings.TrimPrefix(data, "important_dates:edit:")
	} else {
		data = strings.TrimPrefix(data, "important_dates:add:")
	}

	if strings.HasPrefix(data, "month:") {
		month, _ := strconv.Atoi(strings.TrimPrefix(data, "month:"))

		if isEdit {
			draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Editing session expired", err)
				return
			}
			date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
			if err != nil || date == nil {
				h.HandleErr(chatID, "Date not found", err)
				return
			}

			newDate := time.Date(date.Date.Year(), time.Month(month), date.Date.Day(), 0, 0, 0, 0, time.Local)

			var isShared bool
			if date.PartnerID.Valid {
				isShared = true
			} else {
				isShared = false
			}

			req := domain.ImportantDateRequest{
				UserID:           date.UserID,
				IsShared:         isShared,
				Title:            date.Title,
				Date:             newDate,
				IsActive:         date.IsActive,
				NotifyBeforeDays: date.NotifyBeforeDays,
			}
			err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
			if err != nil {
				h.HandleErr(chatID, "Error updating important date", err)
				return
			}

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			err = h.ui.SendDayKeyboard(chatID, date.Date.Year(), month, isEdit)
			if err != nil {
				h.HandleErr(chatID, "Error sending keyboard to select day", err)
			}
		} else {
			draft, err := h.importantDateDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Draft is empty", err)
				return
			}
			draft.Month = month
			err = h.importantDateDrafts.Save(ctx, chatID, draft)
			if err != nil {
				h.HandleErr(chatID, "Error saving month", err)
				return
			}

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			err = h.ui.SendDayKeyboard(chatID, draft.Year, month, isEdit)
			if err != nil {
				h.HandleErr(chatID, "Error sending keyboard to select day", err)
			}
		}
		return
	}

	h.HandleErr(chatID, "Unknown callback for month", nil)
}

func (h *Handler) HandleDayImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID
	data := cq.Data

	var isEdit bool
	if strings.HasPrefix(data, "important_dates:edit:") {
		isEdit = true
		data = strings.TrimPrefix(data, "important_dates:edit:")
	} else {
		data = strings.TrimPrefix(data, "important_dates:add:")
	}

	if strings.HasPrefix(data, "day:") {
		day, _ := strconv.Atoi(strings.TrimPrefix(data, "day:"))

		if isEdit {
			draft, err := h.importantDateEditDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Editing session expired", err)
				return
			}

			date, err := h.api.GetImportantDate(ctx, chatID, draft.ImportantDateID)
			if err != nil || date == nil {
				h.HandleErr(chatID, "Date not found", err)
				return
			}

			newDate := time.Date(date.Date.Year(), date.Date.Month(), day, 0, 0, 0, 0, time.Local)

			var isShared bool
			if date.PartnerID.Valid {
				isShared = true
			} else {
				isShared = false
			}

			req := domain.ImportantDateRequest{
				UserID:           date.UserID,
				IsShared:         isShared,
				Title:            date.Title,
				Date:             newDate,
				IsActive:         date.IsActive,
				NotifyBeforeDays: date.NotifyBeforeDays,
			}
			err = h.api.UpdateImportantDate(ctx, chatID, date.ID, req)
			if err != nil {
				h.HandleErr(chatID, "Error updating important date", err)
				return
			}

			err = h.importantDateEditDrafts.Delete(ctx, chatID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting date", err)
				return
			}

			h.sm.SetStep(state.Empty)

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			h.Reply(chatID, "📅 Дата успешно обновлена")
		} else {
			draft, err := h.importantDateDrafts.Get(ctx, chatID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Draft is empty", err)
				return
			}

			draft.Day = day
			err = h.importantDateDrafts.Save(ctx, chatID, draft)
			if err != nil {
				h.HandleErr(chatID, "Error saving day", err)
				return
			}

			err = h.ui.Client.DeleteMessage(chatID, messageID)
			if err != nil {
				h.HandleErr(chatID, "Error deleting message", err)
			}

			// Далее переход к выбору партнера / уведомлений
			user, err := h.api.GetMe(ctx, chatID)
			if err != nil || user == nil {
				h.HandleErr(chatID, "Error getting me", err)
				return
			}
			partnerID := user.PartnerID

			if partnerID == 0 {
				h.Reply(chatID, "✨ Так как у тебя пока нет партнёра, памятная дата будет твоей личной")
				h.sm.SetStep(state.AwaitingNotifyBeforeImportantDate)
				err = h.ui.SendNotifyBeforeKeyboard(chatID, isEdit)
				if err != nil {
					h.HandleErr(chatID, "Error sending keyboard to select day", err)
					return
				}
			} else {
				h.sm.SetStep(state.AwaitingPartnerImportantDate)
				err = h.ui.SendPartnerKeyboard(chatID, isEdit)
				if err != nil {
					h.HandleErr(chatID, "Error sending keyboard to select day", err)
					return
				}
			}
		}
		return
	}

	h.HandleErr(chatID, "Unknown callback for the day", nil)
}
