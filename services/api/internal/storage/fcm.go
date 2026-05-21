package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type fcmRepo struct {
	db *sqlx.DB
}

func (s *fcmRepo) SetFCMToken(ctx context.Context, userID int64, token string) error {
	_, err := s.db.ExecContext(ctx, `
       INSERT INTO user_devices (user_id, fcm_token) 
       VALUES ($1, $2) 
       ON CONFLICT (user_id) 
       DO UPDATE SET 
           fcm_token = EXCLUDED.fcm_token, 
           created_at = NOW();
    `, userID, token)
	return err
}

func (s *fcmRepo) GetFCMToken(ctx context.Context, userID int64) (string, error) {
	var token string
	err := s.db.QueryRowContext(ctx, `
		SELECT fcm_token FROM user_devices WHERE user_id = $1 LIMIT 1;
	`, userID).Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}
