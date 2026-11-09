package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

const telegramLinkTTL = 10 * time.Minute

func (uc *UseCase) CreateTelegramLink(ctx context.Context, userID int64) (string, error) {
	var bytes [24]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	hash := sha256.Sum256(bytes[:])
	if err := uc.users.SaveTelegramLink(ctx, userID, hash[:], time.Now().Add(telegramLinkTTL)); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func (uc *UseCase) LinkTelegram(ctx context.Context, telegramID int64, code string) error {
	bytes, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil || len(bytes) != 24 {
		return errs.ErrTelegramLinkInvalid
	}
	hash := sha256.Sum256(bytes)
	return uc.users.UseTelegramLink(ctx, telegramID, hash[:])
}

func (uc *UseCase) ResolveTelegramUser(ctx context.Context, telegramID int64) (int64, error) {
	id, err := uc.users.GetTelegramUserID(ctx, telegramID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errs.ErrUserNotFound
	}
	return id, err
}

func (uc *UseCase) ResolveTelegramChat(ctx context.Context, userID int64) (int64, error) {
	id, err := uc.users.GetTelegramID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errs.ErrUserNotFound
	}
	return id, err
}
