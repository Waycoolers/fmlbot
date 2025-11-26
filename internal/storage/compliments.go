package storage

import (
	"context"
	"log"
)

func (s *Storage) AddCompliment(ctx context.Context, telegramID int64, text string) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		return err
	}

	var complimentID int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO compliments (text)
        VALUES ($1)
        RETURNING id
    `, text).Scan(&complimentID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			log.Printf("Ошибка отката транзакции: %v", er)
		}
		log.Printf("Ошибка добавления комплимента: %v", err)
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO user_compliment (telegram_id, compliment_id)
        VALUES ($1, $2)
    `, telegramID, complimentID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			log.Printf("Ошибка отката транзакции: %v", er)
		}
		log.Printf("Ошибка добавления user_compliment: %v", err)
		return err
	}

	return tx.Commit()
}
