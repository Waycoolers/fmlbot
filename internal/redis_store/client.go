package redis_store

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/redis/go-redis/v9"
)

func New(cfg *config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	db, _ := strconv.Atoi(cfg.DB)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       db,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	} else {
		log.Print("Redis подключен")
	}

	return rdb, nil
}
