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

type APIClient interface {
	VerifyUser(ctx context.Context, username string, password string) (int64, error)
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)
	Register(ctx context.Context, userID int64, username string, password string) error
}
