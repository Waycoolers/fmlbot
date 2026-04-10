package domain

import (
	"context"
	"time"
)

type TokensRepo interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	Validate(ctx context.Context, fullToken string) (int64, error)
	Revoke(ctx context.Context, fullToken string) error
}
