package ui

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (ui *MenuUI) PartnerMenu(chatID int64) error {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Добавить партнёра", "partner:add")},
		{tgbotapi.NewInlineKeyboardButtonData("Удалить партнёра", "partner:delete")},
		{tgbotapi.NewInlineKeyboardButtonData("Назад", "menu:main")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	err := ui.Client.SendWithInlineKeyboard(chatID, "Выберите действие:", kb)
	if err != nil {
		return err
	}
	return nil
}
