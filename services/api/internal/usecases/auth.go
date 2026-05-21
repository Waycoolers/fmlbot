package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"golang.org/x/crypto/bcrypt"
)

func (uc *UseCase) VerifyUser(ctx context.Context, username string, password string) (int64, error) {
	userID, err := uc.users.GetUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrUserNotFound
		}
		return 0, err
	}

	hash, err := uc.users.GetUserPasswordHash(ctx, userID)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, errs.ErrWrongPassword
		}
		return 0, err
	}
	return userID, nil
}
