package ui

import (
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (ui *MenuUI) ComplimentsMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.AddCompliment},
					{domain.DeleteCompliment},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{domain.GetCompliments},
					{domain.ReceiveCompliment},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{domain.EditComplimentFrequency},
					{domain.Main},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}

func (ui *MenuUI) EditComplimentFrequencyMenu(chatID int64, text string) error {
	err := ui.Client.SendMessage(chatID, text)
	if err != nil {
		return err
	}
	return nil
}
