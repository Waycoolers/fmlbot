package usecases

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) GetPlanOptions(ctx context.Context, userID int64, request plans.OptionsRequest) (plans.Options, error) {
	partnerID, version, err := uc.users.GetPair(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && partnerID == 0) {
		return plans.Options{}, errs.ErrPartnerNotFound
	}
	if err != nil {
		return plans.Options{}, err
	}
	if request.Pair.PartnerID != partnerID || request.Pair.Version != version {
		return plans.Options{}, errs.ErrPairChanged
	}
	options, err := uc.ai.GetPlanOptions(ctx, request.Prompt)
	if err != nil {
		return plans.Options{}, err
	}
	if !options.Valid() {
		return plans.Options{}, fmt.Errorf("AI returned invalid plan cards")
	}
	return options, nil
}

func (uc *UseCase) CreatePlan(ctx context.Context, userID int64, proposal plans.Proposal) (plans.Plan, error) {
	if _, err := hex.DecodeString(proposal.DraftID); err != nil || len(proposal.DraftID) != 16 ||
		!proposal.Pair.Valid() || proposal.Pair.PartnerID == userID || !proposal.Card.Valid() {
		return plans.Plan{}, errs.ErrBadRequest
	}
	return uc.plans.CreatePlan(ctx, userID, proposal)
}

func (uc *UseCase) GetPlan(ctx context.Context, userID, planID int64) (plans.Plan, error) {
	return uc.plans.GetPlan(ctx, userID, planID)
}

func (uc *UseCase) ListPlans(ctx context.Context, userID int64) ([]plans.Plan, error) {
	return uc.plans.ListPlans(ctx, userID, time.Now())
}

func (uc *UseCase) ChangePlan(ctx context.Context, userID, planID int64, change plans.Change) (plans.Plan, error) {
	if change.Revision < 1 {
		return plans.Plan{}, errs.ErrBadRequest
	}
	switch change.Action {
	case plans.Reschedule:
		if change.StartsAt == nil {
			return plans.Plan{}, errs.ErrPlanTime
		}
	case plans.Accept, plans.Decline, plans.Cancel:
		if change.StartsAt != nil {
			return plans.Plan{}, errs.ErrBadRequest
		}
	default:
		return plans.Plan{}, errs.ErrBadRequest
	}
	return uc.plans.ChangePlan(ctx, userID, planID, change)
}

func (uc *UseCase) DeliverPlanNotices(ctx context.Context, sender domain.PlanNoticeSender) error {
	jobs, err := uc.plans.DuePlanNotices(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := uc.plans.DeliverPlanNotice(ctx, job, time.Now(), func(ctx context.Context, notice plans.Notice) error {
			ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
			defer cancel()
			return sender.SendPlanNotice(ctx, notice)
		})
		if err != nil {
			slog.Warn("Plan notice delivery failed", "notice_id", job.ID, "error", err)
		}
	}
	return nil
}
