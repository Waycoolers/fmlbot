package domain

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotClient interface {
	SendMessage(chatID int64, text string) error
	SendWithInlineKeyboard(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) error
	EditMessageReplyMarkup(chatID int64, messageID int, markup tgbotapi.InlineKeyboardMarkup) error
	GetUpdatesChan() <-chan tgbotapi.Update
	StopReceivingUpdates()
	Send(msg tgbotapi.Chattable) (tgbotapi.Message, error)
	DeleteMessage(chatID int64, messageID int) error
}

type ImportantDateDraftRepo interface {
	Get(ctx context.Context, userID int64) (*ImportantDateDraft, error)
	Save(ctx context.Context, userID int64, draft *ImportantDateDraft) error
	Delete(ctx context.Context, userID int64) error
}
