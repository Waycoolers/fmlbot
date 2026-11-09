package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/plans"
)

type UsersRepo interface {
	AddUser(ctx context.Context, userID int64, username string, password []byte, telegramID sql.NullInt64) error
	GetTelegramID(ctx context.Context, userID int64) (int64, error)
	GetTelegramUserID(ctx context.Context, telegramID int64) (int64, error)
	SaveTelegramLink(ctx context.Context, userID int64, codeHash []byte, expiresAt time.Time) error
	UseTelegramLink(ctx context.Context, telegramID int64, codeHash []byte) error
	SetPassword(ctx context.Context, userID int64, password []byte) error
	GetUserPasswordHash(ctx context.Context, userID int64) ([]byte, error)
	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
	IsUserExists(ctx context.Context, userID int64) (bool, error)
	IsUserExistsByUsername(ctx context.Context, username string) (bool, error)
	GetUsername(ctx context.Context, userID int64) (string, error)
	GetPartnerID(ctx context.Context, userID int64) (int64, error)
	GetPair(ctx context.Context, userID int64) (partnerID, version int64, err error)
	CreatePartnerRequest(ctx context.Context, userID int64, partnerID int64) (int64, error)
	AcceptPartnerRequest(ctx context.Context, requestID int64, userID int64) (int64, error)
	DeclinePartnerRequest(ctx context.Context, requestID int64, userID int64) (int64, error)
	GetPartnerRequests(ctx context.Context, userID int64) ([]PartnerRequest, error)
	RemovePartners(ctx context.Context, userID, partnerID, version int64) error
	DeleteUser(ctx context.Context, userID, partnerID, version int64) error
	UpdateUsername(ctx context.Context, userID int64, username string) error
}

type UserConfigRepo interface {
	GetUserConfig(ctx context.Context, userID int64) (*UserConfig, error)
	GetComplimentMaxCount(ctx context.Context, userID int64) (int, error)
	GetComplimentCount(ctx context.Context, userID int64) (int, error)
	SetComplimentMaxCount(ctx context.Context, userID int64, frequency int) error
	SetDefault(ctx context.Context, userID int64) error
}

type ComplimentsRepo interface {
	AddCompliment(ctx context.Context, userID int64, text string) (*Compliment, error)
	GetCompliments(ctx context.Context, userID int64) (compliments []Compliment, err error)
	GetReceivedCompliments(ctx context.Context, userID int64) (compliments []Compliment, err error)
	UpdateCompliment(ctx context.Context, userID int64, complimentID int64, text string, isSent bool) error
	DeleteCompliment(ctx context.Context, userID int64, complimentID int64) error
	MarkComplimentSent(ctx context.Context, complimentID int64) error
	AcquireCompliment(ctx context.Context, userID, partnerID, version int64, deliver func(context.Context, string) error) (string, error)
}

type ImportantDatesRepo interface {
	AddImportantDate(ctx context.Context, userID int64, partnerID sql.NullInt64, pairVersion int64, title string, date time.Time, notifyBefore int) (*ImportantDate, error)
	GetImportantDates(ctx context.Context, userID int64) (importantDates []ImportantDate, err error)
	GetImportantDateByID(ctx context.Context, id int64, userID int64) (*ImportantDate, error)
	DeleteImportantDate(ctx context.Context, id int64, userID int64) error
	EditImportantDate(ctx context.Context, id int64, userID int64, date ImportantDateRequest) error
	GetAllActiveImportantDates(ctx context.Context) (importantDates []ImportantDate, err error)
	MakeImportantDatePrivate(ctx context.Context, dateID int64, userID int64) error
	MakeImportantDateShared(ctx context.Context, dateID, userID, partnerID, pairVersion int64) error
}

type FCMRepo interface {
	SetFCMToken(ctx context.Context, userID int64, token string) error
	GetFCMToken(ctx context.Context, userID int64) (string, error)
}

type SchedulerRepo interface {
	QueueImportantDateDelivery(ctx context.Context, dateID, userID int64, dueOn time.Time, channel, message string) error
	PendingImportantDateDeliveries(ctx context.Context, pushEnabled bool) ([]ImportantDateDelivery, error)
	DeliverImportantDate(ctx context.Context, delivery ImportantDateDelivery, send func(context.Context) error) error
	DoMidnightTasksWithCompliments(ctx context.Context) error
	GetLastRun(ctx context.Context, name string) (time.Time, bool, error)
	SetLastRun(ctx context.Context, name string, at time.Time) error
}

type Sender interface {
	SendMessage(ctx context.Context, update MessageRequest) error
	SendImportantDatesNotification(ctx context.Context, update ImportantDateMessage) error
}

type PlanNoticeSender interface {
	SendPlanNotice(ctx context.Context, notice plans.Notice) error
}

type PlansRepo interface {
	CreatePlan(ctx context.Context, userID int64, proposal plans.Proposal) (plans.Plan, error)
	GetPlan(ctx context.Context, userID, planID int64) (plans.Plan, error)
	ListPlans(ctx context.Context, userID int64, now time.Time) ([]plans.Plan, error)
	ChangePlan(ctx context.Context, userID, planID int64, change plans.Change) (plans.Plan, error)
	DuePlanNotices(ctx context.Context, now time.Time) ([]PlanNoticeJob, error)
	DeliverPlanNotice(ctx context.Context, job PlanNoticeJob, now time.Time, send func(context.Context, plans.Notice) error) error
}

type PlanNoticeJob struct {
	ID     int64 `db:"id"`
	PlanID int64 `db:"plan_id"`
}

type AuthClient interface {
	RevokeAllTokens(ctx context.Context, userID int64) error
}

type AIClient interface {
	GetDateIdea(ctx context.Context, title string, days int, context string) (text string, err error)
	GetLeisureIdea(ctx context.Context, location, activity, budget, extra string) (string, error)
	GetPlanOptions(ctx context.Context, prompt plans.Prompt) (plans.Options, error)
}
