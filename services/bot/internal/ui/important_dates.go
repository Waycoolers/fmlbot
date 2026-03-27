package ui

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (ui *MenuUI) ImportantDatesMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.AddImportantDate},
					{domain.GetImportantDates},
					{domain.DeleteImportantDate},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{domain.EditImportantDate},
					{domain.Main},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
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
func (ui *MenuUI) BuildYearKeyboard(pageStart int, isEdit bool) domain.InlineKeyboard {
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

	var rows []domain.InlineKeyboardRow
	var row domain.InlineKeyboardRow
	year := pageStart
	for i := 0; i < YearsPerPage && year >= YearStart; i++ {
		btn := domain.InlineKeyboardButton{
			Text: strconv.Itoa(year),
			Data: fmt.Sprintf("%syear:select:%d", prefix, year),
		}

		if len(row.Buttons) < 3 {
			row.Buttons = append(row.Buttons, btn)
			year--
		} else {
			rows = append(rows, row)
			row = domain.InlineKeyboardRow{}
		}
	}

	for _, row := range rows {
		slices.Reverse(row.Buttons)
	}
	slices.Reverse(rows)

	// Навигация
	var navRow domain.InlineKeyboardRow
	if pageStart > YearStart {
		btn := domain.InlineKeyboardButton{
			Text: "⬅️ Назад",
			Data: fmt.Sprintf("%syear:page:%d", prefix, pageStart-YearsPerPage),
		}
		navRow.Buttons = append(navRow.Buttons, btn)
	}
	if pageStart+YearsPerPage <= currentYear {
		btn := domain.InlineKeyboardButton{
			Text: "Вперёд ➡️",
			Data: fmt.Sprintf("%syear:page:%d", prefix, pageStart+YearsPerPage),
		}
		navRow.Buttons = append(navRow.Buttons, btn)
	}
	if len(navRow.Buttons) > 0 {
		rows = append(rows, navRow)
	}

	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}

	return keyboard
}

func (ui *MenuUI) SendYearKeyboard(chatID int64, pageStart int, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери год",
		ui.BuildYearKeyboard(pageStart, isEdit),
	)
}

// ---------------------- MONTH ----------------------
func (ui *MenuUI) BuildMonthKeyboard(isEdit bool) domain.InlineKeyboard {
	prefix := "important_dates:add:month:"
	if isEdit {
		prefix = "important_dates:edit:month:"
	}

	var rows []domain.InlineKeyboardRow
	for i := 0; i < 12; i += 3 {
		row := domain.InlineKeyboardRow{
			Buttons: []domain.InlineKeyboardButton{
				{Text: months[i], Data: fmt.Sprintf("%s%d", prefix, i+1)},
				{Text: months[i+1], Data: fmt.Sprintf("%s%d", prefix, i+2)},
				{Text: months[i+2], Data: fmt.Sprintf("%s%d", prefix, i+3)},
			},
		}
		rows = append(rows, row)
	}

	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}

	return keyboard
}

func (ui *MenuUI) SendMonthKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери месяц",
		ui.BuildMonthKeyboard(isEdit),
	)
}

// ---------------------- DAY ----------------------
func (ui *MenuUI) BuildDayKeyboard(year, month int, isEdit bool) domain.InlineKeyboard {
	days := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
	prefix := "important_dates:add:day:"
	if isEdit {
		prefix = "important_dates:edit:day:"
	}

	var rows []domain.InlineKeyboardRow
	var row domain.InlineKeyboardRow
	for d := 1; d <= days; d++ {
		btn := domain.InlineKeyboardButton{
			Text: strconv.Itoa(d),
			Data: fmt.Sprintf("%s%d", prefix, d),
		}

		if len(row.Buttons) < 7 {
			row.Buttons = append(row.Buttons, btn)
		} else {
			rows = append(rows, row)
			row = domain.InlineKeyboardRow{}
		}
	}
	rows = append(rows, row)

	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}

	return keyboard
}

func (ui *MenuUI) SendDayKeyboard(chatID int64, year, month int, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери день",
		ui.BuildDayKeyboard(year, month, isEdit),
	)
}

// ---------------------- PARTNER ----------------------
func (ui *MenuUI) BuildPartnerKeyboard(isEdit bool) domain.InlineKeyboard {
	prefix := "important_dates:add:partner:"
	if isEdit {
		prefix = "important_dates:edit:partner:"
	}

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "👤 Только для меня", Data: prefix + "false"},
					{Text: "👩‍❤️‍👨 Общая с партнёром", Data: prefix + "true"},
				},
			},
		},
	}

	return keyboard
}

func (ui *MenuUI) SendPartnerKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"👥 Эта дата будет:",
		ui.BuildPartnerKeyboard(isEdit),
	)
}

// ---------------------- NOTIFY BEFORE ----------------------
func (ui *MenuUI) BuildNotifyBeforeKeyboard(isEdit bool) domain.InlineKeyboard {
	prefix := "important_dates:add:notify_before:"
	if isEdit {
		prefix = "important_dates:edit:notify_before:"
	}

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "0", Data: prefix + "0"},
					{Text: "1", Data: prefix + "1"},
					{Text: "3", Data: prefix + "3"},
					{Text: "7", Data: prefix + "7"},
				},
			},
		},
	}

	return keyboard
}

func (ui *MenuUI) SendNotifyBeforeKeyboard(chatID int64, isEdit bool) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"Выбери, за сколько дней до даты тебе напомнить о ней",
		ui.BuildNotifyBeforeKeyboard(isEdit),
	)
}
