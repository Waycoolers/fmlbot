package storage

import (
	"context"
)

func (s *Storage) GetTodayMessage(ctx context.Context) (string, error) {
	var message string
	query := `SELECT text FROM daily_messages WHERE day = CURRENT_DATE;`
	err := s.DB.QueryRowContext(ctx, query).Scan(&message)
	return message, err
}
