package domain

import "errors"

var (
	ErrNoCompliments = errors.New("нет комплиментов")
	ErrLimitExceeded = errors.New("дневной лимит исчерпан")
)

type ErrBucketEmpty struct {
	Minutes int
}

func (e *ErrBucketEmpty) Error() string {
	return "ведро пустое"
}
