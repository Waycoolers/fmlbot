package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

func (s *usersRepo) GetTelegramID(ctx context.Context, userID int64) (int64, error) {
	var id int64
	err := s.db.GetContext(ctx, &id, `SELECT telegram_id FROM users WHERE user_id = $1 AND telegram_id IS NOT NULL`, userID)
	return id, err
}

func (s *usersRepo) GetTelegramUserID(ctx context.Context, telegramID int64) (int64, error) {
	var id int64
	err := s.db.GetContext(ctx, &id, `SELECT user_id FROM users WHERE telegram_id = $1`, telegramID)
	return id, err
}

func (s *usersRepo) SaveTelegramLink(ctx context.Context, userID int64, codeHash []byte, expiresAt time.Time) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO telegram_link_codes (user_id, code_hash, expires_at)
		SELECT user_id, $2, $3 FROM users WHERE user_id = $1 AND telegram_id IS NULL
		ON CONFLICT (user_id) DO UPDATE
		SET code_hash = EXCLUDED.code_hash, expires_at = EXCLUDED.expires_at
	`, userID, codeHash, expiresAt)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errs.ErrTelegramAlreadyLinked
	}
	return nil
}

func (s *usersRepo) UseTelegramLink(ctx context.Context, telegramID int64, codeHash []byte) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	err = tx.GetContext(ctx, &userID, `
		DELETE FROM telegram_link_codes
		WHERE code_hash = $1 AND expires_at > NOW()
		RETURNING user_id
	`, codeHash)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrTelegramLinkInvalid
	}
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE users SET telegram_id = $1 WHERE user_id = $2 AND telegram_id IS NULL
	`, telegramID, userID)
	if isUniqueViolation(err) {
		return errs.ErrTelegramAlreadyLinked
	}
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errs.ErrTelegramAlreadyLinked
	}
	return tx.Commit()
}
