package storage

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db     *sqlx.DB
	Tokens domain.TokensRepo
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
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	slog.Info("Database connection successfully")

	tokens := tokensRepo{db: db}

	return &Storage{
		db:     db,
		Tokens: &tokens,
	}, nil
}

func (s *Storage) Close() {
	err := s.db.Close()
	if err != nil {
		slog.Error("Error closing database", "error", err)
	}
}
