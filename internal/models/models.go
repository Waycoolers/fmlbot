package models

import "time"

type Compliment struct {
	ID        int64     `db:"id"`
	Text      string    `db:"text"`
	IsSent    bool      `db:"is_sent"`
	CreatedAt time.Time `db:"created_at"`
}

type User struct {
	telegramID int64
	state      State
	username   string
	partnerID  int64
}
