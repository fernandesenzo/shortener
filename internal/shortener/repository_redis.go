package shortener

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/redis/go-redis/v9"
)

const linkPrefix = "link:"

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client}
}
func (r *RedisRepository) Save(ctx context.Context, link *domain.TemporaryLink, ttl time.Duration) error {
	key := linkPrefix + link.Code

	_, err := r.client.Set(ctx, key, link.OriginalURL, ttl).Result()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, code string) (*domain.TemporaryLink, error) {
	key := linkPrefix + code

	url, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("unexpected error when getting from redis: %w", err)
	}
	return &domain.TemporaryLink{
		Code:        code,
		OriginalURL: url,
	}, nil
}
