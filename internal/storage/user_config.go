package storage

import (
	"context"
	"database/sql"
	"time"
)

func (s *Storage) GetComplimentMaxCount(ctx context.Context, userID int64) (int, error) {
	var frequency int
	err := s.DB.QueryRowContext(ctx, `
		SELECT max_compliment_count FROM user_config WHERE telegram_id=$1;
	`, userID).Scan(&frequency)
	if err != nil {
		return 0, nil
	}
	return frequency, nil
}

func (s *Storage) GetComplimentCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `
		SELECT compliment_count FROM user_config WHERE telegram_id=$1;
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Storage) SetComplimentMaxCount(ctx context.Context, userID int64, frequency int) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET max_compliment_count=$1 WHERE telegram_id = $2;
	`, frequency, userID)
	return err
}

func (s *Storage) SetComplimentCount(ctx context.Context, userID int64, value int) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=$1 WHERE telegram_id = $2;
	`, value, userID)
	return err
}

func (s *Storage) SetDefault(ctx context.Context, userID int64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE telegram_id = $1;
	`, userID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE user_config SET last_compliment_at=null WHERE telegram_id = $1;
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

func (s *Storage) SetComplimentTime(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET last_compliment_at=(NOW() AT TIME ZONE 'UTC') WHERE telegram_id = $1;
	`, userID)
	return err
}

func (s *Storage) GetComplimentTime(ctx context.Context, userID int64) (time.Time, error) {
	var timestamp sql.NullTime
	err := s.DB.QueryRowContext(ctx, `
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
