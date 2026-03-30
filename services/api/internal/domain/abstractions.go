package domain

import (
	"context"
	"database/sql"
	"time"
)

type UsersRepo interface {
	AddUser(ctx context.Context, telegramID int64, username string) error
	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
	IsUserExists(ctx context.Context, userID int64) (bool, error)
	IsUserExistsByUsername(ctx context.Context, username string) (bool, error)
	SetPartner(ctx context.Context, telegramID int64, partnerID int64) error
	GetUsername(ctx context.Context, userID int64) (string, error)
	GetPartnerID(ctx context.Context, userID int64) (int64, error)
	SetPartners(ctx context.Context, userID int64, partnerID int64) error
	RemovePartners(ctx context.Context, userID int64, partnerID int64) error
	DeleteUser(ctx context.Context, userID int64) error
}

type UserConfigRepo interface {
	GetComplimentMaxCount(ctx context.Context, userID int64) (int, error)
	GetComplimentCount(ctx context.Context, userID int64) (int, error)
	SetComplimentMaxCount(ctx context.Context, userID int64, frequency int) error
	SetDefault(ctx context.Context, userID int64) error
}

type ComplimentsRepo interface {
	AddCompliment(ctx context.Context, telegramID int64, text string) (*Compliment, error)
	GetCompliments(ctx context.Context, telegramID int64) (compliments []Compliment, err error)
	DeleteCompliment(ctx context.Context, telegramID int64, complimentID int64) error
	MarkComplimentSent(ctx context.Context, complimentID int64) error
	AcquireCompliment(ctx context.Context, partnerID int64) (string, error)
}

type ImportantDatesRepo interface {
	AddImportantDate(ctx context.Context, telegramID sql.NullInt64, partnerID sql.NullInt64, title string, date time.Time, notifyBefore int) (*ImportantDate, error)
	GetImportantDates(ctx context.Context, telegramID sql.NullInt64) (importantDates []ImportantDate, err error)
	GetImportantDateByID(ctx context.Context, id int64) (importantDate ImportantDate, err error)
	DeleteImportantDate(ctx context.Context, id int64) error
	EditImportantDate(ctx context.Context, date ImportantDate) error
	GetAllActiveImportantDates(ctx context.Context) (importantDates []ImportantDate, err error)
	UpdateLastNotificationAt(ctx context.Context, id int64, timestamp time.Time) error
}

type SchedulerRepo interface {
	ClearComplimentsCount(ctx context.Context) error
	ClearComplimentTokenBucket(ctx context.Context) error
}
