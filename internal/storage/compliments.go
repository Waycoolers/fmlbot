package storage

import (
	"context"
	"log"
)

func (s *Storage) CanSendCompliment(ctx context.Context, userID int64, limit int) (bool, error) {
	var count int
	query := `
	SELECT COUNT(*) 
	FROM compliment_history
	WHERE user_id = $1 AND created_at >= current_date
	`
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count < limit, nil
}

func (s *Storage) RecordCompliment(ctx context.Context, userID int64, complimentID int) error {
	query := `
	INSERT INTO compliment_history (user_id, compliment_id)
	VALUES ($1, $2)
	`
	_, err := s.DB.ExecContext(ctx, query, userID, complimentID)
	return err
}

func (s *Storage) GetNextCompliment(ctx context.Context) (int, string, error) {
	query := `
	WITH next AS (
	    SELECT id, text
	    FROM compliments
	    WHERE count = (SELECT MIN(count) FROM compliments)
	    ORDER BY random()
	    LIMIT 1
	)
	UPDATE compliments c
	SET count = count + 1
	FROM next n
	WHERE c.id = n.id
	RETURNING n.id, n.text;
	`
	var id int
	var text string

	err := s.DB.QueryRowContext(ctx, query).Scan(&id, &text)
	if err != nil {
		log.Fatalf("Ошибка при получении комплимента: %v", err)
	}

	return id, text, nil
}
