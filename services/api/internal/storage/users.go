package storage

import (
	"context"
	"database/sql"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/jmoiron/sqlx"
)

type usersRepo struct {
	db *sqlx.DB
}

func (s *usersRepo) AddUser(ctx context.Context, userID int64, username string, password []byte) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	res, err := tx.ExecContext(ctx, `
        INSERT INTO users (user_id, username, password_hash)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO NOTHING;
    `, userID, username, password)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}
	if aff != 1 {
		return errs.ErrUserExists
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO user_config (user_id)
        VALUES ($1)
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

func (s *usersRepo) SetPassword(ctx context.Context, userID int64, password []byte) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1 WHERE user_id = $2;
	`, password, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff != 1 {
		return err
	}
	return nil
}

func (s *usersRepo) GetUserPasswordHash(ctx context.Context, userID int64) ([]byte, error) {
	var hash []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT password_hash FROM users WHERE user_id = $1;
	`, userID).Scan(&hash)
	return hash, err
}

func (s *usersRepo) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM users WHERE LOWER(username)=LOWER($1)`, username).Scan(&id)
	return id, err
}

func (s *usersRepo) IsUserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)`, userID).Scan(&exists)
	return exists, err
}

func (s *usersRepo) IsUserExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username)=LOWER($1))`, username).Scan(&exists)
	return exists, err
}

func (s *usersRepo) SetPartner(ctx context.Context, userID int64, partnerID int64) error {
	_, err := s.db.ExecContext(ctx, `
        UPDATE users SET partner_id = $1 WHERE user_id = $2;
    `, partnerID, userID)
	return err
}

func (s *usersRepo) GetUsername(ctx context.Context, userID int64) (string, error) {
	var username string
	err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE user_id=$1;`, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func (s *usersRepo) GetPartnerID(ctx context.Context, userID int64) (int64, error) {
	var id int64
	query := `SELECT partner_id FROM users WHERE user_id = $1;`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *usersRepo) SetPartners(ctx context.Context, userID, partnerID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// user -> partner
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1 WHERE user_id = $2
    `, partnerID, userID)
	if err != nil {
		return err
	}

	// partner -> user
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1 WHERE user_id = $2
    `, userID, partnerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *usersRepo) RemovePartners(ctx context.Context, userID, partnerID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// user -> partner
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1 WHERE user_id = $2
    `, 0, userID)
	if err != nil {
		return err
	}

	// partner -> user
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1 WHERE user_id = $2
    `, 0, partnerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *usersRepo) ClearAllPartnersHistory(ctx context.Context, userID, partnerID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Удаляем общие даты
	_, err = tx.ExecContext(ctx, `
		DELETE from important_dates WHERE (user_id = $1 AND partner_id = $2) OR (partner_id = $1 AND user_id = $2);
	`, userID, partnerID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM compliments AS c
		USING user_compliment AS uc
		WHERE c.id = uc.compliment_id
		  AND (uc.user_id = $1 OR uc.user_id = $2);
	`, userID, partnerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *usersRepo) DeleteUser(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM users WHERE user_id=$1
	`, userID)
	return err
}

func (s *usersRepo) UpdateUser(ctx context.Context, userID int64, username string, partnerID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET username = $1, partner_id = $2 WHERE user_id = $3
	`, username, partnerID, userID)
	return err
}
