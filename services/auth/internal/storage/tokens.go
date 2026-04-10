package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
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

	expiresAt := time.Now().Add(ttl)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE refresh_tokens 
        SET revoked = true 
        WHERE user_id = $1 AND revoked = false
    `, userID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return "", er
		}
		return "", err
	}

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
		return 0, errors.New("invalid token format")
	}
	tokenID, secret := parts[0], parts[1]

	var userID int64
	var tokenHash string
	var expiresAt time.Time
	var revoked bool

	err := s.db.QueryRowContext(ctx, `
        SELECT user_id, token_hash, expires_at, revoked
        FROM refresh_tokens
        WHERE token_id = $1
    `, tokenID).Scan(&userID, &tokenHash, &expiresAt, &revoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrTokenNotFound
		}
		return 0, err
	}

	if revoked {
		return 0, errors.New("token revoked")
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("token expired")
	}

	// Сравниваем хеш с секретом
	err = bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(secret))
	if err != nil {
		return 0, errors.New("invalid token")
	}

	return userID, nil
}

func (s *tokensRepo) Revoke(ctx context.Context, fullToken string) error {
	parts := splitToken(fullToken)
	if parts == nil {
		return errors.New("invalid token format")
	}
	tokenID := parts[0]

	_, err := s.db.ExecContext(ctx, `
        UPDATE refresh_tokens SET revoked = true
        WHERE token_id = $1
    `, tokenID)
	return err
}

func splitToken(fullToken string) []string {
	parts := strings.SplitN(fullToken, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	return parts
}
