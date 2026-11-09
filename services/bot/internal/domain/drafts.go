package domain

import (
	"database/sql"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/plans"
)

type ImportantDateDraft struct {
	Session     string
	Title       string
	Year        int
	Month       int
	Day         int
	PartnerID   sql.NullInt64
	PairVersion *int64
	CreatedAt   time.Time
}

type ImportantDateEditDraft struct {
	Session         string
	Action          string
	PairVersion     *int64
	ImportantDateID int64     `json:"important_date_id"`
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	CreatedAt       time.Time `json:"created_at"`
}

type IdeaDraft struct {
	Session   string
	Step      string
	Location  string
	Activity  string
	CreatedAt time.Time
	Plan      *PlanDraft `json:",omitempty"`
}

type PlanDraft struct {
	Pair           plans.Pair
	Brief          string
	Cards          []plans.Card
	Selected       int
	TargetID       int64
	TargetRevision int64
	PickerSession  string
	PickerTime     time.Time
}
