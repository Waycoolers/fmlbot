package storage

import (
	"context"
	"log"
)

func (s *Storage) GetNextCompliment(ctx context.Context) (string, error) {
	var compliment string

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
	RETURNING n.text;
	`

	err := s.DB.QueryRow(ctx, query).Scan(&compliment)
	if err != nil {
		log.Fatalf("Ошибка при получении комплимента: %v", err)
	}
	return compliment, nil
}
