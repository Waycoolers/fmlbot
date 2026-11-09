package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
