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
