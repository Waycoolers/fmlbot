package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/logger"
	"github.com/Waycoolers/fmlbot/services/api/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load("../../.env")

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.Loglevel)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error("Error connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("Error pinging database", "error", err)
	}
	slog.Info("Database connection successfully")

	ctx := context.Background()

	// Начинаем транзакцию, чтобы все данные вставились согласованно
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		os.Exit(1)
	}

	// 1. Очистка старых данных (если скрипт запускается не первый раз)
	modID := int64(12345678)
	partnerID := int64(87654321)

	tx.ExecContext(ctx, "DELETE FROM user_config WHERE user_id IN ($1, $2)", modID, partnerID)
	tx.ExecContext(ctx, "DELETE FROM important_dates WHERE user_id IN ($1, $2)", modID, partnerID)
	tx.ExecContext(ctx, "DELETE FROM user_compliment WHERE user_id IN ($1, $2)", modID, partnerID)
	tx.ExecContext(ctx, "DELETE FROM users WHERE user_id IN ($1, $2)", modID, partnerID)

	// 2. Генерация реального хэша для пароля "123456"
	password := "12345678"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		slog.Error("Error generating bcrypt hash", "error", err)
		os.Exit(1)
	}
	realHash := string(hashedBytes)

	// 3. Создаем Модератора и его Партнера, связываем их сразу
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (user_id, username, password_hash, partner_id) 
		VALUES 
			($1, 'moderator', $3, $2),
			($2, 'mod_partner', $3, $1);
	`, modID, partnerID, realHash)
	if err != nil {
		tx.Rollback()
		slog.Error("Error inserting users", "error", err)
		os.Exit(1)
	}

	// 4. Создаем базовые конфиги, чтобы приложение не упало
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_config (user_id, compliment_token_bucket, max_compliment_count) 
		VALUES 
			($1, 2, 1),
			($2, 2, 1);
	`, modID, partnerID)
	if err != nil {
		tx.Rollback()
		slog.Error("Error inserting configs", "error", err)
		os.Exit(1)
	}

	// 5. Добавляем важную дату
	date := time.Date(2022, time.May, 25, 0, 0, 0, 0, time.UTC)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO important_dates (user_id, partner_id, title, date, notify_before_days, is_active)
		VALUES ($1, $2, 'Годовщина установки RuStore', $3, 7, true);
	`, modID, partnerID, date)
	if err != nil {
		tx.Rollback()
		slog.Error("Error inserting important date", "error", err)
		os.Exit(1)
	}

	// 6. Добавляем комплименты
	// Комплимент 1: От партнера модератору (уже получен)
	var comp1ID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO compliments (text, is_sent) 
		VALUES ('Привет от команды!', true) 
		RETURNING id;
	`).Scan(&comp1ID)
	if err == nil {
		tx.ExecContext(ctx, "INSERT INTO user_compliment (user_id, compliment_id) VALUES ($1, $2)", partnerID, comp1ID)
	}

	// Комплимент 2: От партнера модератору (ждет отправки)
	var comp2ID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO compliments (text, is_sent) 
		VALUES ('Твой первый комплимент в этом приложении', false) 
		RETURNING id;
	`).Scan(&comp2ID)
	if err == nil {
		tx.ExecContext(ctx, "INSERT INTO user_compliment (user_id, compliment_id) VALUES ($1, $2)", partnerID, comp2ID)
	}

	// Фиксируем изменения в БД
	err = tx.Commit()
	if err != nil {
		slog.Error("Error committing transaction", "error", err)
		os.Exit(1)
	}

	slog.Info("SUCCESS! Moderator and Partner created with fully populated test data!")
}
