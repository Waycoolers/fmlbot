package ui

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/domain"
)

type MenuUI struct {
	Client domain.BotClient
}

func New(client domain.BotClient) *MenuUI {
	return &MenuUI{Client: client}
}

func (ui *MenuUI) StartMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.Register},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}

func (ui *MenuUI) MainMenu(chatID int64, text string) error {
	keyboard := domain.Keyboard{
		Rows: []domain.KeyboardRow{
			{
				Buttons: []domain.KeyboardButton{
					{domain.Account},
					{domain.Partner},
				},
			},
			{
				Buttons: []domain.KeyboardButton{
					{domain.Compliments},
					{domain.ImportantDates},
				},
			},
		},
	}

	_, err := ui.Client.SendKeyboard(chatID, text, keyboard)
	return err
}

func (ui *MenuUI) RemoveButtons(chatID int64, messageID int) {
	if err := ui.Client.DeleteMessageReplyMarkup(chatID, messageID); err != nil {
		log.Printf("Ошибка при удалении кнопок: %v", err)
	}
}
