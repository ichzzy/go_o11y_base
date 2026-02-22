package repository

import (
	"context"
	"time"

	"github.com/ichzzy/go_o11y_base/internal/domain/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
)

type NewRedisClientRepositoryParam struct {
	dig.In
	RDB *redis.ClusterClient `name:"redis-cluster"`
}

type RedisClientRepository struct {
	rdb *redis.ClusterClient
}

func NewRedisClientRepository(param NewRedisClientRepositoryParam) repository.RedisClientRepository {
	return &RedisClientRepository{rdb: param.RDB}
}

func (r *RedisClientRepository) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *RedisClientRepository) Del(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisClientRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.rdb.Set(ctx, key, value, ttl).Err()
}
