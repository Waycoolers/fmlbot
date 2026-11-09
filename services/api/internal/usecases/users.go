package usecases

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log/slog"
	"math/big"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/"

	charsetLen := big.NewInt(int64(len(charset)))
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = charset[index.Int64()]
	}
	return string(password), nil
}

func (uc *UseCase) RegisterUser(ctx context.Context, userID int64, username string, password string) error {
	_, err := uc.users.GetUserIDByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if err == nil {
		return errs.ErrUsernameIsAlreadyTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return uc.users.AddUser(ctx, userID, username, hashedPassword, sql.NullInt64{})
}

func (uc *UseCase) AddUserWithRandomPassword(ctx context.Context, userID int64, username string) (string, error) {
	_, err := uc.users.GetTelegramUserID(ctx, userID)
	if err == nil {
		return "", errs.ErrUserExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	taken, err := uc.users.IsUserExistsByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if taken {
		return "", errs.ErrUsernameIsAlreadyTaken
	}

	password, err := generateRandomPassword(10)
	if err != nil {
		return "", err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return password, uc.users.AddUser(ctx, userID, username, hashedPassword, sql.NullInt64{Int64: userID, Valid: true})
}

func (uc *UseCase) RemoveUser(ctx context.Context, userID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	partnerID, version, err := uc.users.GetPair(ctx, userID)
	if err != nil {
		return err
	}

	if err = uc.users.DeleteUser(ctx, userID, partnerID, version); err != nil {
		return err
	}
	uc.revokeTokens(ctx, userID)
	return nil
}

// revokeTokens завершает сессии пользователя. Сбой здесь не отменяет уже выполненную
// операцию: refresh-токены всё равно истекут, поэтому ошибка только логируется.
func (uc *UseCase) revokeTokens(ctx context.Context, userID int64) {
	if err := uc.auth.RevokeAllTokens(ctx, userID); err != nil {
		slog.Error("Error revoking user tokens", "error", err)
	}
}

// UpdateUser меняет только имя пользователя. Партнёр меняется исключительно через создание и разрыв пары.
func (uc *UseCase) UpdateUser(ctx context.Context, userID int64, username string) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	ownerID, err := uc.users.GetUserIDByUsername(ctx, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil && ownerID != userID {
		return errs.ErrUsernameIsAlreadyTaken
	}

	return uc.users.UpdateUsername(ctx, userID, username)
}

// RequestPartner создаёт заявку: пара появляется только после согласия второго человека.
func (uc *UseCase) RequestPartner(ctx context.Context, userID int64, partnerID int64) (int64, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errs.ErrUserNotFound
	}

	if userID == partnerID {
		return 0, errs.ErrCannotPartnerYourself
	}

	return uc.users.CreatePartnerRequest(ctx, userID, partnerID)
}

func (uc *UseCase) AcceptPartnerRequest(ctx context.Context, requestID int64, userID int64) (int64, error) {
	return uc.users.AcceptPartnerRequest(ctx, requestID, userID)
}

func (uc *UseCase) DeclinePartnerRequest(ctx context.Context, requestID int64, userID int64) (int64, error) {
	return uc.users.DeclinePartnerRequest(ctx, requestID, userID)
}

func (uc *UseCase) GetPartnerRequests(ctx context.Context, userID int64) ([]domain.PartnerRequest, error) {
	return uc.users.GetPartnerRequests(ctx, userID)
}

func (uc *UseCase) GetMe(ctx context.Context, userID int64) (*domain.UserResponse, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	partnerID, version, err := uc.users.GetPair(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	username, err := uc.users.GetUsername(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	return &domain.UserResponse{
		ID:          userID,
		PartnerID:   partnerID,
		PairVersion: version,
		Username:    username,
	}, nil
}

func (uc *UseCase) ChangePassword(ctx context.Context, userID int64, oldPassword string, password string) error {
	currentHash, err := uc.users.GetUserPasswordHash(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrUserNotFound
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword(currentHash, []byte(oldPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrHashTooShort) {
			return errs.ErrWrongPassword
		}
		return err
	}

	return uc.setPassword(ctx, userID, password)
}

// ResetPassword выдаёт новый пароль без проверки старого. Доступен только сервису бота:
// там личность уже подтверждена аккаунтом Telegram.
func (uc *UseCase) ResetPassword(ctx context.Context, userID int64) (string, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errs.ErrUserNotFound
	}

	password, err := generateRandomPassword(10)
	if err != nil {
		return "", err
	}
	if err = uc.setPassword(ctx, userID, password); err != nil {
		return "", err
	}
	return password, nil
}

func (uc *UseCase) setPassword(ctx context.Context, userID int64, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err = uc.users.SetPassword(ctx, userID, hashedPassword); err != nil {
		return err
	}
	uc.revokeTokens(ctx, userID)
	return nil
}

func (uc *UseCase) GetPartner(ctx context.Context, userID int64) (*domain.UserResponse, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	partnerID, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrPartnerNotFound
		}
		return nil, err
	}
	if partnerID == 0 {
		return nil, errs.ErrPartnerNotFound
	}

	username, err := uc.users.GetUsername(ctx, partnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	return &domain.UserResponse{
		ID:        partnerID,
		PartnerID: userID,
		Username:  username,
	}, nil
}

func (uc *UseCase) RemovePartners(ctx context.Context, userID int64, expectedVersion *int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	partnerID, version, err := uc.users.GetPair(ctx, userID)
	if err != nil {
		return err
	}
	if expectedVersion != nil && *expectedVersion != version {
		return errs.ErrPairChanged
	}
	if partnerID == 0 {
		return errs.ErrPartnerNotFound
	}

	return uc.users.RemovePartners(ctx, userID, partnerID, version)
}

func (uc *UseCase) GetUserByUsername(ctx context.Context, username string) (*domain.UserResponse, error) {
	userID, err := uc.users.GetUserIDByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	name, err := uc.users.GetUsername(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.UserResponse{ID: userID, Username: name}, nil
}

func (uc *UseCase) ProcessFCMToken(ctx context.Context, userID int64, token string) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	err = uc.fcm.SetFCMToken(ctx, userID, token)
	if err != nil {
		return err
	}
	return nil
}
