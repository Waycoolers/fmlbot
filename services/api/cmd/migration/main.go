package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/logger"
	"github.com/Waycoolers/fmlbot/services/api/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type userWithPasswordResponse struct {
	ID           int64  `db:"user_id" json:"user_id"`
	Username     string `db:"username" json:"username"`
	PartnerID    int64  `db:"partner_id" json:"partner_id"`
	PasswordHash []byte `db:"password_hash" json:"password_hash"`
}

func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/"

	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	charsetLen := len(charset)
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = charset[randomBytes[i]%byte(charsetLen)]
	}
	return string(password), nil
}

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
		slog.Error("Error connecting to migration database", "error", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		slog.Error("Error pinging migration database", "error", err)
	}

	slog.Info("Migration database connection successfully")

	ctx := context.Background()

	// GivePasswords(ctx, cfg, db)
	SetTelegramIDs(ctx, db)

	slog.Info("Migrations complete")
}

func GivePasswords(ctx context.Context, cfg *config.Config, db *sqlx.DB) {
	var users []userWithPasswordResponse
	err := db.SelectContext(ctx, &users, `
		SELECT * FROM users WHERE password_hash = '';
	`)
	if err != nil {
		slog.Error("Error selecting users", "error", err)
		os.Exit(1)
	}

	for _, user := range users {
		password, err := GenerateRandomPassword(10)
		if err != nil {
			slog.Error("Error generating password", "error", err)
			continue
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Error hashing password", "error", err)
			continue
		}

		_, err = db.ExecContext(ctx, `
			UPDATE users SET password_hash = $1 WHERE user_id = $2;
		`, string(hashedPassword), user.ID)
		if err != nil {
			slog.Error("Error updating user", "error", err)
			continue
		}

		text := fmt.Sprintf("Привет! Я усилил безопасность. Твой новый пароль: %s\n\nP.S. Он тебе может пригодится для входа в мобильное приложение. Ты всегда можешь поменять пароль в настройках аккаунта.", password)

		message := struct {
			Text   string `json:"text"`
			UserID int64  `json:"user_id"`
		}{
			Text:   text,
			UserID: user.ID,
		}
		encoded, err := json.Marshal(message)
		if err != nil {
			slog.Error("Error marshaling message", "error", err)
			continue
		}
		req, err := http.NewRequest(http.MethodPost, cfg.BotURL+"/updates/message", bytes.NewReader(encoded))
		if err != nil {
			slog.Error("Error creating request", "error", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Secret", string(cfg.Server.InternalSecret))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error sending message", "error", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			slog.Error("Unexpected status code", "code", resp.StatusCode)
		}
		_ = resp.Body.Close()

		time.Sleep(3 * time.Second)
	}
}

func SetTelegramIDs(ctx context.Context, db *sqlx.DB) {
	var users []userWithPasswordResponse
	err := db.SelectContext(ctx, &users, `
		SELECT * FROM users WHERE password_hash = '';
	`)
	if err != nil {
		slog.Error("Error selecting users", "error", err)
		os.Exit(1)
	}

	for _, user := range users {
		_, err = db.ExecContext(ctx, `
			UPDATE users SET telegram_id = $1 WHERE user_id = $1;
		`, user.ID)
		if err != nil {
			slog.Error("Error updating user", "error", err)
			continue
		}
	}
}
