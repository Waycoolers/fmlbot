package ui

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (ui *MenuUI) AccountMenu(chatID int64) error {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Зарегистрироваться", "account:register")},
		{tgbotapi.NewInlineKeyboardButtonData("Удалить аккаунт", "account:delete")},
		{tgbotapi.NewInlineKeyboardButtonData("Назад", "menu:main")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	err := ui.Client.SendWithInlineKeyboard(chatID, "Выберите действие:", kb)
	if err != nil {
		return err
	}
	return nil
}
