package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) GetMyUserConfig(ctx context.Context, userID int64) (*domain.UserConfigResponse, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	userConfig, err := uc.userConfig.GetUserConfig(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserConfigNotFound
		}
		return nil, err
	}

	return &domain.UserConfigResponse{
		DailyMessageTime:   userConfig.DailyMessageTime,
		MaxComplimentCount: userConfig.MaxComplimentCount,
		ComplimentCount:    userConfig.ComplimentCount,
	}, nil
}

func (uc *UseCase) GetPartnerUserConfig(ctx context.Context, userID int64) (*domain.UserConfigResponse, error) {
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
	if partnerID <= 0 {
		return nil, errs.ErrUserNotFound
	}

	userConfig, err := uc.userConfig.GetUserConfig(ctx, partnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserConfigNotFound
		}
		return nil, err
	}

	return &domain.UserConfigResponse{
		DailyMessageTime:   userConfig.DailyMessageTime,
		MaxComplimentCount: userConfig.MaxComplimentCount,
		ComplimentCount:    userConfig.ComplimentCount,
	}, nil
}

func (uc *UseCase) UpdateUserConfig(ctx context.Context, userID int64, userConfig *domain.UserConfigPatch) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	if userConfig.MaxComplimentCount != nil {
		err = uc.userConfig.SetComplimentMaxCount(ctx, userID, *userConfig.MaxComplimentCount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) ResetMyUserConfig(ctx context.Context, userID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	return uc.userConfig.SetDefault(ctx, userID)
}

func (uc *UseCase) ResetPartnerUserConfig(ctx context.Context, userID int64) error {
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

	return uc.userConfig.SetDefault(ctx, partnerID)
}
