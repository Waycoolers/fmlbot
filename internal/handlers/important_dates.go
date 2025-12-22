package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) beautifyImportantDates(importantDates []domain.ImportantDate) []domain.ImportantDate {
	var beautifiedImportantDates []domain.ImportantDate
	var otherDates []domain.ImportantDate

	for _, importantDate := range importantDates {
		dateText := strings.Split(importantDate.Date.Format("02.01.2006"), " ")[0]
		days := strconv.Itoa(importantDate.NotifyBeforeDays)
		if importantDate.PartnerID.Valid && importantDate.TelegramID.Valid {
			if importantDate.IsActive {
				importantDate.Title = "👩‍❤️‍👨 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, 30) + " | " + dateText + " | 🟢 | " + days
			} else {
				importantDate.Title = "👩‍❤️‍👨 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, 30) + " | " + dateText + " | ⚪ | " + days
			}
			beautifiedImportantDates = append(beautifiedImportantDates, importantDate)
		} else {
			if importantDate.IsActive {
				importantDate.Title = "👤 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, 30) + " | " + dateText + " | 🟢 | " + days
			} else {
				importantDate.Title = "👤 | " + importantDate.Title
				importantDate.Title = truncateText(importantDate.Title, 30) + " | " + dateText + " | ⚪ | " + days
			}
			otherDates = append(otherDates, importantDate)
		}
	}
	beautifiedImportantDates = append(beautifiedImportantDates, otherDates...)
	return beautifiedImportantDates
}

func (h *Handler) ShowImportantDatesMenu(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := "Важные даты"

	err := h.ui.ImportantDatesMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню важных дат", err)
		return
	}
}

func (h *Handler) AddImportantDate(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	err := h.Store.SetUserState(ctx, userID, domain.AwaitingTitleImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}
	h.Reply(chatID,
		"✍️ Как называется памятная дата?\n"+
			"Например: <b>Годовщина</b>, <b>Твой день рождения</b>, <b>Первое свидание</b>",
	)
}

func (h *Handler) HandleTitleImportantDate(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	title := msg.Text
	draft := domain.ImportantDateDraft{}

	draft.Title = title
	err := h.importantDateDrafts.Save(ctx, userID, &draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении названия важной даты", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingDateImportantDate)
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

func (h *Handler) HandlePartnerImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

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
		partnerID, er := h.Store.GetPartnerID(ctx, userID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при получении id партнера", er)
			return
		}

		if partnerID == 0 {
			h.Reply(chatID, "У тебя не добавлен партнёр. Сначала добавь его")
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

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
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

func (h *Handler) HandleNotifyBeforeImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

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

	err = h.Store.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	err = h.importantDateDrafts.Delete(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении черновика из redis", err)
		return
	}

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
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

	_, err = h.Store.AddImportantDate(ctx, sql.NullInt64{Int64: userID, Valid: true}, finalDraft.PartnerID, finalDraft.Title,
		date, finalDraft.NotifyBeforeDays)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении важной даты", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
	}

	h.Reply(chatID, "Памятная дата добавлена")
	if partnerID != 0 && draft.PartnerID.Valid {
		h.Reply(partnerID, "Твой партнёр добавил памятную дату:\n"+finalDraft.Title)
	}
}

func (h *Handler) GetImportantDates(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	importantDates, err := h.Store.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "Ты пока не добавлял(а) важных дат. Добавь важную дату")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates)

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

func (h *Handler) DeleteImportantDate(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	importantDates, err := h.Store.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "У тебя не добавлены важные даты")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates)

	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, importantDate := range sortedImportantDates {
		callbackData := fmt.Sprintf("important_dates:delete:confirm:%d", importantDate.ID)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(importantDate.Title, callbackData),
		}
		buttons = append(buttons, row)
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "important_dates:delete:cancel"),
	})

	text := "🗑 <b>Выбери важную дату для удаления</b>"
	markup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, markup)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeleteImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	data := cq.Data
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	if strings.HasPrefix(data, "important_dates:delete:confirm") {
		importantDateIDStr := strings.TrimPrefix(data, "important_dates:delete:confirm:")
		importantDateID, _ := strconv.Atoi(importantDateIDStr)

		err := h.Store.DeleteImportantDate(ctx, int64(importantDateID))
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при удалении важной даты", err)
			return
		}

		h.Reply(chatID, "Важная дата успешно удалена! ✅")
	} else if strings.HasPrefix(data, "important_dates:delete:cancel") {
		h.Reply(chatID, "Удаление важной даты отменено")
	} else {
		h.Reply(chatID, "Произошла ошибка")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}

func (h *Handler) EditImportantDate(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	importantDates, err := h.Store.GetImportantDates(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении списка важных дат", err)
		return
	}

	if len(importantDates) == 0 {
		h.Reply(chatID, "У тебя не добавлены важные даты")
		return
	}

	sortedImportantDates := h.beautifyImportantDates(importantDates)

	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, importantDate := range sortedImportantDates {
		callbackData := fmt.Sprintf("important_dates:update_menu:%d", importantDate.ID)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(importantDate.Title, callbackData),
		}
		buttons = append(buttons, row)
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "important_dates:update_menu:cancel"),
	})

	text := "<b>Выбери важную дату</b>"
	markup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, markup)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleEditImportantDate(_ context.Context, cq *tgbotapi.CallbackQuery) {
	data := cq.Data
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	data = strings.TrimPrefix(data, "important_dates:update_menu:")
	if data == "cancel" {
		h.Reply(chatID, "Редактирование важной даты отменено")
	} else {
		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Название", "important_dates:update:title:"+data),
				tgbotapi.NewInlineKeyboardButtonData("Дата", "important_dates:update:date:"+data),
				tgbotapi.NewInlineKeyboardButtonData("Партнёр", "important_dates:update:partner:"+data),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Уведомлять за", "important_dates:update:notify_before:"+data),
				tgbotapi.NewInlineKeyboardButtonData("Активность", "important_dates:update:is_active:"+data),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "important_dates:update:cancel"),
			),
		)

		text := "Что ты хочешь изменить?"

		err := h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
			return
		}
	}
	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}
}

func (h *Handler) CancelCallbackImportantDate(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "Действие отменено")
}

func (h *Handler) HandleEditTitleImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:title:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии редактирования", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingEditTitleImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "✍️ Введи новое название памятной даты")
}

func (h *Handler) HandleEditTitleImportantDateText(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия редактирования истекла", err)
		return
	}

	date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.Title = msg.Text

	err = h.Store.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении названия", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)
	err = h.Store.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	h.Reply(chatID, "✅ Название обновлено")
}

func (h *Handler) HandleEditDateImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:date:"))

	err := h.importantDateEditDrafts.Save(ctx, userID, &domain.ImportantDateEditDraft{
		ImportantDateID: int64(id),
	})
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении сессии редактирования", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingEditDateImportantDate)
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

func (h *Handler) HandleEditPartnerImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

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

func (h *Handler) HandleEditPartnerImportantDateSelect(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия истекла", err)
		return
	}

	date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	switch cq.Data {
	case "important_dates:edit:partner:false":
		date.PartnerID = sql.NullInt64{Valid: false}
	case "important_dates:edit:partner:true":
		partnerID, _ := h.Store.GetPartnerID(ctx, userID)
		date.PartnerID = sql.NullInt64{Int64: partnerID, Valid: true}
	}

	err = h.Store.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении партнёра", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "👥 Партнёр обновлён")
}

func (h *Handler) HandleEditNotifyBeforeImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

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

func (h *Handler) HandleEditNotifyBeforeImportantDateSelect(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

	draft, err := h.importantDateEditDrafts.Get(ctx, userID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Сессия истекла", err)
		return
	}

	days, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:edit:notify_before:"))

	date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.NotifyBeforeDays = days

	err = h.Store.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении уведомлений", err)
		return
	}

	_ = h.importantDateEditDrafts.Delete(ctx, userID)

	if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
	}

	h.Reply(chatID, "⏰ Уведомления обновлены")
}

func (h *Handler) HandleEditIsActiveImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	id, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:update:is_active:"))

	date, err := h.Store.GetImportantDateByID(ctx, int64(id))
	if err != nil {
		h.HandleErr(chatID, "Дата не найдена", err)
		return
	}

	date.IsActive = !date.IsActive

	err = h.Store.EditImportantDate(ctx, date)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при обновлении активности", err)
		return
	}

	h.ui.RemoveButtons(chatID, messageID)

	if date.IsActive {
		h.Reply(chatID, "🟢 Дата активирована")
	} else {
		h.Reply(chatID, "⚪ Дата деактивирована")
	}
}

func (h *Handler) HandleYearImportantDateUniversal(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID
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

			date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}

			date.Date = time.Date(year, date.Date.Month(), date.Date.Day(), 0, 0, 0, 0, time.Local)
			if er := h.Store.EditImportantDate(ctx, date); er != nil {
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

func (h *Handler) HandleMonthImportantDateUniversal(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID
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
			date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}
			date.Date = time.Date(date.Date.Year(), time.Month(month), date.Date.Day(), 0, 0, 0, 0, time.Local)
			if er := h.Store.EditImportantDate(ctx, date); er != nil {
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

func (h *Handler) HandleDayImportantDateUniversal(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID
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

			date, err := h.Store.GetImportantDateByID(ctx, draft.ImportantDateID)
			if err != nil {
				h.HandleErr(chatID, "Дата не найдена", err)
				return
			}

			date.Date = time.Date(date.Date.Year(), date.Date.Month(), day, 0, 0, 0, 0, time.Local)
			if er := h.Store.EditImportantDate(ctx, date); er != nil {
				h.HandleErr(chatID, "Ошибка при обновлении дня", er)
				return
			}

			_ = h.importantDateEditDrafts.Delete(ctx, userID)
			_ = h.Store.SetUserState(ctx, userID, domain.Empty)

			if er := h.ui.Client.DeleteMessage(chatID, messageID); er != nil {
				h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
			}

			h.Reply(chatID, "📅 Дата обновлена")
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
			partnerID, er := h.Store.GetPartnerID(ctx, userID)
			if er != nil {
				h.HandleErr(chatID, "Ошибка при получении id партнера", er)
				return
			}

			if partnerID == 0 {
				h.Reply(chatID, "Так как у тебя не добавлен партнер, памятная дата будет твоей личной")
				_ = h.Store.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
				_ = h.ui.SendNotifyBeforeKeyboard(chatID, isEdit)
			} else {
				_ = h.Store.SetUserState(ctx, userID, domain.AwaitingPartnerImportantDate)
				_ = h.ui.SendPartnerKeyboard(chatID, isEdit)
			}
		}
		return
	}

	h.HandleErr(chatID, "Неизвестный callback для дня", nil)
}
