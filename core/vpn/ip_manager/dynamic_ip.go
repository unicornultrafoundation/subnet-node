package ipmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/dynamic"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

const (
	MAX_ATTEMPTS     = 10
	REQUEST_TIMEOUT  = 30 * time.Second
	RENEW_TIMEOUT    = 30 * time.Second
	RELEASE_TIMEOUT  = 30 * time.Second
	RENEW_PERCENTAGE = 0.8
)

type DynamicIPManager struct {
	running bool
	client  dynamic.DynamicIPClient
	stateCh chan IPManagerStateType
	ipCh    chan string
	lease   dynamic.Lease
	// leaseStart marks when the current lease was received/renewed on this client.
	// It leverages Go's monotonic clock via time.Since for stable elapsed tracking.
	leaseStart time.Time
	// leaseTTL is the remaining time-to-live for the lease at the moment of receipt.
	// It should be populated from server-provided ttl seconds.
	leaseTTL time.Duration
	clk      Clock
	mu       sync.RWMutex

	// Manager context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDynamicIPManager(client dynamic.DynamicIPClient, stateCh chan IPManagerStateType, ipCh chan string) IPManagerState {
	return &DynamicIPManager{client: client, running: false, stateCh: stateCh, ipCh: ipCh, clk: realClock{}}
}

func (d *DynamicIPManager) Start(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = true
	d.mu.Unlock()

	d.ctx, d.cancel = context.WithCancel(context.Background())
	go d.start(d.ctx)
	return nil
}

func (d *DynamicIPManager) Stop(ctx context.Context) error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return nil
	}

	// Release IP with timeout
	releaseCtx, releaseCancel := context.WithTimeout(ctx, RELEASE_TIMEOUT)
	defer releaseCancel()

	if err := d.tryReleaseIP(releaseCtx); err != nil {
		log.WithError(err).Warn("failed to release IP")
	}
	d.running = false

	d.cancel()
	d.mu.Unlock()
	return nil
}

func (d *DynamicIPManager) GetType() IPManagerStateType {
	return IPManagerStateTypeDynamic
}

func (d *DynamicIPManager) NextState() {
	// Do nothing
}

func (d *DynamicIPManager) start(ctx context.Context) error {
	// Request initial IP with timeout
	requestCtx, cancel := context.WithTimeout(ctx, REQUEST_TIMEOUT)
	defer cancel()

	err := d.tryRequestIP(requestCtx)
	if err != nil {
		log.WithError(err).Error("failed to request initial IP")
		return err
	}

	// Check expiry every minute
	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			// Check expiry with timeout
			expiryCtx, expiryCancel := context.WithTimeout(ctx, 20*time.Second)
			err := d.checkExpiry(expiryCtx)
			expiryCancel()

			if err != nil {
				log.WithError(err).Warn("failed to check expiry")

				// Try to request a new IP with timeout
				requestCtx, requestCancel := context.WithTimeout(ctx, REQUEST_TIMEOUT)
				err := d.tryRequestIP(requestCtx)
				requestCancel()

				if err != nil {
					log.WithError(err).Error("failed to request IP after expiry check failure")
					return err
				}
			}
		}
	}
}

// validateLease performs basic sanity checks on the lease payload.
func validateLease(lease dynamic.Lease) error {
	if lease.TokenID <= 0 {
		return fmt.Errorf("invalid token ID from client: %d", lease.TokenID)
	}
	if lease.CreatedAt.IsZero() || lease.ExpiresAt.IsZero() || lease.UpdatedAt.IsZero() {
		return fmt.Errorf("invalid lease data: created_at=%v, expires_at=%v", lease.CreatedAt, lease.ExpiresAt)
	}
	if lease.ExpiresAt.Before(lease.CreatedAt) || lease.ExpiresAt.Before(lease.UpdatedAt) {
		return fmt.Errorf("invalid lease: expires_at (%v) is before created_at (%v)", lease.ExpiresAt, lease.CreatedAt)
	}
	return nil
}

// bootstrapLeaseTiming initializes the monotonic countdown based on server-provided TTL.
// Falls back to server timestamps if TTL is unavailable.
func (d *DynamicIPManager) bootstrapLeaseTiming(lease dynamic.Lease) {
	if lease.Ttl > 0 {
		d.leaseTTL = time.Duration(lease.Ttl) * time.Second
	} else {
		d.leaseTTL = lease.ExpiresAt.Sub(lease.UpdatedAt)
		if d.leaseTTL < 0 {
			d.leaseTTL = 0
		}
	}
	d.leaseStart = d.clk.Now()
}

// applyLease validates, stores, bootstraps timing, and publishes the IP.
func (d *DynamicIPManager) applyLease(ctx context.Context, lease dynamic.Lease) error {
	if err := validateLease(lease); err != nil {
		return err
	}

	ip := utils.ConvertNumberToVirtualIP(uint32(lease.TokenID))
	if ip == "" {
		return fmt.Errorf("failed to convert token ID %d to virtual IP", lease.TokenID)
	}

	d.mu.Lock()
	d.lease = lease
	d.bootstrapLeaseTiming(lease)
	d.mu.Unlock()

	select {
	case d.ipCh <- ip:
		_ = d.client.SetIP(ctx, ip)
	default:
		log.Warn("IP channel is full, dropping IP update")
	}

	return nil
}

// shouldRenew returns true if RENEW_PERCENTAGE of the TTL elapsed or timing is uninitialized.
func (d *DynamicIPManager) shouldRenew() bool {
	d.mu.RLock()
	ttl := d.leaseTTL
	start := d.leaseStart
	d.mu.RUnlock()

	if ttl <= 0 || start.IsZero() {
		return true
	}
	elapsed := d.clk.Since(start)
	threshold := time.Duration(float64(ttl) * RENEW_PERCENTAGE)
	return elapsed > threshold
}

func (d *DynamicIPManager) tryReleaseIP(ctx context.Context) error {
	return withRetry(ctx, MAX_ATTEMPTS, func() error { return d.releaseIP(ctx) })
}

func (d *DynamicIPManager) releaseIP(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := d.client.ReleaseIP(ctx)
	if err != nil {
		return fmt.Errorf("client release IP failed: %w", err)
	}

	d.client.SetIP(ctx, "")

	select {
	case d.ipCh <- "":
	default:
		log.Warn("IP channel is full, dropping IP update")
	}

	return nil
}

func (d *DynamicIPManager) checkExpiry(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if d.shouldRenew() {
		return d.tryRenewIP(ctx)
	}
	return nil
}

func (d *DynamicIPManager) tryRenewIP(ctx context.Context) error {
	return withRetry(ctx, MAX_ATTEMPTS, func() error { return d.renewIP(ctx) })
}

func (d *DynamicIPManager) renewIP(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	lease, err := d.client.RenewIP(ctx)
	if err != nil {
		return fmt.Errorf("client renew IP failed: %w", err)
	}

	return d.applyLease(ctx, lease)
}

func (d *DynamicIPManager) tryRequestIP(ctx context.Context) error {
	return withRetry(ctx, MAX_ATTEMPTS, func() error { return d.requestIP(ctx) })
}

func (d *DynamicIPManager) requestIP(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	lease, err := d.client.RequestIP(ctx)
	if err != nil {
		return fmt.Errorf("client request IP failed: %w", err)
	}

	return d.applyLease(ctx, lease)
}
