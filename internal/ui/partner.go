package ui

import (
	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ui *MenuUI) PartnerMenu(chatID int64, text string) error {
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.AddPartner)),
			tgbotapi.NewKeyboardButton(string(domain.DeletePartner)),
			tgbotapi.NewKeyboardButton(string(domain.Main)),
		),
	)

	menu.ResizeKeyboard = true
	menu.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = menu

	_, err := ui.Client.Send(msg)
	return err
}
