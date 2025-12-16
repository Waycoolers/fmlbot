package redis_store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Waycoolers/fmlbot/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ImportantDateDraftStore struct {
	rdb       *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewImportantDateDraftStore(rdb *redis.Client, ttl time.Duration) *ImportantDateDraftStore {
	return &ImportantDateDraftStore{
		rdb:       rdb,
		keyPrefix: "important_date:draft:",
		ttl:       ttl,
	}
}

func (s *ImportantDateDraftStore) key(userID int64) string {
	return fmt.Sprintf("%s%d", s.keyPrefix, userID)
}

func (s *ImportantDateDraftStore) Get(ctx context.Context, userID int64) (*domain.ImportantDateDraft, error) {
	data, err := s.rdb.Get(ctx, s.key(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var draft domain.ImportantDateDraft
	if er := json.Unmarshal([]byte(data), &draft); er != nil {
		return nil, er
	}
	return &draft, nil
}

func (s *ImportantDateDraftStore) Save(ctx context.Context, userID int64, draft *domain.ImportantDateDraft) error {
	draft.CreatedAt = time.Now()
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(userID), data, s.ttl).Err()
}

func (s *ImportantDateDraftStore) Delete(ctx context.Context, userID int64) error {
	return s.rdb.Del(ctx, s.key(userID)).Err()
}
