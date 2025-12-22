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
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.DeleteImportantDate)),
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

func (ui *MenuUI) BuildYearKeyboard(pageStart int) tgbotapi.InlineKeyboardMarkup {
	currentYear := time.Now().Year()

	if pageStart < YearStart {
		pageStart = YearStart
	}
	if pageStart > currentYear {
		pageStart = currentYear
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	year := pageStart

	for i := 0; i < YearsPerPage && year >= YearStart; i++ {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(year),
			fmt.Sprintf("important_dates:add:year:select:%d", year),
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
				fmt.Sprintf("important_dates:add:year:page:%d", pageStart-YearsPerPage),
			),
		)
	}

	if pageStart+YearsPerPage <= currentYear {
		navRow = append(navRow,
			tgbotapi.NewInlineKeyboardButtonData(
				"Вперёд ➡️",
				fmt.Sprintf("important_dates:add:year:page:%d", pageStart+YearsPerPage),
			),
		)
	}

	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (ui *MenuUI) SendYearKeyboard(chatID int64, pageStart int) error {
	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери год",
		ui.BuildYearKeyboard(pageStart),
	)
}

var months = []string{
	"Янв", "Фев", "Мар",
	"Апр", "Май", "Июн",
	"Июл", "Авг", "Сен",
	"Окт", "Ноя", "Дек",
}

func (ui *MenuUI) SendMonthKeyboard(chatID int64) error {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < 12; i += 3 {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(months[i], fmt.Sprintf("important_dates:add:month:%d", i+1)),
			tgbotapi.NewInlineKeyboardButtonData(months[i+1], fmt.Sprintf("important_dates:add:month:%d", i+2)),
			tgbotapi.NewInlineKeyboardButtonData(months[i+2], fmt.Sprintf("important_dates:add:month:%d", i+3)),
		)
		rows = append(rows, row)
	}

	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери месяц",
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
}

func (ui *MenuUI) SendDayKeyboard(chatID int64, year int, month int) error {
	days := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()

	var rows [][]tgbotapi.InlineKeyboardButton
	for d := 1; d <= days; d++ {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(d),
			fmt.Sprintf("important_dates:add:day:%d", d),
		)

		if len(rows) == 0 || len(rows[len(rows)-1]) == 7 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		} else {
			rows[len(rows)-1] = append(rows[len(rows)-1], btn)
		}
	}

	return ui.Client.SendWithInlineKeyboard(
		chatID,
		"📅 Выбери день",
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
}

func (ui *MenuUI) SendPartnerKeyboard(chatID int64) error {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Только для меня", "important_dates:add:partner:false"),
			tgbotapi.NewInlineKeyboardButtonData("💑 Общая с партнёром", "important_dates:add:partner:true"),
		),
	)

	text := "👥 Эта дата будет:"

	return ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
}

func (ui *MenuUI) SendNotifyBeforeKeyboard(chatID int64) error {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("0", "important_dates:add:notify_before:0"),
			tgbotapi.NewInlineKeyboardButtonData("1", "important_dates:add:notify_before:1"),
			tgbotapi.NewInlineKeyboardButtonData("3", "important_dates:add:notify_before:3"),
			tgbotapi.NewInlineKeyboardButtonData("7", "important_dates:add:notify_before:7"),
		),
	)

	text := "Выбери, за сколько дней до даты тебе напомнить о ней"

	return ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
}
