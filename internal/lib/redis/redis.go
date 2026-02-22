package redis

import (
	"context"
	"fmt"

	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func NewRedisCluster(ctx context.Context, config *config.Config) (*redis.ClusterClient, error) {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           config.Connections.Redis.Cluster.Addrs,
		Username:        config.Connections.Redis.Cluster.Username,
		Password:        config.Connections.Redis.Cluster.Password,
		DialTimeout:     config.Connections.Redis.Cluster.DialTimeout,
		ReadTimeout:     config.Connections.Redis.Cluster.ReadTimeout,
		WriteTimeout:    config.Connections.Redis.Cluster.WriteTimeout,
		PoolSize:        config.Connections.Redis.Cluster.PoolSize,
		MinIdleConns:    config.Connections.Redis.Cluster.MinIdleConns,
		MaxIdleConns:    config.Connections.Redis.Cluster.MaxIdleConns,
		ConnMaxIdleTime: config.Connections.Redis.Cluster.ConnMaxIdleTime,
	})

	if config.Observability.OTEL.Enabled {
		if err := redisotel.InstrumentTracing(rdb); err != nil {
			return nil, fmt.Errorf("redisotel.InstrumentTracing failed: %w", err)
		}
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping error: %w", err)
	}

	return rdb, nil
}
