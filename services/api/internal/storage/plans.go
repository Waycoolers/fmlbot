package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type plansRepo struct {
	db *sqlx.DB
}

type planRow struct {
	plans.Plan
	PairVersion int64  `db:"pair_version"`
	CardData    []byte `db:"card"`
}

const planColumns = `p.id, p.owner_id, p.partner_id, p.pair_version, p.proposed_by,
	p.revision, p.status, p.starts_at, p.card`

const currentPlans = `SELECT ` + planColumns + `
	FROM couple_plans p
	JOIN users owner ON owner.user_id = p.owner_id AND owner.partner_id = p.partner_id AND owner.pair_version = p.pair_version
	JOIN users partner ON partner.user_id = p.partner_id AND partner.partner_id = p.owner_id AND partner.pair_version = p.pair_version
`

func (row planRow) decode() (plans.Plan, error) {
	err := json.Unmarshal(row.CardData, &row.Card)
	return row.Plan, err
}

func lockPlanPair(ctx context.Context, tx *sqlx.Tx, userID, partnerID, version int64) error {
	rows, err := lockPairUsers(ctx, tx, userID, partnerID)
	if err != nil {
		return err
	}
	if len(rows) != 2 {
		return errs.ErrPairChanged
	}
	for _, row := range rows {
		peerID := userID
		if row.UserID == userID {
			peerID = partnerID
		}
		if row.PartnerID != peerID || row.Version != version {
			return errs.ErrPairChanged
		}
	}
	return nil
}

func (r *plansRepo) CreatePlan(ctx context.Context, userID int64, proposal plans.Proposal) (plans.Plan, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return plans.Plan{}, err
	}
	defer tx.Rollback()

	partnerID, version := proposal.Pair.PartnerID, proposal.Pair.Version
	if err = lockPlanPair(ctx, tx, userID, partnerID, version); err != nil {
		return plans.Plan{}, err
	}
	var linked bool
	err = tx.GetContext(ctx, &linked, `SELECT telegram_id IS NOT NULL FROM users WHERE user_id = $1`, partnerID)
	if err != nil {
		return plans.Plan{}, err
	}
	if !linked {
		return plans.Plan{}, errs.ErrPlanPartnerTelegram
	}

	now := time.Now().UTC()
	if !plans.ValidStart(proposal.StartsAt, now) {
		return plans.Plan{}, errs.ErrPlanTime
	}
	card, err := json.Marshal(proposal.Card)
	if err != nil {
		return plans.Plan{}, err
	}
	var row planRow
	err = tx.GetContext(ctx, &row, `
		INSERT INTO couple_plans AS p (owner_id, partner_id, pair_version, draft_id, proposed_by, starts_at, card)
		VALUES ($1, $2, $3, $4, $1, $5, $6::jsonb)
		ON CONFLICT (owner_id, draft_id) DO NOTHING
		RETURNING `+planColumns, userID, partnerID, version, proposal.DraftID, proposal.StartsAt, string(card))
	if errors.Is(err, sql.ErrNoRows) {
		// Повтор запроса после потери ответа возвращает уже сохранённое предложение.
		err = tx.GetContext(ctx, &row, currentPlans+`WHERE p.owner_id = $1 AND p.draft_id = $2`, userID, proposal.DraftID)
		if errors.Is(err, sql.ErrNoRows) {
			return plans.Plan{}, errs.ErrPlanChanged
		}
		if err != nil {
			return plans.Plan{}, err
		}
		return row.decode()
	}
	if err != nil {
		return plans.Plan{}, err
	}
	plan := row.Plan
	plan.Card = proposal.Card
	if err = queuePlanNotice(ctx, tx, plan, partnerID, now, false); err != nil {
		return plans.Plan{}, err
	}
	return plan, tx.Commit()
}

func (r *plansRepo) GetPlan(ctx context.Context, userID, planID int64) (plans.Plan, error) {
	row, err := getPlanForUser(ctx, r.db, userID, planID)
	if err != nil {
		return plans.Plan{}, err
	}
	return row.decode()
}

func getPlanForUser(ctx context.Context, db sqlx.QueryerContext, userID, planID int64) (planRow, error) {
	var row planRow
	err := sqlx.GetContext(ctx, db, &row, currentPlans+`
		WHERE p.id = $1 AND $2 IN (p.owner_id, p.partner_id)
	`, planID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		err = errs.ErrPlanNotFound
	}
	return row, err
}

func (r *plansRepo) ListPlans(ctx context.Context, userID int64, now time.Time) ([]plans.Plan, error) {
	var rows []planRow
	err := r.db.SelectContext(ctx, &rows, currentPlans+`
		WHERE $1 IN (p.owner_id, p.partner_id) AND p.starts_at > $2 AND p.status IN ('pending', 'confirmed')
		ORDER BY p.starts_at, p.id
		LIMIT 10
	`, userID, now)
	if err != nil {
		return nil, err
	}
	result := make([]plans.Plan, 0, len(rows))
	for _, row := range rows {
		plan, err := row.decode()
		if err != nil {
			return nil, err
		}
		result = append(result, plan)
	}
	return result, nil
}

func (r *plansRepo) ChangePlan(ctx context.Context, userID, planID int64, change plans.Change) (plans.Plan, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return plans.Plan{}, err
	}
	defer tx.Rollback()

	row, err := getPlanForUser(ctx, tx, userID, planID)
	if err != nil {
		return plans.Plan{}, err
	}
	if err = lockPlanPair(ctx, tx, row.OwnerID, row.PartnerID, row.PairVersion); err != nil {
		return plans.Plan{}, err
	}
	now := time.Now().UTC()
	if row.Revision != change.Revision || !row.StartsAt.After(now) {
		return plans.Plan{}, errs.ErrPlanChanged
	}
	if !row.Can(userID, change.Action) {
		return plans.Plan{}, errs.ErrForbidden
	}

	plan, err := row.decode()
	if err != nil {
		return plans.Plan{}, err
	}
	switch change.Action {
	case plans.Accept:
		plan.Status = plans.Confirmed
	case plans.Reschedule:
		if change.StartsAt == nil || !plans.ValidStart(*change.StartsAt, now) || plan.StartsAt.Equal(*change.StartsAt) {
			return plans.Plan{}, errs.ErrPlanTime
		}
		plan.StartsAt = *change.StartsAt
		plan.ProposedBy = userID
	case plans.Decline:
		plan.Status = plans.Declined
	case plans.Cancel:
		plan.Status = plans.Cancelled
	}
	err = tx.GetContext(ctx, &plan.Revision, `
		UPDATE couple_plans SET status = $1, starts_at = $2, proposed_by = $3, revision = revision + 1
		WHERE id = $4 AND revision = $5 AND starts_at > clock_timestamp()
		RETURNING revision
	`, plan.Status, plan.StartsAt, plan.ProposedBy, plan.ID, change.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return plans.Plan{}, errs.ErrPlanChanged
	}
	if err != nil {
		return plans.Plan{}, err
	}

	peerID := plan.OwnerID
	if userID == plan.OwnerID {
		peerID = plan.PartnerID
	}
	if err = queuePlanNotice(ctx, tx, plan, peerID, now, false); err != nil {
		return plans.Plan{}, err
	}
	if plan.Status == plans.Confirmed {
		remindAt := plan.StartsAt.Add(-time.Hour)
		if !remindAt.After(now) {
			remindAt = now.Add(plan.StartsAt.Sub(now) / 2)
		}
		for _, recipientID := range []int64{plan.OwnerID, plan.PartnerID} {
			if err = queuePlanNotice(ctx, tx, plan, recipientID, remindAt, true); err != nil {
				return plans.Plan{}, err
			}
		}
	}
	return plan, tx.Commit()
}

func queuePlanNotice(ctx context.Context, tx *sqlx.Tx, plan plans.Plan, recipientID int64, dueAt time.Time, reminder bool) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO couple_plan_notices (plan_id, revision, recipient_id, due_at, reminder)
		VALUES ($1, $2, $3, $4, $5)
	`, plan.ID, plan.Revision, recipientID, dueAt, reminder)
	return err
}

func (r *plansRepo) DuePlanNotices(ctx context.Context, now time.Time) ([]domain.PlanNoticeJob, error) {
	var jobs []domain.PlanNoticeJob
	err := r.db.SelectContext(ctx, &jobs, `
		SELECT id, plan_id FROM couple_plan_notices
		WHERE processed_at IS NULL AND due_at <= $1
		ORDER BY due_at, id
		LIMIT 25
	`, now)
	return jobs, err
}

func (r *plansRepo) DeliverPlanNotice(ctx context.Context, job domain.PlanNoticeJob, now time.Time, send func(context.Context, plans.Notice) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Удаление и изменение плана берут блокировку плана раньше, чем его уведомлений.
	var row planRow
	err = tx.GetContext(ctx, &row, currentPlans+`WHERE p.id = $1 FOR SHARE OF p`, job.PlanID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	activePair := err == nil

	var notice struct {
		Revision    int64         `db:"revision"`
		RecipientID int64         `db:"recipient_id"`
		TelegramID  sql.NullInt64 `db:"telegram_id"`
		Reminder    bool          `db:"reminder"`
	}
	err = tx.GetContext(ctx, &notice, `
		SELECT n.revision, n.recipient_id, u.telegram_id, n.reminder
		FROM couple_plan_notices n JOIN users u ON u.user_id = n.recipient_id
		WHERE n.id = $1 AND n.plan_id = $2 AND n.processed_at IS NULL AND n.due_at <= $3
		FOR UPDATE OF n SKIP LOCKED
	`, job.ID, job.PlanID, now)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var deliveryErr error
	participant := notice.RecipientID == row.OwnerID || notice.RecipientID == row.PartnerID
	if activePair && participant && notice.Revision == row.Revision && row.StartsAt.After(now) && notice.TelegramID.Valid {
		plan, err := row.decode()
		if err != nil {
			return err
		}
		deliveryErr = send(ctx, plans.Notice{
			ChatID:   notice.TelegramID.Int64,
			View:     plans.View{Plan: plan, ViewerID: notice.RecipientID},
			Reminder: notice.Reminder,
		})
	}
	if deliveryErr != nil && !errors.Is(deliveryErr, errs.ErrRecipientUnavailable) {
		_, err = tx.ExecContext(ctx, `UPDATE couple_plan_notices SET due_at = $2 WHERE id = $1`, job.ID, now.Add(time.Minute))
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE couple_plan_notices SET processed_at = $2 WHERE id = $1`, job.ID, now)
		deliveryErr = nil
	}
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return deliveryErr
}
