package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const uniqueViolationCode = "23505"

type usersRepo struct {
	db *sqlx.DB
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == uniqueViolationCode
}

func (s *usersRepo) AddUser(ctx context.Context, userID int64, username string, password []byte, telegramID sql.NullInt64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	res, err := tx.ExecContext(ctx, `
        INSERT INTO users (user_id, username, password_hash, telegram_id)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO NOTHING;
    `, userID, username, password, telegramID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			if pgErr.Constraint == "users_telegram_id_key" {
				return errs.ErrUserExists
			}
			return errs.ErrUsernameIsAlreadyTaken
		}
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
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
		return errs.ErrUserNotFound
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

func (s *usersRepo) GetUsername(ctx context.Context, userID int64) (string, error) {
	var username string
	err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE user_id=$1;`, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func (s *usersRepo) GetPartnerID(ctx context.Context, userID int64) (int64, error) {
	partnerID, _, err := s.GetPair(ctx, userID)
	return partnerID, err
}

func (s *usersRepo) GetPair(ctx context.Context, userID int64) (partnerID, version int64, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT partner_id, pair_version FROM users WHERE user_id = $1`, userID).Scan(&partnerID, &version)
	return
}

type pairState struct {
	UserID    int64 `db:"user_id"`
	PartnerID int64 `db:"partner_id"`
	Version   int64 `db:"pair_version"`
}

func checkCanPair(rows []pairState, userID int64) error {
	if len(rows) != 2 {
		return errs.ErrUserNotFound
	}
	for _, row := range rows {
		if row.PartnerID == 0 {
			continue
		}
		if row.UserID == userID {
			return errs.ErrAlreadyHasPartner
		}
		return errs.ErrPartnerAlreadyHasPartner
	}
	return nil
}

// Все изменения пары блокируют пользователей в одинаковом порядке.
func lockPairUsers(ctx context.Context, tx *sqlx.Tx, userID, partnerID int64) ([]pairState, error) {
	var rows []pairState
	err := tx.SelectContext(ctx, &rows, `
        SELECT user_id, partner_id, pair_version FROM users
        WHERE user_id IN ($1, $2)
        ORDER BY user_id
        FOR NO KEY UPDATE
    `, userID, partnerID)
	return rows, err
}

func linkPartners(ctx context.Context, tx *sqlx.Tx, userID, partnerID, requestID int64) error {
	rows, err := lockPairUsers(ctx, tx, userID, partnerID)
	if err != nil {
		return err
	}
	if err = checkCanPair(rows, userID); err != nil {
		return err
	}

	// После блокировки пользователей заявка ещё должна быть действующей:
	// параллельный отказ мог удалить её с момента первоначального чтения.
	res, err := tx.ExecContext(ctx, `DELETE FROM partner_requests WHERE id = $1 AND to_user_id = $2`, requestID, userID)
	if err != nil {
		return err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if deleted != 1 {
		return errs.ErrPartnerRequestNotFound
	}

	// user -> partner
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1, pair_version = $3 WHERE user_id = $2
    `, partnerID, userID, requestID)
	if err != nil {
		return err
	}

	// partner -> user
	_, err = tx.ExecContext(ctx, `
        UPDATE users SET partner_id = $1, pair_version = $3 WHERE user_id = $2
    `, userID, partnerID, requestID)
	if err != nil {
		return err
	}

	// У пары больше не может быть других заявок
	_, err = tx.ExecContext(ctx, `
        DELETE FROM partner_requests
        WHERE from_user_id IN ($1, $2) OR to_user_id IN ($1, $2)
    `, userID, partnerID)
	return err
}

// CreatePartnerRequest создаёт заявку на пару. Повторная заявка тому же человеку не дублируется.
func (s *usersRepo) CreatePartnerRequest(ctx context.Context, userID, partnerID int64) (int64, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rows, err := lockPairUsers(ctx, tx, userID, partnerID)
	if err != nil {
		return 0, err
	}
	if err = checkCanPair(rows, userID); err != nil {
		return 0, err
	}

	var id int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO partner_requests (from_user_id, to_user_id)
        VALUES ($1, $2)
        ON CONFLICT (from_user_id, to_user_id) DO UPDATE SET created_at = NOW()
        RETURNING id
    `, userID, partnerID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// AcceptPartnerRequest принимает заявку и возвращает id её автора.
func (s *usersRepo) AcceptPartnerRequest(ctx context.Context, requestID, userID int64) (int64, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	var fromUserID int64
	err = tx.GetContext(ctx, &fromUserID, `
        SELECT from_user_id FROM partner_requests WHERE id = $1 AND to_user_id = $2
    `, requestID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrPartnerRequestNotFound
		}
		return 0, err
	}

	if err = linkPartners(ctx, tx, userID, fromUserID, requestID); err != nil {
		return 0, err
	}
	return fromUserID, tx.Commit()
}

// DeclinePartnerRequest удаляет заявку и возвращает id её автора.
func (s *usersRepo) DeclinePartnerRequest(ctx context.Context, requestID, userID int64) (int64, error) {
	var fromUserID int64
	err := s.db.QueryRowContext(ctx, `
        DELETE FROM partner_requests WHERE id = $1 AND to_user_id = $2
        RETURNING from_user_id
    `, requestID, userID).Scan(&fromUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errs.ErrPartnerRequestNotFound
	}
	return fromUserID, err
}

func (s *usersRepo) GetPartnerRequests(ctx context.Context, userID int64) ([]domain.PartnerRequest, error) {
	requests := make([]domain.PartnerRequest, 0)
	err := s.db.SelectContext(ctx, &requests, `
        SELECT r.id, r.from_user_id, u.username AS from_username, r.created_at
        FROM partner_requests AS r
        JOIN users AS u ON u.user_id = r.from_user_id
        WHERE r.to_user_id = $1
        ORDER BY r.created_at
    `, userID)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// unlinkPartners удаляет общую историю пары, разрывает связь и сбрасывает настройки обоих.
func unlinkPartners(ctx context.Context, tx *sqlx.Tx, userID, partnerID int64) error {
	// Удаляем общие даты
	_, err := tx.ExecContext(ctx, `
		DELETE FROM important_dates
		WHERE (user_id = $1 AND partner_id = $2) OR (partner_id = $1 AND user_id = $2);
	`, userID, partnerID)
	if err != nil {
		return err
	}

	// Предложения и напоминания принадлежат только текущей паре.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM couple_plans
		WHERE (owner_id = $1 AND partner_id = $2) OR (owner_id = $2 AND partner_id = $1)
	`, userID, partnerID)
	if err != nil {
		return err
	}

	// Удаляем отправленные комплименты. Заготовленные остаются у автора
	_, err = tx.ExecContext(ctx, `
		DELETE FROM compliments AS c
		USING user_compliment AS uc
		WHERE c.id = uc.compliment_id
		  AND c.is_sent = true
		  AND uc.user_id IN ($1, $2);
	`, userID, partnerID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users SET partner_id = 0, pair_version = 0
		WHERE user_id = $1 OR (user_id = $2 AND partner_id = $1);
	`, userID, partnerID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE user_config
		SET compliment_count = 0, compliment_token_bucket = 2, last_bucket_update = $3
		WHERE user_id IN ($1, $2);
	`, userID, partnerID, time.Now().UTC())
	return err
}

func lockExpectedPair(ctx context.Context, tx *sqlx.Tx, userID, partnerID, version int64) error {
	rows, err := lockPairUsers(ctx, tx, userID, partnerID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.UserID == userID {
			if row.PartnerID != partnerID || row.Version != version {
				return errs.ErrPairChanged
			}
			return nil
		}
	}
	return errs.ErrUserNotFound
}

func (s *usersRepo) RemovePartners(ctx context.Context, userID, partnerID, version int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	if err = lockExpectedPair(ctx, tx, userID, partnerID, version); err != nil {
		return err
	}
	if err = unlinkPartners(ctx, tx, userID, partnerID); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteUser удаляет пользователя и все его данные одной транзакцией.
// partnerID равен 0, если партнёра нет.
func (s *usersRepo) DeleteUser(ctx context.Context, userID, partnerID, version int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	if err = lockExpectedPair(ctx, tx, userID, partnerID, version); err != nil {
		return err
	}
	if partnerID != 0 {
		if err = unlinkPartners(ctx, tx, userID, partnerID); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM compliments AS c
		USING user_compliment AS uc
		WHERE c.id = uc.compliment_id AND uc.user_id = $1;
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM important_dates WHERE user_id = $1 OR partner_id = $1;
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM users WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *usersRepo) UpdateUsername(ctx context.Context, userID int64, username string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET username = $1 WHERE user_id = $2
	`, username, userID)
	if isUniqueViolation(err) {
		return errs.ErrUsernameIsAlreadyTaken
	}
	return err
}
