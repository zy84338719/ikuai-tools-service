package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
)

// Client is the process-wide Redis client (nil when Redis is not configured or
// unreachable). Callers that want to use Redis should nil-check Client first.
var Client *redis.Client

// Init opens the connection and pings it. Returns an error on failure; the
// caller decides whether Redis is required (the service treats it as optional).
func Init(cfg *conf.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}
	return nil
}

// Close closes the client if it was opened.
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Ping reports whether Redis is reachable. Used by the /ready probe.
func Ping(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("redis not configured")
	}
	return Client.Ping(ctx).Err()
}
