package storage

import "context"

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
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE telegram_id = $1;
	`, userID)
	return err
}
