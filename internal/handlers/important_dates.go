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
	h.Reply(chatID, "Введи дату")
}

func (h *Handler) HandleDateImportantDate(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	date := strings.TrimSpace(msg.Text)

	draft, err := h.importantDateDrafts.Get(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении черновика", err)
		return
	}
	if draft == nil {
		h.HandleErr(chatID, "Черновик пустой", err)
		return
	}

	parsedDate, err := time.Parse("02.01.2006", date)
	if err != nil {
		h.Reply(
			chatID,
			"😔 Не смог распознать дату.\n"+
				"Пожалуйста, введи её в формате: `ДД.ММ.ГГГГ`\n"+
				"Например: `14.02.2024`",
		)
		return
	}

	draft.Date = parsedDate
	err = h.importantDateDrafts.Save(ctx, userID, draft)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сохранении даты важной даты", err)
		return
	}

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

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("0", "important_dates:add:notify_before:0"),
				tgbotapi.NewInlineKeyboardButtonData("1", "important_dates:add:notify_before:1"),
				tgbotapi.NewInlineKeyboardButtonData("3", "important_dates:add:notify_before:3"),
				tgbotapi.NewInlineKeyboardButtonData("7", "important_dates:add:notify_before:7"),
			),
		)

		text := "Выбери, за сколько дней до даты тебе напомнить о ней"

		err = h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при отправке кнопок", err)
			return
		}
	} else {
		err = h.Store.SetUserState(ctx, userID, domain.AwaitingPartnerImportantDate)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при установке состояния", err)
			return
		}

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👤 Только для меня", "important_dates:add:partner:false"),
				tgbotapi.NewInlineKeyboardButtonData("💑 Общая с партнёром", "important_dates:add:partner:true"),
			),
		)

		text := "👥 Эта дата будет:"

		err = h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при отправке кнопок", err)
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

	err = h.Store.SetUserState(ctx, userID, domain.AwaitingNotifyBeforeImportantDate)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния", err)
		return
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("0", "important_dates:add:notify_before:0"),
			tgbotapi.NewInlineKeyboardButtonData("1", "important_dates:add:notify_before:1"),
			tgbotapi.NewInlineKeyboardButtonData("3", "important_dates:add:notify_before:3"),
			tgbotapi.NewInlineKeyboardButtonData("7", "important_dates:add:notify_before:7"),
		),
	)

	text := "Выбери, за сколько дней до даты тебе напомнить о ней"

	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке кнопок", err)
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

	_, err = h.Store.AddImportantDate(ctx, sql.NullInt64{Int64: userID, Valid: true}, finalDraft.PartnerID, finalDraft.Title,
		finalDraft.Date, finalDraft.NotifyBeforeDays)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при добавлении важной даты", err)
		return
	}

	h.Reply(chatID, "Памятная дата добавлена")
}
