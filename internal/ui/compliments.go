package ui

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (ui *MenuUI) ComplimentsMenu(chatID int64) error {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Добавить комплимент", "compliments:add")},
		{tgbotapi.NewInlineKeyboardButtonData("Удалить комплимент", "compliments:delete")},
		{tgbotapi.NewInlineKeyboardButtonData("Все комплименты", "compliments:all")},
		{tgbotapi.NewInlineKeyboardButtonData("Получить комплимент", "compliments:receive")},
		{tgbotapi.NewInlineKeyboardButtonData("Назад", "menu:main")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	err := ui.Client.SendWithInlineKeyboard(chatID, "Выберите действие:", kb)
	if err != nil {
		return err
	}
	return nil
}
