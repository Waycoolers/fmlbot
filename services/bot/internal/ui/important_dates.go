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
					{Command: domain.AddImportantDate},
					{Command: domain.GetImportantDates},
					{Command: domain.DeleteImportantDate},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{Command: domain.EditImportantDate},
					{Command: domain.Main},
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

var moscow = loadMoscow()

func loadMoscow() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Local
	}
	return loc
}

var months = []string{
	"Янв", "Фев", "Мар",
	"Апр", "Май", "Июн",
	"Июл", "Авг", "Сен",
	"Окт", "Ноя", "Дек",
}

// ---------------------- YEAR ----------------------
func (ui *MenuUI) BuildYearKeyboard(pageStart int, isEdit bool, session ...string) domain.InlineKeyboard {
	currentYear := time.Now().In(moscow).Year()
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
	for year := pageStart; year >= YearStart && year > pageStart-YearsPerPage; year-- {
		row.Buttons = append(row.Buttons, domain.InlineKeyboardButton{
			Text: strconv.Itoa(year),
			Data: fmt.Sprintf("%syear:select:%d", prefix, year),
		})

		if len(row.Buttons) == 3 {
			rows = append(rows, row)
			row = domain.InlineKeyboardRow{}
		}
	}
	if len(row.Buttons) > 0 {
		rows = append(rows, row)
	}

	for _, row := range rows {
		slices.Reverse(row.Buttons)
	}
	slices.Reverse(rows)

	// Навигация
	var navRow domain.InlineKeyboardRow
	if pageStart-YearsPerPage >= YearStart {
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

	return dateKeyboardSession(keyboard, session)
}

func (ui *MenuUI) SendYearKeyboard(chatID int64, pageStart int, isEdit bool, session ...string) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери год",
		ui.BuildYearKeyboard(pageStart, isEdit, session...),
	)
}

// ---------------------- MONTH ----------------------
func (ui *MenuUI) BuildMonthKeyboard(year int, isEdit bool) domain.InlineKeyboard {
	// Можно выбрать только уже случившееся событие
	lastMonth := 12
	if now := time.Now().In(moscow); year == now.Year() {
		lastMonth = int(now.Month())
	}

	prefix := "important_dates:add:month:"
	if isEdit {
		prefix = "important_dates:edit:month:"
	}

	var rows []domain.InlineKeyboardRow
	for i := 0; i < lastMonth; i += 3 {
		var row domain.InlineKeyboardRow
		for month := i; month < i+3 && month < lastMonth; month++ {
			row.Buttons = append(row.Buttons, domain.InlineKeyboardButton{
				Text: months[month],
				Data: fmt.Sprintf("%s%d", prefix, month+1),
			})
		}
		rows = append(rows, row)
	}

	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}

	return keyboard
}

func (ui *MenuUI) SendMonthKeyboard(chatID int64, year int, isEdit bool, session ...string) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери месяц",
		dateKeyboardSession(ui.BuildMonthKeyboard(year, isEdit), session),
	)
}

// ---------------------- DAY ----------------------
func (ui *MenuUI) BuildDayKeyboard(year, month int, isEdit bool) domain.InlineKeyboard {
	days := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	// Можно выбрать только уже случившееся событие
	if now := time.Now().In(moscow); year == now.Year() && time.Month(month) == now.Month() {
		days = now.Day()
	}
	prefix := "important_dates:add:day:"
	if isEdit {
		prefix = "important_dates:edit:day:"
	}

	var rows []domain.InlineKeyboardRow
	var row domain.InlineKeyboardRow
	for d := 1; d <= days; d++ {
		row.Buttons = append(row.Buttons, domain.InlineKeyboardButton{
			Text: strconv.Itoa(d),
			Data: fmt.Sprintf("%s%d", prefix, d),
		})

		if len(row.Buttons) == 7 {
			rows = append(rows, row)
			row = domain.InlineKeyboardRow{}
		}
	}
	if len(row.Buttons) > 0 {
		rows = append(rows, row)
	}

	keyboard := domain.InlineKeyboard{
		Rows: rows,
	}

	return keyboard
}

func (ui *MenuUI) SendDayKeyboard(chatID int64, year, month int, isEdit bool, session ...string) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери день",
		dateKeyboardSession(ui.BuildDayKeyboard(year, month, isEdit), session),
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

func (ui *MenuUI) SendPartnerKeyboard(chatID int64, isEdit bool, session ...string) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"👥 Эта дата будет:",
		dateKeyboardSession(ui.BuildPartnerKeyboard(isEdit), session),
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

func (ui *MenuUI) SendNotifyBeforeKeyboard(chatID int64, isEdit bool, session ...string) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"Выбери, за сколько дней до даты тебе напомнить о ней",
		dateKeyboardSession(ui.BuildNotifyBeforeKeyboard(isEdit), session),
	)
}

func dateKeyboardSession(keyboard domain.InlineKeyboard, session []string) domain.InlineKeyboard {
	if len(session) == 0 || session[0] == "" {
		return keyboard
	}
	for i := range keyboard.Rows {
		for j := range keyboard.Rows[i].Buttons {
			keyboard.Rows[i].Buttons[j].Data += ":" + session[0]
		}
	}
	return keyboard
}
