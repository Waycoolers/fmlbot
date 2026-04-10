package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) AddImportantDate(ctx context.Context, userID int64, req domain.ImportantDateRequest) (*domain.ImportantDate, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	var partnerIDNull sql.NullInt64
	if req.IsShared {
		partner, err := uc.users.GetPartnerID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if partner == 0 {
			return nil, errs.ErrPartnerNotFound
		}
		partnerIDNull = sql.NullInt64{Int64: partner, Valid: true}
	} else {
		partnerIDNull = sql.NullInt64{Valid: false}
	}

	date, err := uc.importantDates.AddImportantDate(ctx, userID, partnerIDNull, req.Title, req.Date, req.NotifyBeforeDays)
	if err != nil {
		return nil, err
	}

	return date, nil
}

func (uc *UseCase) RemoveImportantDate(ctx context.Context, userID int64, id int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	err = uc.importantDates.DeleteImportantDate(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrImportantDateNotFound
		}
		return err
	}
	return nil
}

func (uc *UseCase) UpdateImportantDate(ctx context.Context, userID int64, id int64, date domain.ImportantDateRequest) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	err = uc.importantDates.EditImportantDate(ctx, id, userID, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrImportantDateNotFound
		}
		return err
	}
	return nil
}

func (uc *UseCase) GetImportantDate(ctx context.Context, userID int64, id int64) (*domain.ImportantDate, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	date, err := uc.importantDates.GetImportantDateByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrImportantDateNotFound
		}
		return nil, err
	}
	return date, nil
}

func (uc *UseCase) GetAllImportantDates(ctx context.Context, userID int64) ([]domain.ImportantDate, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	dates, err := uc.importantDates.GetImportantDates(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dates, nil
}

func (uc *UseCase) UpdateImportantDateSharing(ctx context.Context, dateID int64, userID int64, makeShared bool) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	if makeShared {
		partner, err := uc.users.GetPartnerID(ctx, userID)
		if err != nil {
			return err
		}
		if partner == 0 {
			return errs.ErrPartnerNotFound
		}
		return uc.importantDates.MakeImportantDateShared(ctx, dateID, userID, partner)
	} else {
		return uc.importantDates.MakeImportantDatePrivate(ctx, dateID, userID)
	}
}
