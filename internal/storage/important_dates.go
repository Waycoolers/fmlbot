package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
)

func (s *Storage) AddImportantDate(ctx context.Context, telegramID sql.NullInt64, partnerID sql.NullInt64, title string,
	date time.Time, notifyBefore int) (*domain.ImportantDate, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var importantDate domain.ImportantDate
	err = tx.QueryRowContext(ctx, `
		INSERT INTO important_dates(telegram_id, partner_id, title, date, notify_before_days)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, telegram_id, partner_id, title, date, is_active, last_notification_at, notify_before_days, created_at;
	`, telegramID, partnerID, title, date, notifyBefore).Scan(&importantDate.ID, &importantDate.TelegramID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays, &importantDate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &importantDate, tx.Commit()
}

func (s *Storage) GetImportantDates(ctx context.Context, telegramID sql.NullInt64) (importantDates []domain.ImportantDate, err error) {
	importantDates = make([]domain.ImportantDate, 0)
	err = s.DB.SelectContext(ctx, &importantDates, `
		SELECT * FROM important_dates
		WHERE telegram_id = $1 OR partner_id = $1
		ORDER BY created_at;
	`, telegramID)
	if err != nil {
		return nil, err
	}
	return importantDates, nil
}

func (s *Storage) GetImportantDateByID(ctx context.Context, id int64) (importantDate domain.ImportantDate, err error) {
	err = s.DB.QueryRowContext(ctx, `SELECT * FROM important_dates WHERE id=$1`, id).Scan(&importantDate.ID, &importantDate.TelegramID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays, &importantDate.CreatedAt)
	if err != nil {
		return domain.ImportantDate{}, err
	}
	return importantDate, err
}

func (s *Storage) DeleteImportantDate(ctx context.Context, id int64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM important_dates WHERE id=$1`, id)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	return tx.Commit()
}

func (s *Storage) EditImportantDate(ctx context.Context, date domain.ImportantDate) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE important_dates
		SET
			title = $1,
			date = $2,
			is_active = $3,
			notify_before_days = $4,
			last_notification_at = $5,
			telegram_id = $6,
			partner_id = $7
		WHERE id = $8
	`,
		date.Title,
		date.Date,
		date.IsActive,
		date.NotifyBeforeDays,
		date.LastNotificationAt,
		date.TelegramID,
		date.PartnerID,
		date.ID,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
