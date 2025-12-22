package ui

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ui *MenuUI) ImportantDatesMenu(chatID int64, text string) error {
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.AddImportantDate)),
			tgbotapi.NewKeyboardButton(string(domain.GetImportantDates)),
			tgbotapi.NewKeyboardButton(string(domain.DeleteImportantDate)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.EditImportantDate)),
			tgbotapi.NewKeyboardButton(string(domain.Main)),
		),
	)

	menu.ResizeKeyboard = true
	menu.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = menu

	_, err := ui.Client.Send(msg)
	return err
}

const (
	YearStart    = 1920
	YearsPerPage = 12
)

var months = []string{
	"Янв", "Фев", "Мар",
	"Апр", "Май", "Июн",
	"Июл", "Авг", "Сен",
	"Окт", "Ноя", "Дек",
}

// ---------------------- YEAR ----------------------
func (ui *MenuUI) BuildYearKeyboard(pageStart int, isEdit bool) tgbotapi.InlineKeyboardMarkup {
	currentYear := time.Now().Year()
	if pageStart < YearStart {
		pageStart = YearStart
	}
	if pageStart > currentYear {
		pageStart = currentYear
	}

	prefix := "important_dates:add:"
	if isEdit {
		prefix = "important_dates:edit:"
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	year := pageStart
	for i := 0; i < YearsPerPage && year >= YearStart; i++ {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(year),
			fmt.Sprintf("%syear:select:%d", prefix, year),
		)

		if len(rows) == 0 || len(rows[len(rows)-1]) == 3 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		} else {
			rows[len(rows)-1] = append(rows[len(rows)-1], btn)
		}
		year--
	}

	for _, row := range rows {
		slices.Reverse(row)
	}
	slices.Reverse(rows)

	// Навигация
	var navRow []tgbotapi.InlineKeyboardButton
	if pageStart > YearStart {
		navRow = append(navRow,
			tgbotapi.NewInlineKeyboardButtonData(
				"⬅️ Назад",
				fmt.Sprintf("%syear:page:%d", prefix, pageStart-YearsPerPage),
			),
		)
	}
	if pageStart+YearsPerPage <= currentYear {
		navRow = append(navRow,
			tgbotapi.NewInlineKeyboardButtonData(
				"Вперёд ➡️",
				fmt.Sprintf("%syear:page:%d", prefix, pageStart+YearsPerPage),
			),
		)
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (ui *MenuUI) SendYearKeyboard(chatID int64, pageStart int, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери год",
		ui.BuildYearKeyboard(pageStart, isEdit),
	)
}

// ---------------------- MONTH ----------------------
func (ui *MenuUI) BuildMonthKeyboard(isEdit bool) tgbotapi.InlineKeyboardMarkup {
	prefix := "important_dates:add:month:"
	if isEdit {
		prefix = "important_dates:edit:month:"
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < 12; i += 3 {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(months[i], fmt.Sprintf("%s%d", prefix, i+1)),
			tgbotapi.NewInlineKeyboardButtonData(months[i+1], fmt.Sprintf("%s%d", prefix, i+2)),
			tgbotapi.NewInlineKeyboardButtonData(months[i+2], fmt.Sprintf("%s%d", prefix, i+3)),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (ui *MenuUI) SendMonthKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери месяц",
		ui.BuildMonthKeyboard(isEdit),
	)
}

// ---------------------- DAY ----------------------
func (ui *MenuUI) BuildDayKeyboard(year, month int, isEdit bool) tgbotapi.InlineKeyboardMarkup {
	days := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
	prefix := "important_dates:add:day:"
	if isEdit {
		prefix = "important_dates:edit:day:"
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for d := 1; d <= days; d++ {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(d),
			fmt.Sprintf("%s%d", prefix, d),
		)

		if len(rows) == 0 || len(rows[len(rows)-1]) == 7 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		} else {
			rows[len(rows)-1] = append(rows[len(rows)-1], btn)
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (ui *MenuUI) SendDayKeyboard(chatID int64, year, month int, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери день",
		ui.BuildDayKeyboard(year, month, isEdit),
	)
}

// ---------------------- PARTNER ----------------------
func (ui *MenuUI) BuildPartnerKeyboard(isEdit bool) tgbotapi.InlineKeyboardMarkup {
	prefix := "important_dates:add:partner:"
	if isEdit {
		prefix = "important_dates:edit:partner:"
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Только для меня", prefix+"false"),
			tgbotapi.NewInlineKeyboardButtonData("👩‍❤️‍👨 Общая с партнёром", prefix+"true"),
		),
	)

	return buttons
}

func (ui *MenuUI) SendPartnerKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"👥 Эта дата будет:",
		ui.BuildPartnerKeyboard(isEdit),
	)
}

// ---------------------- NOTIFY BEFORE ----------------------
func (ui *MenuUI) BuildNotifyBeforeKeyboard(isEdit bool) tgbotapi.InlineKeyboardMarkup {
	prefix := "important_dates:add:notify_before:"
	if isEdit {
		prefix = "important_dates:edit:notify_before:"
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("0", prefix+"0"),
			tgbotapi.NewInlineKeyboardButtonData("1", prefix+"1"),
			tgbotapi.NewInlineKeyboardButtonData("3", prefix+"3"),
			tgbotapi.NewInlineKeyboardButtonData("7", prefix+"7"),
		),
	)

	return buttons
}

func (ui *MenuUI) SendNotifyBeforeKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"Выбери, за сколько дней до даты тебе напомнить о ней",
		ui.BuildNotifyBeforeKeyboard(isEdit),
	)
}
