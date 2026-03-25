package domain

type BotClient interface {
	SendMessage(chatID int64, text string) error
	SendWithInlineKeyboard(chatID int64, text string, keyboard InlineKeyboard) error
	EditMessageReplyMarkup(chatID int64, messageID int, keyboard InlineKeyboard) error
	GetUpdatesChan() <-chan Update
	StopReceivingUpdates()
	SendKeyboard(chatID int64, text string, keyboard Keyboard) (Message, error)
	DeleteMessage(chatID int64, messageID int) error
}
