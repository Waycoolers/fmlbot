package storage

import (
	"context"
	"database/sql"
	"time"

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

func (s *userConfigRepo) SetComplimentCount(ctx context.Context, userID int64, value int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=$1 WHERE telegram_id = $2;
	`, value, userID)
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

func (s *userConfigRepo) SetComplimentTime(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET last_compliment_at=(NOW() AT TIME ZONE 'UTC') WHERE telegram_id = $1;
	`, userID)
	return err
}

func (s *userConfigRepo) TakeComplimentFromBucket(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=compliment_token_bucket - 1
		                   WHERE telegram_id = $1
		                     AND compliment_token_bucket > 0;
	`, userID)
	return err
}

func (s *userConfigRepo) UpdateComplimentBucket(ctx context.Context, userID int64, value int, time time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=$1, last_bucket_update=$2 WHERE telegram_id = $3;
	`, value, time, userID)
	return err
}

func (s *userConfigRepo) GetComplimentsBucket(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT compliment_token_bucket FROM user_config WHERE telegram_id = $1;
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *userConfigRepo) GetLastBucketUpdate(ctx context.Context, userID int64) (time.Time, error) {
	var lastBucketUpdate time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT last_bucket_update FROM user_config WHERE telegram_id = $1;
	`, userID).Scan(&lastBucketUpdate)
	if err != nil {
		return time.Time{}, err
	}
	return lastBucketUpdate, nil
}

func (s *userConfigRepo) GetComplimentTime(ctx context.Context, userID int64) (time.Time, error) {
	var timestamp sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT last_compliment_at FROM user_config WHERE telegram_id = $1;
	`, userID).Scan(&timestamp)
	if err != nil {
		return time.Time{}, err
	}

	if timestamp.Valid {
		return timestamp.Time, nil
	}
	return time.Time{}, nil
}
