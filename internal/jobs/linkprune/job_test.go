package linkprune_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fernandesenzo/shortener/internal/jobs/linkprune"
)

type SpyPruner struct {
	callCount int
	mutex     sync.Mutex
}

func (s *SpyPruner) PruneExpired(ctx context.Context, limit time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.callCount++
	return nil
}

func TestJobRun(t *testing.T) {
	spy := &SpyPruner{}
	job := linkprune.NewJob(spy, time.Millisecond*10, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())

	go job.Run(ctx)
	time.Sleep(time.Millisecond * 100)
	cancel()
	time.Sleep(time.Millisecond * 10)

	spy.mutex.Lock()
	count := spy.callCount
	spy.mutex.Unlock()
	if count < 5 {
		t.Fatalf("expected more frequent calls, got %d", spy.callCount)
	}
}
