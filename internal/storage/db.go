package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(cfg *config.DatabaseConfig) (*Storage, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	// Проверим подключение
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("БД успешно подключена")
	return &Storage{DB: pool}, nil
}
