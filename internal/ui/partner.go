package ui

import (
	"github.com/Waycoolers/fmlbot/internal/domain"
)

func (ui *MenuUI) PartnerMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.AddPartner},
					{domain.DeletePartner},
					{domain.Main},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}
