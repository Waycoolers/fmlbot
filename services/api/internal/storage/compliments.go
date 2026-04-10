package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type complimentsRepo struct {
	db *sqlx.DB
}

func (s *complimentsRepo) AddCompliment(ctx context.Context, telegramID int64, text string) (*domain.Compliment, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var compliment domain.Compliment
	err = tx.QueryRowContext(ctx, `
        INSERT INTO compliments (text)
        VALUES ($1)
        RETURNING id, text, is_sent, created_at
    `, text).Scan(&compliment.ID, &compliment.Text, &compliment.IsSent, &compliment.CreatedAt)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return nil, er
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO user_compliment (telegram_id, compliment_id)
        VALUES ($1, $2)
    `, telegramID, compliment.ID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return nil, er
		}
		return nil, err
	}

	return &compliment, tx.Commit()
}

func (s *complimentsRepo) GetCompliments(ctx context.Context, telegramID int64) (compliments []domain.Compliment, err error) {
	compliments = []domain.Compliment{}
	err = s.db.SelectContext(ctx, &compliments, `
		SELECT c.id, c.text, c.is_sent, c.created_at
		FROM compliments AS c
		JOIN user_compliment AS uc ON c.id = uc.compliment_id
		WHERE uc.telegram_id = $1
		ORDER BY c.created_at;
	`, telegramID)
	if err != nil {
		return nil, err
	}
	return compliments, nil
}

func (s *complimentsRepo) DeleteCompliment(ctx context.Context, telegramID int64, complimentID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM user_compliment WHERE telegram_id=$1 AND compliment_id=$2
	`, telegramID, complimentID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if aff == 0 {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	res, err = tx.ExecContext(ctx, `
		DELETE FROM compliments WHERE id=$1
	`, complimentID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	aff, err = res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if aff == 0 {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (s *complimentsRepo) MarkComplimentSent(ctx context.Context, complimentID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE compliments SET is_sent=true WHERE id=$1;
	`, complimentID)
	return err
}

func (s *complimentsRepo) AcquireCompliment(ctx context.Context, partnerID int64) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	var bucket int
	var lastBucketUpdate time.Time
	var complimentCount int
	var maxComplimentCount int

	err = tx.QueryRowContext(ctx, `
        SELECT compliment_token_bucket, last_bucket_update, compliment_count, max_compliment_count
        FROM user_config
        WHERE telegram_id = $1
        FOR UPDATE
    `, partnerID).Scan(&bucket, &lastBucketUpdate, &complimentCount, &maxComplimentCount)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	now := time.Now().UTC()

	const maxBucket = 2
	refillInterval := time.Hour

	// --- 1. REFILL ---
	elapsed := now.Sub(lastBucketUpdate)
	hoursPassed := int(elapsed / refillInterval)

	tokensToAdd := hoursPassed
	maxAdd := maxBucket - bucket
	if tokensToAdd > maxAdd {
		tokensToAdd = maxAdd
	}

	currentBucket := bucket + tokensToAdd
	newLastBucketUpdate := lastBucketUpdate.Add(time.Duration(tokensToAdd) * refillInterval)

	// --- 2. ЕСЛИ ПУСТО ---
	if currentBucket == 0 {
		nextRefill := lastBucketUpdate.Add(refillInterval)
		minutes := int(nextRefill.Sub(now).Minutes())
		if minutes < 0 {
			minutes = 0
		}
		_ = tx.Rollback()
		return "", &errs.ErrBucketEmpty{Minutes: minutes}
	}

	// --- 3. ЛИМИТ ---
	if maxComplimentCount != -1 && complimentCount >= maxComplimentCount {
		_ = tx.Rollback()
		return "", errs.ErrLimitExceeded
	}

	// --- 4. КРИТИЧЕСКАЯ ЛОГИКА ---
	// если bucket был полный — стартуем таймер заново
	if currentBucket == maxBucket {
		newLastBucketUpdate = now
	}

	// --- 5. СПИСЫВАЕМ ТОКЕН ---
	newBucket := currentBucket - 1
	newComplimentCount := complimentCount + 1

	// --- 6. БЕРЁМ КОМПЛИМЕНТ ---
	var complimentText string
	var complimentID int64

	err = tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT c.id, c.text
			FROM user_compliment uc
			JOIN compliments c ON c.id = uc.compliment_id
			WHERE uc.telegram_id = $1 AND c.is_sent = false
			ORDER BY c.created_at
			LIMIT 1
			FOR UPDATE OF c SKIP LOCKED
		)
		UPDATE compliments c
		SET is_sent = true
		FROM candidate
		WHERE c.id = candidate.id
		RETURNING c.id, c.text
	`, partnerID).Scan(&complimentID, &complimentText)

	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return "", errs.ErrNoCompliments
		}
		return "", err
	}

	// --- 7. СОХРАНЯЕМ ---
	_, err = tx.ExecContext(ctx, `
        UPDATE user_config
        SET compliment_token_bucket = $1,
            compliment_count = $2,
            last_compliment_at = $3,
            last_bucket_update = $4
        WHERE telegram_id = $5
    `, newBucket, newComplimentCount, now, newLastBucketUpdate, partnerID)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return complimentText, nil
}

func (s *complimentsRepo) UpdateCompliment(ctx context.Context, userID int64, complimentID int64, text string, isSent bool) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE compliments AS c
		SET text = $1, is_sent = $2
		FROM user_compliment AS uc
		WHERE c.id = uc.compliment_id
		AND uc.telegram_id = $3
		AND uc.compliment_id = $4;
	`, text, isSent, userID, complimentID)
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
