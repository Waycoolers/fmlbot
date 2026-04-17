package config

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Bot       *BotConfig
	RDB       *RedisConfig
	Server    *ServerConfig
	Api       *ApiConfig
	Auth      *AuthConfig
	Loglevel  string
	JwtSecret []byte
}

type BotConfig struct {
	Token          string
	UpdatesTimeout int
}
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       string
}

type ServerConfig struct {
	Host string
	Port int
}

type ApiConfig struct {
	Host        string
	Port        int
	HTTPTimeout int
}

type AuthConfig struct {
	Host        string
	Port        int
	HTTPTimeout int
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

	bot, err := loadBotConfig()
	if err != nil {
		return nil, err
	}

	rdb, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}

	server, err := loadServerConfig()
	if err != nil {
		return nil, err
	}

	api, err := loadApiConfig()
	if err != nil {
		return nil, err
	}

	auth, err := loadAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Bot:       bot,
		RDB:       rdb,
		Server:    server,
		Api:       api,
		Auth:      auth,
		Loglevel:  loglevel,
		JwtSecret: []byte(jwtSecret),
	}, nil
}

func loadBotConfig() (*BotConfig, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("не найден TELEGRAM_BOT_TOKEN")
	}
	updatesTimeout := os.Getenv("BOT_UPDATES_TIMEOUT")
	if updatesTimeout == "" {
		updatesTimeout = "60"
		log.Printf("Не найден BOT_UPDATES_TIMEOUT")
	}
	intUpdatesTimeout, err := strconv.Atoi(updatesTimeout)
	if err != nil {
		return nil, err
	}

	return &BotConfig{
		Token:          token,
		UpdatesTimeout: intUpdatesTimeout,
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

func loadServerConfig() (*ServerConfig, error) {
	port := os.Getenv("BOT_PORT")
	if port == "" {
		port = "8080"
		slog.Warn("not found BOT_PORT")
	}
	host := os.Getenv("BOT_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found BOT_HOST")
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Host: host,
		Port: intPort,
	}, nil
}

func loadApiConfig() (*ApiConfig, error) {
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

	return &ApiConfig{
		Host:        host,
		Port:        intPort,
		HTTPTimeout: intHTTPTimeout,
	}, nil
}

func loadAuthConfig() (*AuthConfig, error) {
	host := os.Getenv("AUTH_HOST")
	if host == "" {
		host = "localhost"
		slog.Warn("not found AUTH_HOST")
	}
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8080"
		slog.Warn("not found AUTH_PORT")
	}
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	httpTimeout := os.Getenv("AUTH_HTTP_TIMEOUT")
	if httpTimeout == "" {
		httpTimeout = "60"
		slog.Warn("not found AUTH_HTTP_TIMEOUT")
	}
	intHTTPTimeout, err := strconv.Atoi(httpTimeout)
	if err != nil {
		return nil, err
	}

	return &AuthConfig{
		Host:        host,
		Port:        intPort,
		HTTPTimeout: intHTTPTimeout,
	}, nil
}
