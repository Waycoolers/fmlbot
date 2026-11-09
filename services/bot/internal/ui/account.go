package ui

import (
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (ui *MenuUI) AccountMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{Command: domain.ChangePassword},
					{Command: domain.DeleteAccount},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{Command: domain.Main},
				},
			},
		},
	}
	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}
