package storage

import "context"

func (s *Storage) ClearPartnerReferences(ctx context.Context, userID int64) error {
	var username string
	err := s.DB.QueryRowContext(ctx, `
		SELECT username FROM users WHERE telegram_id=$1
	`, userID).Scan(&username)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE users SET partner_username=NULL
		WHERE partner_username=$1
	`, username)
	return err
}

func (s *Storage) DeleteUser(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx, `
		DELETE FROM users WHERE telegram_id=$1
	`, userID)
	return err
}
