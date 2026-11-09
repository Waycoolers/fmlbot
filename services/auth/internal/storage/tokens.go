package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type tokensRepo struct {
	db *sqlx.DB
}

type newToken struct {
	id   string
	hash string
	full string
}

func generateToken() (newToken, error) {
	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return newToken{}, err
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return newToken{}, err
	}
	hash, err := bcrypt.GenerateFromPassword(secret[:], bcrypt.DefaultCost)
	if err != nil {
		return newToken{}, err
	}

	id := base64.URLEncoding.EncodeToString(idBytes[:])
	return newToken{id: id, hash: string(hash), full: id + "." + base64.URLEncoding.EncodeToString(secret[:])}, nil
}

func insertToken(ctx context.Context, db sqlx.ExecerContext, userID int64, token newToken, ttl time.Duration) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, token.id, token.hash, userID, time.Now().UTC().Add(ttl))
	return err
}

func (s *tokensRepo) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	if err = insertToken(ctx, s.db, userID, token, ttl); err != nil {
		return "", err
	}
	return token.full, nil
}

func (s *tokensRepo) Validate(ctx context.Context, fullToken string) (int64, error) {
	parts := splitToken(fullToken)
	if parts == nil {
		return 0, errs.ErrTokenInvalid
	}
	tokenID, secretBase64 := parts[0], parts[1] // secretBase64 — это то, что мы вытащили из строки

	// 1. Декодируем Base64 обратно в байты!
	secretBytes, err := base64.URLEncoding.DecodeString(secretBase64)
	if err != nil {
		return 0, errs.ErrTokenInvalid
	}

	var userID int64
	var tokenHash string
	var expiresAt time.Time
	var revoked bool

	err = s.db.QueryRowContext(ctx, `
        SELECT user_id, token_hash, expires_at, revoked
        FROM refresh_tokens
        WHERE token_id = $1
    `, tokenID).Scan(&userID, &tokenHash, &expiresAt, &revoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrTokenNotFound
		}
		return 0, err
	}

	if revoked {
		return 0, errs.ErrTokenInvalid
	}
	if time.Now().After(expiresAt) {
		return 0, errs.ErrTokenInvalid
	}

	// 2. Сравниваем хеш с сырыми байтами (которые мы получили после декодирования)
	err = bcrypt.CompareHashAndPassword([]byte(tokenHash), secretBytes)
	if err != nil {
		return 0, errs.ErrTokenInvalid
	}

	return userID, nil
}

func (s *tokensRepo) Revoke(ctx context.Context, fullToken string) error {
	parts := splitToken(fullToken)
	if parts == nil {
		return errs.ErrTokenInvalid
	}
	tokenID := parts[0]

	_, err := s.db.ExecContext(ctx, `
        UPDATE refresh_tokens SET revoked = true
        WHERE token_id = $1
    `, tokenID)
	return err
}

// Refresh rotation and password resets use the same per-user transaction lock.
func lockTokenUser(ctx context.Context, tx *sqlx.Tx, userID int64) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID)
	return err
}

func (s *tokensRepo) Rotate(ctx context.Context, fullToken string, ttl time.Duration) (int64, string, error) {
	userID, err := s.Validate(ctx, fullToken)
	if err != nil {
		return 0, "", err
	}
	token, err := generateToken()
	if err != nil {
		return 0, "", err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()
	if err = lockTokenUser(ctx, tx, userID); err != nil {
		return 0, "", err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked = true
		WHERE token_id = $1 AND revoked = false AND expires_at > $2
	`, splitToken(fullToken)[0], time.Now().UTC())
	if err != nil {
		return 0, "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, "", err
	}
	if affected != 1 {
		return 0, "", errs.ErrTokenInvalid
	}
	if err = insertToken(ctx, tx, userID, token, ttl); err != nil {
		return 0, "", err
	}
	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return userID, token.full, nil
}

func (s *tokensRepo) RevokeAll(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockTokenUser(ctx, tx, userID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *tokensRepo) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
        DELETE FROM refresh_tokens WHERE expires_at < $1
    `, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func splitToken(fullToken string) []string {
	parts := strings.SplitN(fullToken, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	return parts
}
