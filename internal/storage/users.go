package storage

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/models"
)

func (s *Storage) AddUser(ctx context.Context, telegramID int64, username string) error {
	_, err := s.DB.ExecContext(ctx, `
        INSERT INTO users (telegram_id, username)
        VALUES ($1, $2)
        ON CONFLICT (telegram_id) DO NOTHING;
    `, telegramID, username)
	return err
}

func (s *Storage) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx, `SELECT telegram_id FROM users WHERE LOWER(username)=LOWER($1)`, username).Scan(&id)
	return id, err
}

func (s *Storage) IsUserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id=$1)`, userID).Scan(&exists)
	return exists, err
}

func (s *Storage) IsUserExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username)=LOWER($1))`, username).Scan(&exists)
	return exists, err
}

func (s *Storage) SetPartner(ctx context.Context, telegramID int64, partnerUsername string) error {
	_, err := s.DB.ExecContext(ctx, `
        UPDATE users SET partner_username = $1 WHERE telegram_id = $2;
    `, partnerUsername, telegramID)
	return err
}

func (s *Storage) GetUsername(ctx context.Context, userID int64) (string, error) {
	var username string
	err := s.DB.QueryRowContext(ctx, `SELECT username FROM users WHERE telegram_id=$1;`, userID).Scan(&username)
	if err != nil {
		return "", nil
	}
	return username, nil
}

func (s *Storage) GetPartnerUsername(ctx context.Context, userID int64) (string, error) {
	var partner string
	query := `SELECT partner_username FROM users WHERE telegram_id = $1;`

	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&partner)
	if err != nil {
		// если нет записи, просто вернём пустую строку
		return "", nil
	}
	return partner, nil
}

func (s *Storage) SetUserState(ctx context.Context, userID int64, state models.State) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE users SET state=$1 WHERE telegram_id=$2;
	`, state, userID)
	return err
}

func (s *Storage) GetUserState(ctx context.Context, userID int64) (models.State, error) {
	var state models.State
	err := s.DB.QueryRowContext(ctx, `
		SELECT state FROM users WHERE telegram_id=$1;
	`, userID).Scan(&state)
	if err != nil {
		return models.Empty, err
	}
	return state, nil
}

func (s *Storage) SetPartners(ctx context.Context, userID, partnerID int64, userUsername, partnerUsername string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// user -> partner
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_username = $1 WHERE telegram_id = $2
    `, partnerUsername, userID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	// partner -> user
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_username = $1 WHERE telegram_id = $2
    `, userUsername, partnerID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	return tx.Commit()
}
