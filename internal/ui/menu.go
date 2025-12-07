package ui

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/client"
	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MenuUI struct {
	Client client.BotClient
}

func New(client client.BotClient) *MenuUI {
	return &MenuUI{Client: client}
}

func (ui *MenuUI) StartMenu(chatID int64) error {
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.Register)),
		),
	)

	menu.ResizeKeyboard = true
	menu.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Чтобы разбудить бота, зарегистрируйся по кнопке ниже")
	msg.ReplyMarkup = menu

	_, err := ui.Client.Send(msg)
	return err
}

func (ui *MenuUI) MainMenu(chatID int64) error {
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(string(domain.Account)),
			tgbotapi.NewKeyboardButton(string(domain.Partner)),
			tgbotapi.NewKeyboardButton(string(domain.Compliments)),
		),
	)

	menu.ResizeKeyboard = true
	menu.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, "fmlbot приветствует тебя! 💖")
	msg.ReplyMarkup = menu

	_, err := ui.Client.Send(msg)
	return err
}

func (ui *MenuUI) RemoveButtons(chatID int64, messageID int) {
	empty := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	if err := ui.Client.EditMessageReplyMarkup(chatID, messageID, empty); err != nil {
		log.Printf("Ошибка при удалении кнопок: %v", err)
	}
}
