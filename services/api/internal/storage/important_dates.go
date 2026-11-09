package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

// API responses include delivery records while the scheduler still reads the
// legacy timestamp to avoid resending notices on the migration day.
const importantDateWithDeliveryTime = `
	SELECT d.id, d.user_id, d.partner_id, d.title, d.date, d.is_active,
	       GREATEST(d.last_notification_at, (
	           SELECT MAX(delivered_at) FROM important_date_deliveries WHERE date_id = d.id
	       )) AS last_notification_at,
	       d.notify_before_days, d.created_at
	FROM important_dates d
`

type importantDatesRepo struct {
	db *sqlx.DB
}

func (s *importantDatesRepo) AddImportantDate(ctx context.Context, userID int64, partnerID sql.NullInt64, pairVersion int64, title string,
	date time.Time, notifyBefore int) (*domain.ImportantDate, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if partnerID.Valid {
		if err = lockExpectedPair(ctx, tx, userID, partnerID.Int64, pairVersion); err != nil {
			return nil, err
		}
	}
	var importantDate domain.ImportantDate
	err = tx.QueryRowContext(ctx, `
		INSERT INTO important_dates(user_id, partner_id, title, date, notify_before_days)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, partner_id, title, date, is_active, last_notification_at, notify_before_days, created_at;
	`, userID, partnerID, title, date, notifyBefore).Scan(&importantDate.ID, &importantDate.UserID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays, &importantDate.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &importantDate, tx.Commit()
}

func (s *importantDatesRepo) GetImportantDates(ctx context.Context, userID int64) (importantDates []domain.ImportantDate, err error) {
	importantDates = make([]domain.ImportantDate, 0)
	err = s.db.SelectContext(ctx, &importantDates, importantDateWithDeliveryTime+`
		WHERE d.user_id = $1 OR d.partner_id = $1
		ORDER BY d.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	return importantDates, nil
}

func (s *importantDatesRepo) GetImportantDateByID(ctx context.Context, id int64, userID int64) (*domain.ImportantDate, error) {
	var importantDate domain.ImportantDate
	err := s.db.GetContext(ctx, &importantDate, importantDateWithDeliveryTime+`
		WHERE d.id = $1 AND (d.user_id = $2 OR d.partner_id = $2)
	`, id, userID)
	if err != nil {
		return nil, err
	}
	return &importantDate, err
}

func (s *importantDatesRepo) DeleteImportantDate(ctx context.Context, id int64, userID int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM important_dates WHERE id=$1 AND user_id = $2;
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
		WHERE id = $5 AND (user_id = $6 OR partner_id = $6);
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

func (s *importantDatesRepo) MakeImportantDatePrivate(ctx context.Context, dateID int64, userID int64) error {
	res, err := s.db.ExecContext(ctx, `
        UPDATE important_dates
        SET user_id = COALESCE(user_id, $2), partner_id = NULL
        WHERE id = $1 AND (user_id = $2 OR partner_id = $2)
    `, dateID, userID)
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

func (s *importantDatesRepo) MakeImportantDateShared(ctx context.Context, dateID, userID, partnerID, pairVersion int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockExpectedPair(ctx, tx, userID, partnerID, pairVersion); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
        UPDATE important_dates
        SET partner_id = $1
        WHERE id = $2 AND user_id = $3 AND (partner_id IS NULL OR partner_id != $1)
    `, partnerID, dateID, userID)
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
	return tx.Commit()
}
