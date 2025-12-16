package storage

import (
	"fmt"
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DB *sqlx.DB
}

func New(cfg *config.DatabaseConfig) (*Storage, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	if er := db.Ping(); er != nil {
		log.Fatalf("Ошибка пинга в БД: %v", er)
	}

	log.Println("БД успешно подключена")
	return &Storage{DB: db}, nil
}
