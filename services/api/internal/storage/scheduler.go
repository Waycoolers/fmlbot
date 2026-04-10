package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type schedulerRepo struct {
	db *sqlx.DB
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
