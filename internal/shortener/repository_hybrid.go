package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type HybridLinkRepository struct {
	postgres *PostgresRepository
	redis    *RedisRepository
}

func NewHybridLinkRepository(postgres *PostgresRepository, redis *RedisRepository) *HybridLinkRepository {
	return &HybridLinkRepository{
		postgres: postgres,
		redis:    redis,
	}
}

func (r *HybridLinkRepository) TempSave(ctx context.Context, link *domain.TemporaryLink, ttl time.Duration) error {
	linkExists, err := r.exists(ctx, link.Code)
	if err != nil {
		return err
	}
	if linkExists {
		return ErrRecordAlreadyExists
	}
	err = r.redis.Save(ctx, link, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (r *HybridLinkRepository) PermSave(ctx context.Context, link *domain.PermanentLink) error {
	linkExists, err := r.exists(ctx, link.Code)
	if err != nil {
		return err
	}
	if linkExists {
		return ErrRecordAlreadyExists
	}
	err = r.postgres.Save(ctx, link)
	if err != nil {
		return err
	}
	_ = r.redis.Save(ctx, &domain.TemporaryLink{
		OriginalURL: link.OriginalURL,
		Code:        link.Code,
	}, 24*time.Hour)

	return nil
}

func (r *HybridLinkRepository) Get(ctx context.Context, code string) (domain.Link, error) {
	link, err := r.redis.Get(ctx, code)
	if err != nil {
		if !errors.Is(err, ErrRecordNotFound) {
			slog.ErrorContext(ctx, "redis error", "err", err)
		}
	} else {
		return link, nil
	}

	linkdb, err := r.postgres.Get(ctx, code)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("error obtaining link from postgres: %w", err)
	}

	_ = r.redis.Save(ctx, &domain.TemporaryLink{
		Code:        linkdb.GetCode(),
		OriginalURL: linkdb.GetOriginalURL(),
	}, 24*time.Hour)

	return linkdb, nil
}

func (r *HybridLinkRepository) exists(ctx context.Context, code string) (bool, error) {
	exists, err := r.postgres.Exists(ctx, code)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	_, err = r.redis.Get(ctx, code)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrRecordNotFound) {
		return false, nil
	}

	return false, err
}
