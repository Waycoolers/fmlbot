package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (h *Handler) beautifyImportantDates(importantDates []domain.ImportantDate, maxLength int) []domain.ImportantDate {
	var beautifiedImportantDates []domain.ImportantDate
	var otherDates []domain.ImportantDate

	for _, importantDate := range importantDates {
		dateText := strings.Split(importantDate.Date.Format("02.01.2006"), " ")[0]
		days := strconv.Itoa(importantDate.NotifyBeforeDays)
		if importantDate.PartnerID.Valid && importantDate.TelegramID.Valid {
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

func (h *Handler) detailImportantDate(importantDate domain.ImportantDate, maxLength int) string {
	var title string
	dateText := strings.Split(importantDate.Date.Format("02.01.2006"), " ")[0]
	days := strconv.Itoa(importantDate.NotifyBeforeDays)

	if importantDate.PartnerID.Valid && importantDate.TelegramID.Valid {
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

func (h *Handler) nearestImportantDate(dates []domain.ImportantDate, now time.Time) (domain.ImportantDate, bool) {
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

	return nearest, found
}

func (h *Handler) ShowImportantDatesMenu(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	userID := msg.UserID
	var text string

	importantDates, err := h.Store.ImportantDates.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
		return
	}

	if len(importantDates) == 0 {
		text = "📅 Ближайших дат нет"
	} else {
		nearest, found := h.nearestImportantDate(importantDates, time.Now())
		if !found {
			text = "📅 Ближайших дат нет"
		}
		title := h.detailImportantDate(nearest, 256)
		text = "📅 Ближайшая важная дата: \n\n" + title
	}

	err = h.ui.ImportantDatesMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню важных дат", err)
		return
	}
}

func (h *Handler) AddImportantDate(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	userID := msg.UserID

	err := h.Store.Users.SetUserState(ctx, userID, domain.AwaitingTitleImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}
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
	userID := msg.UserID
	title := msg.Text
	draft := domain.ImportantDateDraft{}

	draft.Title = title
	err := h.importantDateDrafts.Save(ctx, userID, &draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении названия важной даты", err)
		return
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingDateImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	err = h.ui.SendYearKeyboard(chatID, time.Now().Year(), false)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора года", err)
		return
	}
}

func (h *Handler) HandlePartnerImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	draft, err := h.importantDateDrafts.Get(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении черновика", err)
		return
	}
	if draft == nil {
		h.HandleErr(chatID, "Черновик пустой", err)
		return
	}

	h.ui.RemoveButtons(chatID, messageID)

	switch cq.Data {
	case "important_dates:add:partner:false":
		draft.PartnerID = sql.NullInt64{Valid: false}
		err = h.importantDateDrafts.Save(ctx, userID, draft)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сохранении партнера важной даты", err)
			return
		}
	case "important_dates:add:partner:true":
		partnerID, er := h.Store.Users.GetPartnerID(ctx, userID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при получении id партнера", er)
			return
		}

		if partnerID == 0 {
			h.Reply(chatID, "У тебя пока нет партнёра 💭\nСначала добавь его, и сможете делить даты вместе")
			return
		}

		draft.PartnerID = sql.NullInt64{Int64: partnerID, Valid: true}
		err = h.importantDateDrafts.Save(ctx, userID, draft)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сохранении партнера важной даты", err)
			return
		}
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	err = h.ui.SendNotifyBeforeKeyboard(chatID, false)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора количества дней", err)
		return
	}
}

func (h *Handler) HandleNotifyBeforeImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	h.ui.RemoveButtons(chatID, messageID)

	draft, err := h.importantDateDrafts.Get(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении черновика", err)
		return
	}
	if draft == nil {
		h.HandleErr(chatID, "Черновик пустой", err)
		return
	}

	days, err := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:notify_before:"))
	if err != nil {
		h.HandleErr(chatID, "Ошибка преобразования строки в число", err)
		return
	}

	draft.NotifyBeforeDays = days
	err = h.importantDateDrafts.Save(ctx, userID, draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении количества дней до важной даты", err)
		return
	}

	finalDraft, err := h.importantDateDrafts.Get(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении черновика", err)
		return
	}
	if finalDraft == nil {
		h.HandleErr(chatID, "Черновик пустой", err)
		return
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	err = h.importantDateDrafts.Delete(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении черновика из redis", err)
		return
	}

	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	date := time.Date(
		draft.Year,
		time.Month(draft.Month),
		draft.Day,
		0, 0, 0, 0,
		time.Local,
	)

	_, err = h.Store.ImportantDates.AddImportantDate(ctx, sql.NullInt64{Int64: userID, Valid: true}, finalDraft.PartnerID, finalDraft.Title,
		date, finalDraft.NotifyBeforeDays)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении важной даты", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
	}

	stringDate := date.Format("02.01.2006")

	h.Reply(chatID, "🎉 Важная дата добавлена!")
	if partnerID != 0 && draft.PartnerID.Valid && partnerID == draft.PartnerID.Int64 {
		h.Reply(partnerID, "🎉 Твой партнёр добавил важную дату:\n"+"<b>"+finalDraft.Title+"</b>"+"\n"+stringDate)
	}
}

func (h *Handler) GetImportantDates(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	userID := msg.UserID

	importantDates, err := h.Store.ImportantDates.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
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
	userID := msg.UserID

	importantDates, err := h.Store.ImportantDates.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
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
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeleteImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	data := cq.Data
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	if strings.HasPrefix(data, "important_dates:delete:confirm") {
		importantDateIDStr := strings.TrimPrefix(data, "important_dates:delete:confirm:")
		importantDateID, _ := strconv.Atoi(importantDateIDStr)

		partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при получении id партнера", err)
			return
		}

		importantDate, err := h.Store.ImportantDates.GetImportantDateByID(ctx, int64(importantDateID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при получении важной даты", err)
			return
		}

		title := importantDate.Title
		date := importantDate.Date.Format("02.01.2006")

		err = h.Store.ImportantDates.DeleteImportantDate(ctx, int64(importantDateID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при удалении важной даты", err)
			return
		}

		h.Reply(chatID, "✅ Готово! Важная дата удалена")

		if (partnerID != 0 && importantDate.PartnerID.Valid && importantDate.PartnerID.Int64 == partnerID) ||
			(partnerID != 0 && importantDate.TelegramID.Valid && importantDate.TelegramID.Int64 == partnerID) {
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
	userID := msg.UserID

	importantDates, err := h.Store.ImportantDates.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
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
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
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

		importantDate, err := h.Store.ImportantDates.GetImportantDateByID(ctx, int64(id))
		if err != nil {
			h.HandleErr(chatID, "Ошибка при получении важной даты", err)
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
			h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
			return
		}
	}
	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}
}

func (h *Handler) CancelCallbackImportantDate(_ context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "😉 Действие отменено")
}

func (h *Handler) HandleEditTitleImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:title:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии редактирования", err)
		return
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingEditTitleImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "✍️ Введи новое название памятной даты")
}

func (h *Handler) HandleEditTitleImportantDateText(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	userID := msg.UserID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия редактирования истекла", err)
		return
	}

	date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.Title = msg.Text

	err = h.Store.ImportantDates.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении названия", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)
	err = h.Store.Users.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	h.Reply(chatID, "✅ Отлично! Название обновлено")
}

func (h *Handler) HandleEditDateImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:date:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии редактирования", err)
		return
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingEditDateImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	err = h.ui.SendYearKeyboard(chatID, time.Now().Year(), true)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора года", err)
		return
	}
}

func (h *Handler) HandleEditPartnerImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:partner:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	err = h.ui.SendPartnerKeyboard(chatID, true)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора партнера в важной дате", err)
		return
	}
}

func (h *Handler) HandleEditPartnerImportantDateSelect(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия истекла", err)
		return
	}

	date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	switch cq.Data {
	case "important_dates:edit:partner:false":
		date.PartnerID = sql.NullInt64{Valid: false}
	case "important_dates:edit:partner:true":
		partnerID, _ := h.Store.Users.GetPartnerID(ctx, userID)
		date.PartnerID = sql.NullInt64{Int64: partnerID, Valid: true}
	}

	err = h.Store.ImportantDates.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении партнёра", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "👥 Партнёр успешно обновлён")
}

func (h *Handler) HandleEditNotifyBeforeImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:notify_before:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	err = h.ui.SendNotifyBeforeKeyboard(chatID, true)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора количества дней", err)
		return
	}
}

func (h *Handler) HandleEditNotifyBeforeImportantDateSelect(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
	messageID := cq.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия истекла", err)
		return
	}

	days, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:edit:notify_before:"))

	date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.NotifyBeforeDays = days

	err = h.Store.ImportantDates.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении уведомлений", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "⏰ Уведомления успешно обновлены")
}

func (h *Handler) HandleEditIsActiveImportantDate(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:is_active:"))

	date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, int64(id))
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.IsActive = !date.IsActive

	err = h.Store.ImportantDates.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении активности", err)
		return
	}

	_ = h.ui.Client.DeleteMessage(chatID, messageID)

	if date.IsActive {
		h.Reply(chatID, "🟢 Дата теперь активна")
	} else {
		h.Reply(chatID, "⚪ Дата деактивирована")
	}
}

func (h *Handler) HandleYearImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
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
			h.HandleErr(chatID, "Ошибка при обновлении клавиатуры", err)
		}
		return
	}

	// Выбор конкретного года
	if strings.HasPrefix(data, "year:select:") {
		year, _ := strconv.Atoi(strings.TrimPrefix(data, "year:select:"))

		if isEdit {
			// Редактируем дату
			draft, err := h.importantDateEditDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Сессия редактирования истекла", err)
				return
			}

			date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}

			date.Date = time.Date(year, date.Date.Month(), date.Date.Day(), 0, 0, 0, 0, time.Local)
			if er := h.Store.ImportantDates.EditImportantDate(ctx, date); er != nil {
				h.HandleErr(chatID, "Ошибка при обновлении года", er)
				return
			}

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			if er := h.ui.SendMonthKeyboard(chatID, isEdit); er != nil {
				h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора месяца", er)
			}

		} else {
			// Добавляем дату
			draft, err := h.importantDateDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Черновик пустой", err)
				return
			}

			draft.Year = year
			if er := h.importantDateDrafts.Save(ctx, userID, draft); er != nil {
				h.HandleErr(chatID, "Ошибка при сохранении года", er)
				return
			}

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			if er := h.ui.SendMonthKeyboard(chatID, isEdit); er != nil {
				h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора месяца", er)
			}
		}

		return
	}

	h.HandleErr(chatID, "Неизвестный callback для года", nil)
}

func (h *Handler) HandleMonthImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
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
			draft, err := h.importantDateEditDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Сессия редактирования истекла", err)
				return
			}
			date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}
			date.Date = time.Date(date.Date.Year(), time.Month(month), date.Date.Day(), 0, 0, 0, 0, time.Local)
			if er := h.Store.ImportantDates.EditImportantDate(ctx, date); er != nil {
				h.HandleErr(chatID, "Ошибка при обновлении месяца", er)
				return
			}

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			if er := h.ui.SendDayKeyboard(chatID, date.Date.Year(), month, isEdit); er != nil {
				h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора дня", er)
			}
		} else {
			draft, err := h.importantDateDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Черновик пустой", err)
				return
			}
			draft.Month = month
			if er := h.importantDateDrafts.Save(ctx, userID, draft); er != nil {
				h.HandleErr(chatID, "Ошибка при сохранении месяца", er)
				return
			}

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			if er := h.ui.SendDayKeyboard(chatID, draft.Year, month, isEdit); er != nil {
				h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора дня", er)
			}
		}
		return
	}

	h.HandleErr(chatID, "Неизвестный callback для месяца", nil)
}

func (h *Handler) HandleDayImportantDateUniversal(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	userID := cq.UserID
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
			draft, err := h.importantDateEditDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Сессия редактирования истекла", err)
				return
			}

			date, err := h.Store.ImportantDates.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}

			date.Date = time.Date(date.Date.Year(), date.Date.Month(), day, 0, 0, 0, 0, time.Local)
			if er := h.Store.ImportantDates.EditImportantDate(ctx, date); er != nil {
				h.HandleErr(chatID, "Ошибка при обновлении дня", er)
				return
			}

			_ = h.importantDateEditDrafts.Delete(ctx, userID)
			_ = h.Store.Users.SetUserState(ctx, userID, domain.Empty)

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			h.Reply(chatID, "📅 Дата успешно обновлена")
		} else {
			draft, err := h.importantDateDrafts.Get(ctx, userID)
			if err != nil || draft == nil {
				h.HandleErr(chatID, "Черновик пустой", err)
				return
			}

			draft.Day = day
			if er := h.importantDateDrafts.Save(ctx, userID, draft); er != nil {
				h.HandleErr(chatID, "Ошибка при сохранении дня", er)
				return
			}

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			// Далее переход к выбору партнера / уведомлений
			partnerID, er := h.Store.Users.GetPartnerID(ctx, userID)
			if er != nil {
				h.HandleErr(chatID, "Ошибка при получении id партнера", er)
				return
			}

			if partnerID == 0 {
				h.Reply(chatID, "✨ Так как у тебя пока нет партнёра, памятная дата будет твоей личной")
				_ = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
				_ = h.ui.SendNotifyBeforeKeyboard(chatID, isEdit)
			} else {
				_ = h.Store.Users.SetUserState(ctx, userID, domain.AwaitingPartnerImportantDate)
				_ = h.ui.SendPartnerKeyboard(chatID, isEdit)
			}
		}
		return
	}

	h.HandleErr(chatID, "Неизвестный callback для дня", nil)
}

func (h *Handler) NotifyImportantDatesCron(ctx context.Context) {
	now := time.Now()
	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.Local,
	)

	importantDates, err := h.Store.ImportantDates.GetAllActiveImportantDates(ctx)
	if err != nil {
		log.Println("Ошибка получения всех важных дат:", err)
		return
	}

	for _, importantDate := range importantDates {
		if !importantDate.IsActive {
			continue
		}

		eventDate := importantDate.Date.In(time.Local)
		eventDay := time.Date(
			eventDate.Year(),
			eventDate.Month(),
			eventDate.Day(),
			0, 0, 0, 0,
			time.Local,
		)

		notifyDay := eventDay.AddDate(0, 0, -importantDate.NotifyBeforeDays)

		isNotifyDay := notifyDay.Equal(today)
		isEventDay := eventDay.Equal(today)

		if !isNotifyDay && !isEventDay {
			continue
		}

		if importantDate.LastNotificationAt.Valid {
			last := importantDate.LastNotificationAt.Time.In(time.Local)
			lastDay := time.Date(
				last.Year(),
				last.Month(),
				last.Day(),
				0, 0, 0, 0,
				time.Local,
			)
			if lastDay.Equal(today) {
				continue
			}
		}

		var text string
		if isEventDay {
			text = fmt.Sprintf("🎉 Ура! Сегодня важная дата!\n\n<b>%s</b>\n%s",
				importantDate.Title,
				eventDate.Format("02.01.2006"),
			)
		} else {
			text = fmt.Sprintf(
				"⏰ Напоминание: через %d дн.\n\n<b>%s</b>\n%s",
				importantDate.NotifyBeforeDays,
				importantDate.Title,
				eventDate.Format("02.01.2006"),
			)
		}

		if importantDate.TelegramID.Valid && importantDate.TelegramID.Int64 != 0 {
			h.Reply(importantDate.TelegramID.Int64, text)
		}
		if importantDate.PartnerID.Valid && importantDate.PartnerID.Int64 != 0 {
			h.Reply(importantDate.PartnerID.Int64, text)
		}

		_ = h.Store.ImportantDates.UpdateLastNotificationAt(ctx, importantDate.ID, now)
	}
}
