package usecases

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

const (
	maxImportantDateTitle = 256
	maxNotifyBeforeDays   = 366
)

func validateImportantDate(req domain.ImportantDateRequest) error {
	if strings.TrimSpace(req.Title) == "" || utf8.RuneCountInString(req.Title) > maxImportantDateTitle {
		return errs.ErrInvalidImportantDate
	}
	if req.Date.IsZero() || req.NotifyBeforeDays < 0 || req.NotifyBeforeDays > maxNotifyBeforeDays {
		return errs.ErrInvalidImportantDate
	}
	return nil
}

func (uc *UseCase) AddImportantDate(ctx context.Context, userID int64, req domain.ImportantDateRequest) (*domain.ImportantDate, error) {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	if err = validateImportantDate(req); err != nil {
		return nil, err
	}
	// Выбрать можно только уже случившееся событие. Сутки запаса на часовые пояса клиентов
	if req.Date.After(time.Now().Add(24 * time.Hour)) {
		return nil, errs.ErrInvalidImportantDate
	}

	var partnerIDNull sql.NullInt64
	var pairVersion int64
	if req.IsShared {
		partner, version, err := uc.users.GetPair(ctx, userID)
		if err != nil {
			return nil, err
		}
		if req.ExpectedPairVersion != nil && (partner == 0 || *req.ExpectedPairVersion != version) {
			return nil, errs.ErrPairChanged
		}
		if partner == 0 {
			return nil, errs.ErrPartnerNotFound
		}
		pairVersion = version
		partnerIDNull = sql.NullInt64{Int64: partner, Valid: true}
	} else {
		partnerIDNull = sql.NullInt64{Valid: false}
	}

	date, err := uc.importantDates.AddImportantDate(ctx, userID, partnerIDNull, pairVersion, req.Title, req.Date, req.NotifyBeforeDays)
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

	if err = validateImportantDate(date); err != nil {
		return err
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

func (uc *UseCase) UpdateImportantDateSharing(ctx context.Context, dateID int64, userID int64, makeShared bool, expectedPairVersion *int64) error {
	exists, err := uc.users.IsUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}

	if makeShared {
		partner, version, err := uc.users.GetPair(ctx, userID)
		if err != nil {
			return err
		}
		if expectedPairVersion != nil && (partner == 0 || *expectedPairVersion != version) {
			return errs.ErrPairChanged
		}
		if partner == 0 {
			return errs.ErrPartnerNotFound
		}
		err = uc.importantDates.MakeImportantDateShared(ctx, dateID, userID, partner, version)
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrImportantDateNotFound
		}
		return err
	} else {
		err = uc.importantDates.MakeImportantDatePrivate(ctx, dateID, userID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrImportantDateNotFound
	}
	return err
}
