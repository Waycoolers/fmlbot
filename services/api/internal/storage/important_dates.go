package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type importantDatesRepo struct {
	db *sqlx.DB
}

func (s *importantDatesRepo) AddImportantDate(ctx context.Context, telegramID int64, partnerID sql.NullInt64, title string,
	date time.Time, notifyBefore int) (*domain.ImportantDate, error) {
	var importantDate domain.ImportantDate
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO important_dates(telegram_id, partner_id, title, date, notify_before_days)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, telegram_id, partner_id, title, date, is_active, last_notification_at, notify_before_days, created_at;
	`, telegramID, partnerID, title, date, notifyBefore).Scan(&importantDate.ID, &importantDate.TelegramID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays, &importantDate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &importantDate, nil
}

func (s *importantDatesRepo) GetImportantDates(ctx context.Context, telegramID int64) (importantDates []domain.ImportantDate, err error) {
	importantDates = make([]domain.ImportantDate, 0)
	err = s.db.SelectContext(ctx, &importantDates, `
		SELECT * FROM important_dates
		WHERE telegram_id = $1 OR partner_id = $1
		ORDER BY created_at;
	`, telegramID)
	if err != nil {
		return nil, err
	}
	return importantDates, nil
}

func (s *importantDatesRepo) GetImportantDateByID(ctx context.Context, id int64, userID int64) (*domain.ImportantDate, error) {
	var importantDate domain.ImportantDate
	err := s.db.QueryRowContext(ctx, `SELECT * FROM important_dates WHERE id=$1 AND (telegram_id = $2 OR partner_id = $2)`, id, userID).Scan(&importantDate.ID, &importantDate.TelegramID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays, &importantDate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &importantDate, err
}

func (s *importantDatesRepo) DeleteImportantDate(ctx context.Context, id int64, userID int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM important_dates WHERE id=$1 AND (telegram_id = $2 OR partner_id = $2);
	`, id, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *importantDatesRepo) EditImportantDate(ctx context.Context, id int64, userID int64, date domain.ImportantDateRequest) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE important_dates
		SET
			title = $1,
			date = $2,
			is_active = $3,
			notify_before_days = $4
		WHERE id = $5 AND (telegram_id = $6 OR partner_id = $6);
	`,
		date.Title,
		date.Date,
		date.IsActive,
		date.NotifyBeforeDays,
		id,
		userID,
	)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *importantDatesRepo) GetAllActiveImportantDates(ctx context.Context) (importantDates []domain.ImportantDate, err error) {
	importantDates = make([]domain.ImportantDate, 0)
	err = s.db.SelectContext(ctx, &importantDates, `
		SELECT * FROM important_dates
		WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, err
	}
	return importantDates, nil
}

func (s *importantDatesRepo) UpdateLastNotificationAt(ctx context.Context, id int64, timestamp time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE important_dates SET last_notification_at=$1 WHERE id=$2;
	`, timestamp, id)
	return err
}

func (s *importantDatesRepo) MakeImportantDatePrivate(ctx context.Context, dateID int64, userID int64) error {
	res, err := s.db.ExecContext(ctx, `
        UPDATE important_dates
        SET 
            telegram_id = CASE WHEN partner_id = $1 THEN $1 ELSE telegram_id END,
            partner_id = NULL
        WHERE id = $2 AND (telegram_id = $3 OR partner_id = $3)
    `, userID, dateID, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *importantDatesRepo) MakeImportantDateShared(ctx context.Context, dateID int64, userID int64, partnerID int64) error {
	res, err := s.db.ExecContext(ctx, `
        UPDATE important_dates
        SET partner_id = $1
        WHERE id = $2 AND telegram_id = $3 AND (partner_id IS NULL OR partner_id != $1)
    `, partnerID, dateID, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows // или errs.ErrCannotMakeShared
	}
	return nil
}
