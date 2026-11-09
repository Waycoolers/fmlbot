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

func (s *tokensRepo) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	idBytes := make([]byte, 16)
	if _, err := rand.Reader.Read(idBytes); err != nil {
		return "", err
	}
	tokenID := base64.URLEncoding.EncodeToString(idBytes)

	secretBytes := make([]byte, 32)
	_, err := rand.Reader.Read(secretBytes)
	if err != nil {
		return "", err
	}
	secret := base64.URLEncoding.EncodeToString(secretBytes)

	// Полный токен: "tokenID.secret"
	fullToken := tokenID + "." + secret

	// Хешируем секрет
	hash, err := bcrypt.GenerateFromPassword(secretBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Колонка без часового пояса, поэтому всегда пишем UTC
	expiresAt := time.Now().UTC().Add(ttl)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}

	//_, err = tx.ExecContext(ctx, `
	//    UPDATE refresh_tokens
	//    SET revoked = true
	//    WHERE user_id = $1 AND revoked = false
	//`, userID)
	//if err != nil {
	//	er := tx.Rollback()
	//	if er != nil {
	//		return "", er
	//	}
	//	return "", err
	//}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at)
        VALUES ($1, $2, $3, $4)
    `, tokenID, string(hash), userID, expiresAt)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return "", er
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return fullToken, nil
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

// Consume проверяет токен и гасит его. Один токен можно обменять только один раз:
// при параллельных запросах проходит ровно один.
func (s *tokensRepo) Consume(ctx context.Context, fullToken string) (int64, error) {
	userID, err := s.Validate(ctx, fullToken)
	if err != nil {
		return 0, err
	}

	res, err := s.db.ExecContext(ctx, `
        UPDATE refresh_tokens SET revoked = true
        WHERE token_id = $1 AND revoked = false
    `, splitToken(fullToken)[0])
	if err != nil {
		return 0, err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if aff != 1 {
		return 0, errs.ErrTokenInvalid
	}
	return userID, nil
}

func (s *tokensRepo) RevokeAll(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
        UPDATE refresh_tokens SET revoked = true
        WHERE user_id = $1 AND revoked = false
    `, userID)
	return err
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
