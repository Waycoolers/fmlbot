package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) AddCompliment(ctx context.Context, userID int64, text string) (*domain.Compliment, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	compliment, err := uc.compliments.AddCompliment(ctx, userID, text)
	return compliment, err
}

func (uc *UseCase) GetAllCompliments(ctx context.Context, userID int64) (*[]domain.Compliment, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	compliments, err := uc.compliments.GetCompliments(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &compliments, nil
}

func (uc *UseCase) RemoveCompliment(ctx context.Context, userID int64, complimentID int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	err = uc.compliments.DeleteCompliment(ctx, userID, complimentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrComplimentNotFound
		}
		return err
	}
	return nil
}

func (uc *UseCase) UpdateCompliment(ctx context.Context, userID int64, complimentID int64, request *domain.ComplimentRequest) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	err = uc.compliments.UpdateCompliment(ctx, userID, complimentID, request.Text, request.IsSent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrComplimentNotFound
		}
		return err
	}

	return nil
}

func (uc *UseCase) AcquireCompliment(ctx context.Context, userID int64) (*domain.Compliment, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	partnerID, err := uc.users.GetPartnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if partnerID == 0 {
		return nil, errs.ErrUserNotFound
	}

	text, err := uc.compliments.AcquireCompliment(ctx, partnerID)
	return &domain.Compliment{
		Text: text,
	}, err
}
