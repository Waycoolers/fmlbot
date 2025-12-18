package domain

import (
	"database/sql"
	"time"
)

type Compliment struct {
	ID        int64     `db:"id"`
	Text      string    `db:"text"`
	IsSent    bool      `db:"is_sent"`
	CreatedAt time.Time `db:"created_at"`
}

type User struct {
	telegramID int64  `db:"telegram_id"`
	state      State  `db:"state"`
	username   string `db:"username"`
	partnerID  int64  `db:"partner_id"`
}

type ImportantDate struct {
	ID                 int64         `db:"id"`
	TelegramID         sql.NullInt64 `db:"telegram_id"`
	PartnerID          sql.NullInt64 `db:"partner_id"`
	Title              string        `db:"title"`
	Date               time.Time     `db:"date"`
	IsActive           bool          `db:"is_active"`
	LastNotificationAt sql.NullTime  `db:"last_notification_at"`
	NotifyBeforeDays   int           `db:"notify_before_days"`
	CreatedAt          time.Time     `db:"created_at"`
}
