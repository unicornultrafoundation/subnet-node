package ipmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/static"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

const (
	STATIC_CHECK_INTERVAL = 30 * time.Second
	STATIC_CHECK_TIMEOUT  = 10 * time.Second
)

type StaticIPManager struct {
	client  static.StaticIPClient
	ip      string
	peerID  string
	running bool
	stateCh chan IPManagerStateType
	ipCh    chan string
	mu      sync.RWMutex
	stopCh  chan struct{}
}

func NewStaticIPManager(client static.StaticIPClient, ip string, peerID string, stateCh chan IPManagerStateType, ipCh chan string) IPManagerState {
	return &StaticIPManager{
		client:  client,
		ip:      ip,
		running: false,
		stateCh: stateCh,
		ipCh:    ipCh,
		peerID:  peerID,
		stopCh:  make(chan struct{}),
	}
}

func (s *StaticIPManager) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	go s.start(ctx)
	return nil
}

func (s *StaticIPManager) start(ctx context.Context) {
	// Try to update IP before starting
	if !s.tryUpdateIP(ctx) {
		return
	}

	// Start periodic checks
	tick := time.NewTicker(STATIC_CHECK_INTERVAL)
	defer func() {
		tick.Stop()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			// Check IP with timeout
			checkCtx, cancel := context.WithTimeout(ctx, STATIC_CHECK_TIMEOUT)
			ok := s.tryUpdateIP(checkCtx)
			cancel()

			if !ok {
				return
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *StaticIPManager) tryUpdateIP(ctx context.Context) bool {
	// Check IP status
	err := s.checkIP(ctx)
	if err != nil {
		log.WithField("ip", s.ip).WithError(err).Error("failed to check static IP status")
		s.NextState()
		return false
	}

	// Update IP
	select {
	case s.ipCh <- s.ip:
	default:
		log.Warn("IP channel is full, dropping IP update")
	}

	return true
}

func (s *StaticIPManager) checkIP(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if s.ip == "" {
		return fmt.Errorf("IP is empty")
	}

	if utils.GetIPType(s.ip) != utils.STATIC_IP_TYPE {
		return fmt.Errorf("IP is not a static IP")
	}

	owned, err := s.client.IsIPOwnedByNode(ctx, s.ip)
	if err != nil {
		return fmt.Errorf("ownership check failed: %w", err)
	}

	if !owned {
		return fmt.Errorf("IP %s is not owned by node", s.ip)
	}

	return nil
}

func (s *StaticIPManager) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	s.running = false
	close(s.stopCh)
	return nil
}

func (s *StaticIPManager) GetType() IPManagerStateType {
	return IPManagerStateTypeStatic
}

func (s *StaticIPManager) NextState() {
	log.Info("Switching to dynamic IP")
	s.stateCh <- IPManagerStateTypeDynamic
}
