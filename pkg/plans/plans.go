package plans

import (
	"strings"
	"time"
	"unicode/utf8"
)

const MaxBriefLength = 600

type Pair struct {
	PartnerID int64 `json:"partner_id"`
	Version   int64 `json:"version"`
}

func (p Pair) Valid() bool {
	return p.PartnerID > 0 && p.Version >= 0
}

type Card struct {
	Title           string   `json:"title"`
	DurationMinutes int      `json:"duration_minutes"`
	EstimatedCost   string   `json:"estimated_cost"`
	Preparation     string   `json:"preparation"`
	Steps           []string `json:"steps"`
	Backup          string   `json:"backup"`
}

func (c Card) Valid() bool {
	if !validText(c.Title, 80) || !validText(c.EstimatedCost, 60) ||
		!validText(c.Preparation, 220) || !validText(c.Backup, 220) ||
		c.DurationMinutes < 1 || c.DurationMinutes > 1440 || len(c.Steps) < 2 || len(c.Steps) > 4 {
		return false
	}
	for _, step := range c.Steps {
		if !validText(step, 160) {
			return false
		}
	}
	return true
}

func ValidBrief(brief string) bool {
	return validText(brief, MaxBriefLength)
}

func ValidStart(startsAt, now time.Time) bool {
	return startsAt.After(now) && !startsAt.After(now.AddDate(1, 0, 0))
}

func validText(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && utf8.RuneCountInString(value) <= limit
}

type Prompt struct {
	Brief      string `json:"brief"`
	Adjustment string `json:"adjustment,omitempty"`
	Reference  *Card  `json:"reference,omitempty"`
}

func (p Prompt) Valid() bool {
	if !ValidBrief(p.Brief) {
		return false
	}
	switch p.Adjustment {
	case "":
		return p.Reference == nil
	case "simpler", "cheaper", "more_involved", "more_expensive":
		return p.Reference != nil && p.Reference.Valid()
	default:
		return false
	}
}

type OptionsRequest struct {
	Pair Pair `json:"pair"`
	Prompt
}

type Options struct {
	Cards []Card `json:"cards"`
}

func (o Options) Valid() bool {
	return len(o.Cards) == 2 && o.Cards[0].Valid() && o.Cards[1].Valid() &&
		!strings.EqualFold(strings.TrimSpace(o.Cards[0].Title), strings.TrimSpace(o.Cards[1].Title))
}

type Status string

const (
	Pending   Status = "pending"
	Confirmed Status = "confirmed"
	Declined  Status = "declined"
	Cancelled Status = "cancelled"
)

type Action string

const (
	Accept     Action = "accept"
	Reschedule Action = "reschedule"
	Decline    Action = "decline"
	Cancel     Action = "cancel"
)

type Plan struct {
	ID         int64     `db:"id" json:"id"`
	OwnerID    int64     `db:"owner_id" json:"owner_id"`
	PartnerID  int64     `db:"partner_id" json:"partner_id"`
	ProposedBy int64     `db:"proposed_by" json:"proposed_by"`
	Revision   int64     `db:"revision" json:"revision"`
	Status     Status    `db:"status" json:"status"`
	StartsAt   time.Time `db:"starts_at" json:"starts_at"`
	Card       Card      `db:"-" json:"card"`
}

func (p Plan) Can(userID int64, action Action) bool {
	if userID != p.OwnerID && userID != p.PartnerID {
		return false
	}
	switch action {
	case Accept, Reschedule, Decline:
		return p.Status == Pending && userID != p.ProposedBy
	case Cancel:
		return p.Status == Confirmed || (p.Status == Pending && userID == p.ProposedBy)
	default:
		return false
	}
}

type View struct {
	Plan
	ViewerID int64 `json:"viewer_id"`
}

type Proposal struct {
	Pair     Pair      `json:"pair"`
	DraftID  string    `json:"draft_id"`
	StartsAt time.Time `json:"starts_at"`
	Card     Card      `json:"card"`
}

type Change struct {
	Action   Action     `json:"action"`
	Revision int64      `json:"revision"`
	StartsAt *time.Time `json:"starts_at,omitempty"`
}

type Notice struct {
	ChatID   int64 `json:"chat_id"`
	View     View  `json:"view"`
	Reminder bool  `json:"reminder"`
}
