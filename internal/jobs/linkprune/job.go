package linkprune

import (
	"context"
	"log/slog"
	"time"
)

type Pruner interface {
	PruneExpired(ctx context.Context, limit time.Duration) error
}

type Job struct {
	pruner        Pruner
	interval      time.Duration
	expirationAge time.Duration
}

func NewJob(pruner Pruner, interval time.Duration, expirationAge time.Duration) *Job {
	return &Job{pruner, interval, expirationAge}
}

func (j *Job) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)

	for {
		select {
		case <-ticker.C:
			if err := j.pruner.PruneExpired(ctx, j.expirationAge); err != nil {
				slog.ErrorContext(ctx, "failed to prune expired links", "error", err)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}

	}
}
