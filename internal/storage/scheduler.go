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

func (s *Storage) ClearComplimentTime(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE user_config SET last_compliment_at=null WHERE TRUE
	`)
	if err != nil {
		return err
	}
	return nil
}
