package config

import (
	"errors"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token string
	DB    *DatabaseConfig
	RDB   *RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("не найден TELEGRAM_BOT_TOKEN")
	}

	db, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	rdb, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Token: token,
		DB:    db,
		RDB:   rdb,
	}, nil
}

func loadDatabaseConfig() (*DatabaseConfig, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found DB_HOST")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
		slog.Warn("not found DB_PORT")
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
		slog.Warn("not found DB_USER")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
		slog.Warn("not found DB_PASSWORD")
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "fmlbot"
		slog.Warn("not found DB_NAME")
	}

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}, nil
}

func loadRedisConfig() (*RedisConfig, error) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found REDIS_HOST")
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
		slog.Warn("not found REDIS_PORT")
	}
	password := os.Getenv("REDIS_PASSWORD")
	name := os.Getenv("REDIS_DB")
	if name == "" {
		name = "0"
		slog.Warn("not found REDIS_DB")
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       name,
	}, nil
}
