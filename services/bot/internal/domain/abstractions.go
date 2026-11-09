package domain

import (
	"context"
	"errors"
)

// ErrRecipientUnavailable - Telegram окончательно отказал в доставке (бот заблокирован,
// чата не существует). Повторять отправку бессмысленно.
var ErrRecipientUnavailable = errors.New("recipient unavailable")

type BotClient interface {
	SendMessage(chatID int64, text string) error
	SendWithInlineKeyboard(chatID int64, text string, keyboard InlineKeyboard) error
	EditMessageReplyMarkup(chatID int64, messageID int, keyboard InlineKeyboard) error
	DeleteMessageReplyMarkup(chatID int64, messageID int) error
	GetUpdatesChan() <-chan Update
	StopReceivingUpdates()
	SendKeyboard(chatID int64, text string, keyboard Keyboard) (Message, error)
	DeleteMessage(chatID int64, messageID int) error
	AnswerCallback(callbackID string) error
}

type ApiClient interface {
	TelegramChatID(ctx context.Context, userID int64) (int64, error)
	LinkTelegram(ctx context.Context, chatID int64, code string) error
	CreateUser(ctx context.Context, chatID int64, username string) (string, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword string, password string) error
	ResetPassword(ctx context.Context, userID int64) (string, error)
	GetMe(ctx context.Context, chatID int64) (*User, error)
	GetPartner(ctx context.Context, chatID int64) (*User, error)
	DeleteMe(ctx context.Context, chatID int64) error
	RequestPartner(ctx context.Context, requesterID int64, partnerID int64) (int64, error)
	AcceptPartnerRequest(ctx context.Context, chatID int64, requestID int64) (int64, error)
	DeclinePartnerRequest(ctx context.Context, chatID int64, requestID int64) (int64, error)
	Unpair(ctx context.Context, chatID, pairVersion int64) error
	GetUserByUsername(ctx context.Context, requesterID int64, username string) (*User, error)

	GetMyUserConfig(ctx context.Context, chatID int64) (*UserConfig, error)
	GetPartnerUserConfig(ctx context.Context, chatID int64) (*UserConfig, error)
	UpdateUserConfig(ctx context.Context, chatID int64, maxCount *int) error

	AddCompliment(ctx context.Context, chatID int64, text string) (*Compliment, error)
	GetAllCompliments(ctx context.Context, chatID int64) ([]Compliment, error)
	UpdateCompliment(ctx context.Context, chatID int64, complimentID int64, text string, isSent bool) error
	DeleteCompliment(ctx context.Context, chatID int64, complimentID int64) error
	ReceiveNextCompliment(ctx context.Context, chatID int64) (*Compliment, error)

	AddImportantDate(ctx context.Context, chatID int64, req ImportantDateRequest) (*ImportantDate, error)
	GetImportantDate(ctx context.Context, chatID int64, dateID int64) (*ImportantDate, error)
	GetAllImportantDates(ctx context.Context, chatID int64) ([]ImportantDate, error)
	UpdateImportantDate(ctx context.Context, chatID int64, dateID int64, req ImportantDateRequest) error
	UpdateImportantDateSharing(ctx context.Context, chatID int64, dateID int64, makeShared bool, expectedPairVersion *int64) error
	DeleteImportantDate(ctx context.Context, chatID int64, dateID int64) error

	GetLeisureIdea(ctx context.Context, chatID int64, location, level, budget, context string) (string, error)
}

type AuthClient interface {
	GetAccessToken(ctx context.Context, userID int64) (string, error)
}
