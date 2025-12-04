package ui

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/client"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MenuUI struct {
	Client client.BotClient
}

func New(client client.BotClient) *MenuUI {
	return &MenuUI{Client: client}
}

func (ui *MenuUI) MainMenu(chatID int64) error {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("👤 Партнёр", "menu:partner")},
		{tgbotapi.NewInlineKeyboardButtonData("❤️ Комплименты", "menu:compliments")},
		{tgbotapi.NewInlineKeyboardButtonData("⚙ Аккаунт", "menu:account")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if err := ui.Client.SendWithInlineKeyboard(chatID, "Выберите действие:", kb); err != nil {
		return err
	}
	return nil
}

//func (ui *MenuUI) MainMenu(chatID int64) error {
//	menu := tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton("Аккаунт"),
//			tgbotapi.NewKeyboardButton("Партнёр"),
//			tgbotapi.NewKeyboardButton("Комплименты"),
//		),
//	)
//
//	menu.ResizeKeyboard = true
//	menu.OneTimeKeyboard = false
//
//	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать")
//	msg.ReplyMarkup = menu
//
//	_, err := ui.Client.Send(msg)
//	return err
//}

func (ui *MenuUI) RemoveButtons(chatID int64, messageID int) {
	empty := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	if err := ui.Client.EditMessageReplyMarkup(chatID, messageID, empty); err != nil {
		log.Printf("Ошибка при удалении кнопок: %v", err)
	}
}
