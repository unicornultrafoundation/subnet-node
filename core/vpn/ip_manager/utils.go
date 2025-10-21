package ipmanager

import (
	"context"
	"fmt"
	"time"
)

// Clock is a tiny abstraction to enable deterministic tests of time-based logic.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type realClock struct{}

func (realClock) Now() time.Time                  { return time.Now() }
func (realClock) Since(t time.Time) time.Duration { return time.Since(t) }

// withRetry runs a function with exponential backoff and context cancellation.
func withRetry(ctx context.Context, attempts int, fn func() error) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := fn(); err != nil {
			lastErr = err
			if i == attempts-1 {
				return fmt.Errorf("after %d attempts: %w", attempts, err)
			}
			log.WithField("attempt", i+1).WithError(err).Warn("operation failed, retrying...")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}
		return nil
	}
	return lastErr
}
