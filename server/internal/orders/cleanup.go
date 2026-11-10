package orders

import (
	"context"
	"log/slog"
	"time"
)

// cleanupBatch caps how many orders one run expires, so a backlog after a long sleep is
// worked off over a few runs instead of one long burst of transactions.
const cleanupBatch = 100

// RunCleanup expires stale orders once at start and then every `every`, until ctx ends.
// It is tidy-up only (PLAN.md section 5, step 3): on Render's free tier the server sleeps and
// this doesn't run, and nothing breaks, because expired holds are claimable without it.
func (s *Service) RunCleanup(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		n, err := s.CleanupOnce(ctx)
		if err != nil && ctx.Err() == nil {
			slog.Error("order cleanup", "err", err)
		} else if n > 0 {
			slog.Info("order cleanup", "expired", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// CleanupOnce expires one batch of pending orders whose hold ran out and returns how many.
func (s *Service) CleanupOnce(ctx context.Context) (int, error) {
	ids, err := s.repo.Stale(ctx, s.now(), cleanupBatch)
	if err != nil {
		return 0, err
	}
	for i, id := range ids {
		if err := s.expire(ctx, id); err != nil {
			return i, err
		}
	}
	return len(ids), nil
}
