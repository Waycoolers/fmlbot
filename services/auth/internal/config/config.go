package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DB              *DatabaseConfig
	Server          *ServerConfig
	API             *APIConfig
	Loglevel        string
	JwtSecret       []byte
	InternalSecret  []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
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

type APIConfig struct {
	Host        string
	Port        int
	HTTPTimeout time.Duration
}

func Load() (*Config, error) {
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "info"
		slog.Warn("not found LOG_LEVEL")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("not found JWT_SECRET")
		return nil, errors.New("no JWT_SECRET")
	}
	if len(jwtSecret) < 32 {
		slog.Warn("JWT_SECRET is too short, minimum 32 bytes recommended")
	}

	accessTokenTTL := os.Getenv("ACCESS_TOKEN_TTL")
	if accessTokenTTL == "" {
		accessTokenTTL = "30"
		slog.Warn("not found ACCESS_TOKEN_TTL")
	}
	intAccessTokenTTL, err := strconv.Atoi(accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshTokenTTL := os.Getenv("REFRESH_TOKEN_TTL")
	if refreshTokenTTL == "" {
		refreshTokenTTL = "30"
		slog.Warn("not found REFRESH_TOKEN_TTL")
	}
	intRefreshTokenTTL, err := strconv.Atoi(refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	secret := os.Getenv("INTERNAL_API_SECRET")
	if secret == "" {
		slog.Error("not found INTERNAL_API_SECRET")
		return nil, errors.New("no INTERNAL_API_SECRET")
	}

	server, err := loadServerConfig()
	if err != nil {
		return nil, err
	}

	db, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	api, err := loadAPIConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		DB:              db,
		Server:          server,
		API:             api,
		Loglevel:        loglevel,
		JwtSecret:       []byte(jwtSecret),
		InternalSecret:  []byte(secret),
		AccessTokenTTL:  time.Duration(intAccessTokenTTL) * time.Minute,
		RefreshTokenTTL: time.Duration(intRefreshTokenTTL) * 24 * time.Hour,
	}, nil
}

func loadServerConfig() (*ServerConfig, error) {
	host := os.Getenv("AUTH_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found AUTH_HOST")
	}
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8081"
		slog.Warn("not found AUTH_PORT")
	}
	return &ServerConfig{
		Host: host,
		Port: port,
	}, nil
}

func loadDatabaseConfig() (*DatabaseConfig, error) {
	host := os.Getenv("AUTH_DB_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found AUTH_DB_HOST")
	}
	port := os.Getenv("AUTH_DB_PORT")
	if port == "" {
		port = "5432"
		slog.Warn("not found AUTH_DB_PORT")
	}
	user := os.Getenv("AUTH_DB_USER")
	if user == "" {
		user = "postgres"
		slog.Warn("not found AUTH_DB_USER")
	}
	password := os.Getenv("AUTH_DB_PASSWORD")
	if password == "" {
		password = "postgres"
		slog.Warn("not found AUTH_DB_PASSWORD")
	}
	name := os.Getenv("AUTH_DB_NAME")
	if name == "" {
		name = "fmlbot_auth"
		slog.Warn("not found AUTH_DB_NAME")
	}

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}, nil
}

func loadAPIConfig() (*APIConfig, error) {
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
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	httpTimeout := os.Getenv("API_HTTP_TIMEOUT")
	if httpTimeout == "" {
		httpTimeout = "60"
		slog.Warn("not found HTTP_TIMEOUT")
	}
	intHTTPTimeout, err := strconv.Atoi(httpTimeout)
	if err != nil {
		return nil, err
	}

	return &APIConfig{
		Host:        host,
		Port:        intPort,
		HTTPTimeout: time.Duration(intHTTPTimeout) * time.Second,
	}, nil
}
