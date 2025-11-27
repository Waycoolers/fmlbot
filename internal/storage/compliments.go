package storage

import (
	"context"
	"database/sql"
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

func (s *Storage) GetCompliments(ctx context.Context, telegramID int64) (compliments []string, isSentList []bool, err error) {
	compliments = make([]string, 0)
	isSentList = make([]bool, 0)
	rows, err := s.DB.QueryContext(ctx, `
		SELECT c.text, c.is_sent FROM compliments AS c 
		JOIN user_compliment AS uc ON c.id = uc.compliment_id
		WHERE uc.telegram_id = $1 ORDER BY c.created_at;
	`, telegramID)
	if err != nil {
		return nil, nil, err
	}
	defer func(rows *sql.Rows) {
		er := rows.Close()
		if er != nil {
			log.Printf("Ошибка при закрытии соединения: %v", er)
		}
	}(rows)

	for rows.Next() {
		var compliment string
		var isSent bool
		if er := rows.Scan(&compliment, &isSent); er != nil {
			return nil, nil, er
		}
		compliments = append(compliments, compliment)
		isSentList = append(isSentList, isSent)
	}

	return compliments, isSentList, nil
}

func (s *Storage) GetActiveCompliments(ctx context.Context, telegramID int64) (compliments []string, err error) {
	compliments = make([]string, 0)
	rows, err := s.DB.QueryContext(ctx, `
		SELECT c.text FROM compliments AS c 
		JOIN user_compliment AS uc ON c.id = uc.compliment_id
		WHERE uc.telegram_id = $1 
		AND c.is_sent = false
		ORDER BY c.created_at;
	`, telegramID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		er := rows.Close()
		if er != nil {
			log.Printf("Ошибка при закрытии соединения: %v", er)
		}
	}(rows)

	for rows.Next() {
		var compliment string
		if er := rows.Scan(&compliment); er != nil {
			return nil, er
		}
		compliments = append(compliments, compliment)
	}

	return compliments, nil
}

func (s *Storage) getComplimentID(ctx context.Context, telegramID int64, complimentIndex int) (int64, error) {
	var complimentID int64
	err := s.DB.QueryRowContext(ctx, `
        SELECT c.id 
        FROM compliments AS c 
        JOIN user_compliment AS uc ON c.id = uc.compliment_id 
        WHERE uc.telegram_id = $1 
        AND c.is_sent = false
        ORDER BY c.created_at 
        LIMIT 1 OFFSET $2
    `, telegramID, complimentIndex).Scan(&complimentID)
	if err != nil {
		return 0, err
	}

	return complimentID, nil
}

func (s *Storage) DeleteCompliment(ctx context.Context, telegramID int64, complimentIndex int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	complimentID, err := s.getComplimentID(ctx, telegramID, complimentIndex)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM user_compliment 
        WHERE telegram_id = $1 AND compliment_id = $2
    `, telegramID, complimentID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM compliments 
        WHERE id = $1
    `, complimentID)
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return er
		}
		return err
	}

	return tx.Commit()
}
