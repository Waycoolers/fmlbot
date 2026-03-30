package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Server   *ServerConfig
	DB       *DatabaseConfig
	Loglevel string
}

type ServerConfig struct {
	Host string
	Port string
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

	server, err := loadServerConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:   server,
		Loglevel: loglevel,
	}, nil
}

func loadServerConfig() (*ServerConfig, error) {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found SERVER_HOST")
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		slog.Warn("not found SERVER_PORT")
	}
	return &ServerConfig{host, port}, nil
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
