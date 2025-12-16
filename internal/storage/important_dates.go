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
	err = s.DB.QueryRowContext(ctx, `
		INSERT INTO important_dates(telegram_id, partner_id, title, date, notify_before_days)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, telegram_id, partner_id, title, date, is_active, last_notification_at, notify_before_days;
	`, telegramID, partnerID, title, date, notifyBefore).Scan(&importantDate.ID, &importantDate.TelegramID,
		&importantDate.PartnerID, &importantDate.Title, &importantDate.Date, &importantDate.IsActive,
		&importantDate.LastNotificationAt, &importantDate.NotifyBeforeDays)
	if err != nil {
		return nil, err
	}
	return &importantDate, tx.Commit()
}
