package storage

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/domain"
)

func (s *Storage) AddCompliment(ctx context.Context, telegramID int64, text string) (*domain.Compliment, error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
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
			log.Printf("Ошибка отката транзакции: %v", er)
		}
		log.Printf("Ошибка добавления комплимента: %v", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO user_compliment (telegram_id, compliment_id)
        VALUES ($1, $2)
    `, telegramID, compliment.ID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			log.Printf("Ошибка отката транзакции: %v", er)
		}
		log.Printf("Ошибка добавления user_compliment: %v", err)
		return nil, err
	}

	return &compliment, tx.Commit()
}

func (s *Storage) GetCompliments(ctx context.Context, telegramID int64) (compliments []domain.Compliment, err error) {
	compliments = []domain.Compliment{}
	err = s.DB.SelectContext(ctx, &compliments, `
		SELECT c.id, c.text, c.is_sent, c.created_at
		FROM compliments AS c
		JOIN user_compliment AS uc ON c.id = uc.compliment_id
		WHERE uc.telegram_id = $1
		ORDER BY c.created_at;
	`, telegramID)
	if err != nil {
		log.Printf("Ошибка получения списка комплиментов: %v", err)
		return nil, err
	}
	return compliments, nil
}

func (s *Storage) DeleteCompliment(ctx context.Context, telegramID int64, complimentID int64) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM user_compliment WHERE telegram_id=$1 AND compliment_id=$2
	`, telegramID, complimentID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM compliments WHERE id=$1
	`, complimentID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Storage) MarkComplimentSent(ctx context.Context, complimentID int64) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE compliments SET is_sent=true WHERE id=$1;
	`, complimentID)
	return err
}
