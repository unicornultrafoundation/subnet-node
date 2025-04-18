package vpn

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RetryOperation retries an operation with exponential backoff
func (s *Service) RetryOperation(ctx context.Context, operation func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 30 * time.Second

	return backoff.Retry(operation, backoff.WithContext(bo, ctx))
}
