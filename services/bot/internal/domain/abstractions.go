package domain

import "context"

type BotClient interface {
	SendMessage(chatID int64, text string) error
	SendWithInlineKeyboard(chatID int64, text string, keyboard InlineKeyboard) error
	EditMessageReplyMarkup(chatID int64, messageID int, keyboard InlineKeyboard) error
	DeleteMessageReplyMarkup(chatID int64, messageID int) error
	GetUpdatesChan() <-chan Update
	StopReceivingUpdates()
	SendKeyboard(chatID int64, text string, keyboard Keyboard) (Message, error)
	DeleteMessage(chatID int64, messageID int) error
}

type ApiClient interface {
	CreateUser(ctx context.Context, chatID int64, username string) error
	GetMe(ctx context.Context, chatID int64) (*User, error)
	GetPartner(ctx context.Context, chatID int64) (*User, error)
	DeleteMe(ctx context.Context, chatID int64) error
	UpdateMe(ctx context.Context, chatID int64, username string, partnerID int64) error
	UpdatePartner(ctx context.Context, requesterID int64, username string, partnerID int64) error
	PairUsers(ctx context.Context, requesterID int64, partnerID int64) error
	Unpair(ctx context.Context, chatID int64) error
	GetUserByUsername(ctx context.Context, requesterID int64, username string) (*User, error)

	GetMyUserConfig(ctx context.Context, chatID int64) (*UserConfig, error)
	GetPartnerUserConfig(ctx context.Context, chatID int64) (*UserConfig, error)
	UpdateUserConfig(ctx context.Context, chatID int64, maxCount *int) error
	ResetMyUserConfig(ctx context.Context, chatID int64) error
	ResetPartnerUserConfig(ctx context.Context, chatID int64) error

	AddCompliment(ctx context.Context, chatID int64, text string) (*Compliment, error)
	GetAllCompliments(ctx context.Context, chatID int64) ([]Compliment, error)
	UpdateCompliment(ctx context.Context, chatID int64, complimentID int64, text string, isSent bool) error
	DeleteCompliment(ctx context.Context, chatID int64, complimentID int64) error
	ReceiveNextCompliment(ctx context.Context, chatID int64) (*Compliment, error)

	AddImportantDate(ctx context.Context, chatID int64, req ImportantDateRequest) (*ImportantDate, error)
	GetImportantDate(ctx context.Context, chatID int64, dateID int64) (*ImportantDate, error)
	GetAllImportantDates(ctx context.Context, chatID int64) ([]ImportantDate, error)
	UpdateImportantDate(ctx context.Context, chatID int64, dateID int64, req ImportantDateRequest) error
	UpdateImportantDateSharing(ctx context.Context, chatID int64, dateID int64, makeShared bool) error
	DeleteImportantDate(ctx context.Context, chatID int64, dateID int64) error
}

type AuthClient interface {
	GetAccessToken(ctx context.Context, userID int64) (string, error)
}
