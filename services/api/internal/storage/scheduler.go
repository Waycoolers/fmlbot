package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type schedulerRepo struct {
	db *sqlx.DB
}

func (s *schedulerRepo) GetLastRun(ctx context.Context, name string) (time.Time, bool, error) {
	var lastRun time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT last_run_at FROM cron_tasks WHERE name = $1
	`, name).Scan(&lastRun)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, nil
	}
	return lastRun, err == nil, err
}

func (s *schedulerRepo) SetLastRun(ctx context.Context, name string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cron_tasks (name, last_run_at)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET last_run_at = EXCLUDED.last_run_at
	`, name, at)
	return err
}

func (s *schedulerRepo) DoMidnightTasksWithCompliments(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE TRUE
	`)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=2, last_bucket_update=$1 WHERE TRUE
	`, now)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	return tx.Commit()
}

func (s *schedulerRepo) QueueImportantDateDelivery(ctx context.Context, dateID, userID int64, dueOn time.Time, channel, message string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO important_date_deliveries (date_id, user_id, due_on, channel, message)
		SELECT id, $2, $3, $4, $5 FROM important_dates
		WHERE id = $1 AND is_active AND (user_id = $2 OR partner_id = $2)
		ON CONFLICT (date_id, user_id, due_on, channel) DO NOTHING
	`, dateID, userID, dueOn.Format("2006-01-02"), channel, message)
	return err
}

func (s *schedulerRepo) PendingImportantDateDeliveries(ctx context.Context, pushEnabled bool) ([]domain.ImportantDateDelivery, error) {
	var pending []domain.ImportantDateDelivery
	err := s.db.SelectContext(ctx, &pending, `
		SELECT id, date_id, user_id, channel, message
		FROM important_date_deliveries
		WHERE processed_at IS NULL AND (channel <> 'push' OR $1)
		ORDER BY id
	`, pushEnabled)
	return pending, err
}

func (s *schedulerRepo) DeliverImportantDate(ctx context.Context, delivery domain.ImportantDateDelivery, send func(context.Context) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the date first, matching deletion and the cascading removal of deliveries.
	var allowed bool
	err = tx.GetContext(ctx, &allowed, `
		SELECT is_active AND COALESCE(user_id = $2 OR partner_id = $2, false)
		FROM important_dates WHERE id = $1 FOR SHARE
	`, delivery.DateID, delivery.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	var id int64
	err = tx.GetContext(ctx, &id, `
		SELECT id FROM important_date_deliveries
		WHERE id = $1 AND processed_at IS NULL FOR UPDATE SKIP LOCKED
	`, delivery.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	delivered := false
	if allowed {
		err = send(ctx)
		if err != nil && !errors.Is(err, errs.ErrRecipientUnavailable) {
			return err
		}
		delivered = err == nil
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE important_date_deliveries
		SET processed_at = NOW(), delivered_at = CASE WHEN $2 THEN NOW() END, message = ''
		WHERE id = $1
	`, id, delivered)
	if err != nil {
		return err
	}
	return tx.Commit()
}
