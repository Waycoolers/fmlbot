package config

import (
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	Server   *ServerConfig
	DB       *DatabaseConfig
	Loglevel string
	BotURL   string
}

type ServerConfig struct {
	Host      string
	Port      string
	JwtSecret []byte
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() (*Config, error) {
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "info"
		slog.Warn("not found LOG_LEVEL")
	}

	botURL := os.Getenv("BOT_URL")
	if botURL == "" {
		slog.Error("not found BOT_URL")
		return nil, errors.New("no BOT_URL")
	}

	server, err := loadServerConfig()
	if err != nil {
		return nil, err
	}

	db, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:   server,
		DB:       db,
		Loglevel: loglevel,
		BotURL:   os.Getenv("BOT_URL"),
	}, nil
}

func loadServerConfig() (*ServerConfig, error) {
	host := os.Getenv("API_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found API_HOST")
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
		slog.Warn("not found API_PORT")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("not found JWT_SECRET")
		return nil, errors.New("no JWT_SECRET")
	}
	return &ServerConfig{
		Host:      host,
		Port:      port,
		JwtSecret: []byte(jwtSecret),
	}, nil
}

func loadDatabaseConfig() (*DatabaseConfig, error) {
	host := os.Getenv("API_DB_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found API_DB_HOST")
	}
	port := os.Getenv("API_DB_PORT")
	if port == "" {
		port = "5432"
		slog.Warn("not found API_DB_PORT")
	}
	user := os.Getenv("API_DB_USER")
	if user == "" {
		user = "postgres"
		slog.Warn("not found API_DB_USER")
	}
	password := os.Getenv("API_DB_PASSWORD")
	if password == "" {
		password = "postgres"
		slog.Warn("not found API_DB_PASSWORD")
	}
	name := os.Getenv("API_DB_NAME")
	if name == "" {
		name = "fmlbot"
		slog.Warn("not found API_DB_NAME")
	}

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}, nil
}
