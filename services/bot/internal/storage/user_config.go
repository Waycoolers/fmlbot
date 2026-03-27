package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type userConfigRepo struct {
	db *sqlx.DB
}

func (s *userConfigRepo) GetComplimentMaxCount(ctx context.Context, userID int64) (int, error) {
	var frequency int
	err := s.db.QueryRowContext(ctx, `
		SELECT max_compliment_count FROM user_config WHERE telegram_id=$1;
	`, userID).Scan(&frequency)
	if err != nil {
		return 0, err
	}
	return frequency, nil
}

func (s *userConfigRepo) GetComplimentCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT compliment_count FROM user_config WHERE telegram_id=$1;
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *userConfigRepo) SetComplimentMaxCount(ctx context.Context, userID int64, frequency int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET max_compliment_count=$1 WHERE telegram_id = $2;
	`, frequency, userID)
	return err
}

func (s *userConfigRepo) SetDefault(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE telegram_id = $1;
	`, userID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=2, last_bucket_update=now() WHERE telegram_id = $1;
	`, userID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	return tx.Commit()
}
