package domain

import (
	"database/sql"
	"time"
)

type ImportantDateDraft struct {
	Title            string
	Year             int
	Month            int
	Day              int
	PartnerID        sql.NullInt64
	NotifyBeforeDays int
	CreatedAt        time.Time
}

type ImportantDateEditDraft struct {
	ImportantDateID int64     `json:"important_date_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type IdeaDraft struct {
	Location  string
	Activity  string
	Budget    string
	CreatedAt time.Time
}
