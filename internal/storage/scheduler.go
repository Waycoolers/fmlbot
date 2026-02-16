package storage

import "context"

func (s *Storage) ClearComplimentsCount(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET compliment_count=0 WHERE TRUE
	`)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) ClearComplimentTokenBucket(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET compliment_token_bucket=2 WHERE TRUE
	`)
	if err != nil {
		return err
	}
	return nil
}
