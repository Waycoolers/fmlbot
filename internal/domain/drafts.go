package domain

import (
	"database/sql"
	"time"
)

type ImportantDateDraft struct {
	Title            string
	Date             time.Time
	PartnerID        sql.NullInt64
	NotifyBeforeDays int
	CreatedAt        time.Time
}
