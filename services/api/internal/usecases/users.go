package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) AddUser(ctx context.Context, userID int64, username string) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrUserExists
	}

	return uc.users.AddUser(ctx, userID, username)
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
			return nil, errs.ErrUserNotFound
		}
		return nil, err
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
