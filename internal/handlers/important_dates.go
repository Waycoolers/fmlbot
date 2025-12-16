package handlers

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	err = h.ui.SendYearKeyboard(chatID, time.Now().Year())
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора года", err)
		return
	}
}

func (h *Handler) HandleYearImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	messageID := cq.Message.MessageID

	draft, err := h.importantDateDrafts.Get(ctx, userID)
	if err != nil {
		h.ui.RemoveButtons(chatID, messageID)
		h.HandleErr(chatID, "Ошибка при получении черновика", err)
		return
	}
	if draft == nil {
		h.ui.RemoveButtons(chatID, messageID)
		h.HandleErr(chatID, "Черновик пустой", err)
		return
	}

	switch {
	case strings.HasPrefix(cq.Data, "important_dates:add:year:select:"):
		year, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:year:select:"))

		draft.Year = year

		er := h.importantDateDrafts.Save(ctx, userID, draft)
		if er != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при сохранении года важной даты", er)
			return
		}

		h.ui.RemoveButtons(chatID, messageID)
		er = h.ui.Client.DeleteMessage(chatID, messageID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при удалении сообщения", er)
		}

		err = h.ui.SendMonthKeyboard(chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора месяца", err)
			return
		}

	case strings.HasPrefix(cq.Data, "important_dates:add:year:page:"):
		startYear, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:year:page:"))
		keyboard := h.ui.BuildYearKeyboard(startYear)

		er := h.ui.Client.EditMessageReplyMarkup(
			chatID,
			messageID,
			keyboard,
		)
		if er != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при редактировании кнопок", er)
			return
		}
	default:
		h.HandleErr(chatID, "Неизвестный префикс у cq.Data", nil)
		err = h.ui.Client.DeleteMessage(chatID, messageID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
		}
		return
	}
}

func (h *Handler) HandleMonthImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
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

	month, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:month:"))
	draft.Month = month
	err = h.importantDateDrafts.Save(ctx, userID, draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении месяца важной даты", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
	}

	err = h.ui.SendDayKeyboard(chatID, draft.Year, month)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора дня", err)
		return
	}
}

func (h *Handler) HandleDayImportantDate(ctx context.Context, cq *tgbotapi.CallbackQuery) {
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

	day, _ := strconv.Atoi(strings.TrimPrefix(cq.Data, "important_dates:add:day:"))
	draft.Day = day
	err = h.importantDateDrafts.Save(ctx, userID, draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении дня важной даты", err)
		return
	}

	err = h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при удалении сообщения", err)
	}

	// Далее
	partnerID, er := h.Store.GetPartnerID(ctx, userID)
	if er != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", er)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "Так как у тебя не добавлен партнер, памятная дата будет твоей личной")

		err = h.Store.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при установке состояния", err)
			return
		}

		err = h.ui.SendNotifyBeforeKeyboard(chatID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора количества дней", err)
			return
		}
	} else {
		err = h.Store.SetUserState(ctx, userID, domain.AwaitingPartnerImportantDate)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при установке состояния", err)
			return
		}

		err = h.ui.SendPartnerKeyboard(chatID)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при отправке клавиатуры для выбора партнера в важной дате", err)
			return
		}
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

	err = h.ui.SendNotifyBeforeKeyboard(chatID)
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
}
