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
