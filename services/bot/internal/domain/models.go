package domain

import (
	"database/sql"
	"time"
)

type Message struct {
	ChatID    int64
	UserName  string
	FirstName string
	Text      string
}

type CallbackQuery struct {
	ChatID    int64
	MessageID int
	Data      string
	UserName  string
	Message   string
}

type InlineKeyboard struct {
	Rows []InlineKeyboardRow
}

type InlineKeyboardRow struct {
	Buttons []InlineKeyboardButton
}

type InlineKeyboardButton struct {
	Text string
	Data string
}

type Keyboard struct {
	Rows []KeyboardRow
}

type KeyboardRow struct {
	Buttons []KeyboardButton
}

type KeyboardButton struct {
	Command Command
}

type Update struct {
	Message       *Message
	CallbackQuery *CallbackQuery
}

type Compliment struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	IsSent    bool      `json:"is_sent"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	PartnerID int64  `json:"partner_id"`
}

type ImportantDate struct {
	ID                 int64         `json:"id"`
	UserID             sql.NullInt64 `json:"user_id"`
	PartnerID          sql.NullInt64 `json:"partner_id"`
	Title              string        `json:"title"`
	Date               time.Time     `json:"date"`
	IsActive           bool          `json:"is_active"`
	LastNotificationAt sql.NullTime  `json:"last_notification_at"`
	NotifyBeforeDays   int           `json:"notify_before_days"`
	CreatedAt          time.Time     `json:"created_at"`
}

type UserConfig struct {
	DailyMessageTime   time.Time `json:"daily_message_time"`
	MaxComplimentCount int       `json:"max_compliment_count"`
	ComplimentCount    int       `json:"compliment_count"`
}

type ImportantDateRequest struct {
	UserID           sql.NullInt64 `json:"user_id"`
	IsShared         bool          `json:"is_shared"`
	Title            string        `json:"title"`
	Date             time.Time     `json:"date"`
	IsActive         bool          `json:"is_active"`
	NotifyBeforeDays int           `json:"notify_before_days"`
}

type ImportantDateMessage struct {
	ImportantDateID int64   `json:"important_date_id"`
	UserIDs         []int64 `json:"user_ids"`
	Message         string  `json:"message"`
}
