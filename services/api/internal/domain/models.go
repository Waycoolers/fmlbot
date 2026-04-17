package domain

import (
	"database/sql"
	"time"
)

type Compliment struct {
	ID        int64     `db:"id" json:"id"`
	Text      string    `db:"text" json:"text"`
	IsSent    bool      `db:"is_sent" json:"is_sent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ComplimentRequest struct {
	Text   string `db:"text" json:"text"`
	IsSent bool   `db:"is_sent" json:"is_sent"`
}

type ComplimentResponse struct {
	Text      string    `db:"text" json:"text"`
	IsSent    bool      `db:"is_sent" json:"is_sent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UserRequest struct {
	Username  string `db:"username" json:"username"`
	PartnerID int64  `db:"partner_id" json:"partner_id"`
}

type UserResponse struct {
	ID        int64  `db:"user_id" json:"user_id"`
	Username  string `db:"username" json:"username"`
	PartnerID int64  `db:"partner_id" json:"partner_id"`
}

type UserConfig struct {
	DailyMessageTime   time.Time `db:"daily_message_time" json:"daily_message_time"`
	MaxComplimentCount int       `db:"max_compliment_count" json:"max_compliment_count"`
	ComplimentCount    int       `db:"compliment_count" json:"compliment_count"`
}

type UserConfigResponse struct {
	DailyMessageTime      time.Time `db:"daily_message_time" json:"daily_message_time"`
	MaxComplimentCount    int       `db:"max_compliment_count" json:"max_compliment_count"`
	ComplimentCount       int       `db:"compliment_count" json:"compliment_count"`
	LastComplimentAt      time.Time `db:"last_compliment_at" json:"last_compliment_at"`
	ComplimentTokenBucket int       `db:"compliment_token_bucket" json:"compliment_token_bucket"`
	LastBucketUpdate      time.Time `db:"last_bucket_update" json:"last_bucket_update"`
}

type UserConfigPatch struct {
	MaxComplimentCount *int `json:"max_compliment_count"`
}

type ImportantDate struct {
	ID                 int64         `db:"id" json:"id"`
	UserID             sql.NullInt64 `db:"user_id" json:"user_id"`
	PartnerID          sql.NullInt64 `db:"partner_id" json:"partner_id"`
	Title              string        `db:"title" json:"title"`
	Date               time.Time     `db:"date" json:"date"`
	IsActive           bool          `db:"is_active" json:"is_active"`
	LastNotificationAt sql.NullTime  `db:"last_notification_at" json:"last_notification_at"`
	NotifyBeforeDays   int           `db:"notify_before_days" json:"notify_before_days"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
}

type ImportantDateRequest struct {
	UserID           sql.NullInt64 `db:"user_id" json:"user_id"`
	IsShared         bool          `db:"is_shared" json:"is_shared"`
	Title            string        `db:"title" json:"title"`
	Date             time.Time     `db:"date" json:"date"`
	IsActive         bool          `db:"is_active" json:"is_active"`
	NotifyBeforeDays int           `db:"notify_before_days" json:"notify_before_days"`
}

type Repos struct {
	Users          UsersRepo
	UserConfig     UserConfigRepo
	Compliments    ComplimentsRepo
	ImportantDates ImportantDatesRepo
	Scheduler      SchedulerRepo
}

type ImportantDateMessage struct {
	ImportantDateID int64   `db:"important_date_id" json:"important_date_id"`
	UserIDs         []int64 `db:"user_ids" json:"user_ids"`
	Message         string  `db:"message" json:"message"`
}
