package redis_store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/redis/go-redis/v9"
)

type IdeaDraftStore struct {
	rdb       *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewIdeaDraftStore(rdb *redis.Client, ttl time.Duration) *IdeaDraftStore {
	return &IdeaDraftStore{
		rdb:       rdb,
		keyPrefix: "idea_draft:",
		ttl:       ttl,
	}
}

func (s *IdeaDraftStore) key(userID int64) string {
	return fmt.Sprintf("%s%d", s.keyPrefix, userID)
}

func (s *IdeaDraftStore) Get(ctx context.Context, userID int64) (*domain.IdeaDraft, error) {
	data, err := s.rdb.Get(ctx, s.key(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var draft domain.IdeaDraft
	if er := json.Unmarshal([]byte(data), &draft); er != nil {
		return nil, er
	}
	return &draft, nil
}

func (s *IdeaDraftStore) Save(ctx context.Context, userID int64, draft *domain.IdeaDraft) error {
	draft.CreatedAt = time.Now()
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(userID), data, s.ttl).Err()
}

func (s *IdeaDraftStore) Delete(ctx context.Context, userID int64) error {
	return s.rdb.Del(ctx, s.key(userID)).Err()
}
