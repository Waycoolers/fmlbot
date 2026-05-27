package usecases

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/"

	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	charsetLen := len(charset)
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = charset[randomBytes[i]%byte(charsetLen)]
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

	return uc.users.AddUser(ctx, userID, username, hashedPassword)
}

func (uc *UseCase) AddUserWithRandomPassword(ctx context.Context, userID int64, username string) (string, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errs.ErrUserExists
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

	return password, uc.users.AddUser(ctx, userID, username, hashedPassword)
}

func (uc *UseCase) RemoveUser(ctx context.Context, userID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	partnerID, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		return err
	}
	if partnerID != 0 {
		err = uc.users.ClearAllPartnersHistory(ctx, userID, partnerID)
		if err != nil {
			return err
		}

		err = uc.users.RemovePartners(ctx, userID, partnerID)
		if err != nil {
			return err
		}

		err = uc.userConfig.SetDefault(ctx, partnerID)
		if err != nil {
			return err
		}
	}

	err = uc.users.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (uc *UseCase) UpdateUser(ctx context.Context, userID int64, username string, partnerID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	return uc.users.UpdateUser(ctx, userID, username, partnerID)
}

func (uc *UseCase) UpdatePartner(ctx context.Context, userID int64, username string, partnerID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	id, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrPartnerNotFound
		}
		return err
	}

	return uc.users.UpdateUser(ctx, id, username, partnerID)
}

func (uc *UseCase) AddPartners(ctx context.Context, userID int64, partnerID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	if userID == partnerID {
		return errs.ErrCannotPartnerYourself
	}

	userPartner, _ := uc.users.GetPartnerID(ctx, userID)
	if userPartner != 0 {
		return errs.ErrAlreadyHasPartner
	}

	partnerPartner, _ := uc.users.GetPartnerID(ctx, partnerID)
	if partnerPartner != 0 {
		return errs.ErrPartnerAlreadyHasPartner
	}

	return uc.users.SetPartners(ctx, userID, partnerID)
}

func (uc *UseCase) GetMe(ctx context.Context, userID int64) (*domain.UserResponse, error) {
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
		ID:        userID,
		PartnerID: partnerID,
		Username:  username,
	}, nil
}

func (uc *UseCase) ChangePassword(ctx context.Context, userID int64, password string) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	err = uc.users.SetPassword(ctx, userID, hashedPassword)
	if err != nil {
		return err
	}
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

func (uc *UseCase) RemovePartners(ctx context.Context, userID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	partnerID, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		return err
	}
	if partnerID == 0 {
		return errs.ErrPartnerNotFound
	}

	err = uc.users.ClearAllPartnersHistory(ctx, userID, partnerID)
	if err != nil {
		return err
	}

	err = uc.users.RemovePartners(ctx, userID, partnerID)
	if err != nil {
		return err
	}

	err = uc.userConfig.SetDefault(ctx, userID)
	if err != nil {
		return err
	}

	err = uc.userConfig.SetDefault(ctx, partnerID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) GetUserByUsername(ctx context.Context, username string) (*domain.UserResponse, error) {
	userID, err := uc.users.GetUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	me, err := uc.GetMe(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	res := &domain.UserResponse{
		ID:        me.ID,
		Username:  me.Username,
		PartnerID: me.PartnerID,
	}
	return res, nil
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
