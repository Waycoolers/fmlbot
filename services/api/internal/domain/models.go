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

type ComplimentRequest struct {
	Text   string `db:"text"`
	IsSent bool   `db:"is_sent"`
}

type ComplimentResponse struct {
	Text      string    `db:"text"`
	IsSent    bool      `db:"is_sent"`
	CreatedAt time.Time `db:"created_at"`
}

type UserRequest struct {
	Username  string `db:"username"`
	PartnerID int64  `db:"partner_id"`
}

type UserResponse struct {
	ID        int64  `db:"id"`
	Username  string `db:"username"`
	PartnerID int64  `db:"partner_id"`
}

type UserConfig struct {
	ID                    int64     `db:"id"`
	TelegramID            int64     `db:"telegram_id"`
	DailyMessageTime      time.Time `db:"daily_message_time"`
	MaxComplimentCount    int       `db:"max_compliment_count"`
	ComplimentCount       int       `db:"compliment_count"`
	LastComplimentAt      time.Time `db:"last_compliment_at"`
	ComplimentTokenBucket int       `db:"compliment_token_bucket"`
	LastBucketUpdate      time.Time `db:"last_bucket_update"`
}

type UserConfigResponse struct {
	DailyMessageTime      time.Time `db:"daily_message_time"`
	MaxComplimentCount    int       `db:"max_compliment_count"`
	ComplimentCount       int       `db:"compliment_count"`
	LastComplimentAt      time.Time `db:"last_compliment_at"`
	ComplimentTokenBucket int       `db:"compliment_token_bucket"`
	LastBucketUpdate      time.Time `db:"last_bucket_update"`
}

type UserConfigPatch struct {
	MaxComplimentCount *int
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

type ImportantDateRequest struct {
	TelegramID       sql.NullInt64 `db:"telegram_id"`
	IsShared         bool          `db:"is_shared"`
	Title            string        `db:"title"`
	Date             time.Time     `db:"date"`
	IsActive         bool          `db:"is_active"`
	NotifyBeforeDays int           `db:"notify_before_days"`
}

type Repos struct {
	Users          UsersRepo
	UserConfig     UserConfigRepo
	Compliments    ComplimentsRepo
	ImportantDates ImportantDatesRepo
	Scheduler      SchedulerRepo
}

type ImportantDateMessage struct {
	ImportantDateID int64   `db:"important_date_id"`
	TgIDs           []int64 `db:"tg_ids"`
	Message         string  `db:"message"`
}
