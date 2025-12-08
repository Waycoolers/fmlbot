package ui

import (
	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ui *MenuUI) ComplimentsMenu(chatID int64, text string) error {
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.AddCompliment)),
			tgbotapi.NewKeyboardButton(string(domain.DeleteCompliment)),
			tgbotapi.NewKeyboardButton(string(domain.GetCompliments)),
			tgbotapi.NewKeyboardButton(string(domain.ReceiveCompliment)),
			tgbotapi.NewKeyboardButton(string(domain.EditComplimentFrequency)),
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

func (ui *MenuUI) EditComplimentFrequencyMenu(chatID int64, text string) error {
	err := ui.Client.SendMessage(chatID, text)
	if err != nil {
		return err
	}
	return nil
}
