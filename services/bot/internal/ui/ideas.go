package ui

import "github.com/Waycoolers/fmlbot/services/bot/internal/domain"

func (ui *MenuUI) IdeasMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{Command: domain.GenerateLeisureIdea},
					{Command: domain.PlanForTwo},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{Command: domain.MyPlans},
					{Command: domain.Main},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}
