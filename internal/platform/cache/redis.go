package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: url,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			slog.Error("failed closing redis client after error on ping", "error", closeErr)
		}
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
