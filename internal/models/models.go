package models

import "time"

type Compliment struct {
	id        int64
	text      string
	isSent    bool
	createdAt time.Time
}

type User struct {
	telegramID      int64
	state           State
	username        string
	partnerUsername string
}
