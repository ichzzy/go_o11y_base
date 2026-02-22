package repository

import (
	"context"
	"time"
)

type RedisClientRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
