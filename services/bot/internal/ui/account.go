package ui

import (
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (ui *MenuUI) AccountMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.ChangePassword},
					{domain.DeleteAccount},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{domain.Main},
				},
			},
		},
	}
	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}
