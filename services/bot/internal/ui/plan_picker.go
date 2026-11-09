package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

const planPickerLayout = "200601021504"

func ParsePlanPickerTime(value string) (time.Time, error) {
	return time.ParseInLocation(planPickerLayout, value, planTimeZone)
}

func (ui *MenuUI) SendPlanPicker(chatID int64, session, screen string, at, now time.Time, question string) error {
	now = now.In(planTimeZone)
	if at.IsZero() {
		at = now
	}
	at = at.In(planTimeZone)
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, planTimeZone)
	month := time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, planTimeZone)
	hour := day.Add(time.Duration(at.Hour()) * time.Hour)
	limit := now.AddDate(1, 0, 0)
	var keyboard domain.InlineKeyboard
	button := func(label, next string, value time.Time) domain.InlineKeyboardButton {
		data := "plan:time:" + session + ":" + next
		if !value.IsZero() {
			data += ":" + value.Format(planPickerLayout)
		}
		return domain.InlineKeyboardButton{Text: label, Data: data}
	}
	addRow := func(buttons ...domain.InlineKeyboardButton) {
		if len(buttons) > 0 {
			keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{Buttons: buttons})
		}
	}
	var choices []domain.InlineKeyboardButton
	addChoices := func(width int) {
		for len(choices) > width {
			addRow(choices[:width]...)
			choices = choices[width:]
		}
		addRow(choices...)
	}
	var text string
	switch screen {
	case "month":
		if !month.AddDate(0, 1, 0).After(now) || month.After(limit) {
			month = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, planTimeZone)
		}
		nextMonth := month.AddDate(0, 1, 0)
		text = fmt.Sprintf("📅 Выбери день · %s %d", months[month.Month()-1], month.Year())
		for date := month; date.Before(nextMonth) && !date.After(limit); date = date.AddDate(0, 0, 1) {
			if date.Add(23*time.Hour + 59*time.Minute).After(now) {
				choices = append(choices, button(strconv.Itoa(date.Day()), "hours", date))
			}
		}
		addChoices(7)
		var navigation []domain.InlineKeyboardButton
		if month.After(now) {
			navigation = append(navigation, button("← Месяц", "month", month.AddDate(0, -1, 0)))
		}
		if !nextMonth.After(limit) {
			navigation = append(navigation, button("Месяц →", "month", nextMonth))
		}
		addRow(navigation...)
	case "hours":
		text = "🕓 Выбери час · " + day.Format("02.01.2006") + " МСК"
		for i := 0; i < 24; i++ {
			slot := day.Add(time.Duration(i) * time.Hour)
			if slot.Add(59*time.Minute).After(now) && !slot.After(limit) {
				choices = append(choices, button(slot.Format("15:00"), "minutes", slot))
			}
		}
		addChoices(6)
		addRow(button("← Календарь", "month", month))
	case "minutes", "exact":
		text = "🕓 Выбери минуты · " + hour.Format("02.01.2006, 15:00") + " МСК"
		increment, width := 5, 4
		if screen == "exact" {
			increment, width = 1, 6
		}
		for i := 0; i < 60; i += increment {
			slot := hour.Add(time.Duration(i) * time.Minute)
			if plans.ValidStart(slot, now) {
				choices = append(choices, button(slot.Format("15:04"), "review", slot))
			}
		}
		addChoices(width)
		if screen == "minutes" {
			addRow(button("Точная минута", "exact", hour))
		}
		addRow(button("← Выбор часа", "hours", day))
	case "review":
		text = "🗓 Предлагаем партнёру <b>" + at.Format(planTimeLayout) + " МСК</b>?"
		addRow(button("💌 Отправить", "send", time.Time{}))
		back := "minutes"
		if at.Minute()%5 != 0 {
			back = "exact"
		}
		addRow(button("← Выбор минут", back, hour))
	default:
		return fmt.Errorf("unknown plan picker screen %q", screen)
	}
	addRow(button("Отмена", "cancel", time.Time{}))
	if question != "" {
		text = question + "\n\n" + text
	}
	return ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
}
