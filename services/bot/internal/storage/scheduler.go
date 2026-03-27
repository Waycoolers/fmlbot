package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type schedulerRepo struct {
	db *sqlx.DB
}

func (s *schedulerRepo) ClearComplimentsCount(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE TRUE
	`)
	if err != nil {
		return err
	}
	return nil
}

func (s *schedulerRepo) ClearComplimentTokenBucket(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=2 WHERE TRUE
	`)
	if err != nil {
		return err
	}
	return nil
}
