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
}

func NewStaticIPManager(client static.StaticIPClient, ip string, peerID string, stateCh chan IPManagerStateType, ipCh chan string) IPManagerState {
	return &StaticIPManager{client: client, ip: ip, running: false, stateCh: stateCh, ipCh: ipCh, peerID: peerID}
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

func (s *StaticIPManager) start(ctx context.Context) error {
	// Validate IP before starting
	if err := s.checkIP(ctx); err != nil {
		log.WithField("ip", s.ip).WithError(err).Error("failed to check IP ownership")
		return err
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
			return nil
		case <-tick.C:
			// Check IP with timeout
			checkCtx, cancel := context.WithTimeout(ctx, STATIC_CHECK_TIMEOUT)
			err := s.checkIP(checkCtx)
			cancel()

			if err != nil {
				log.WithField("ip", s.ip).WithError(err).Error("failed to check IP ownership")
				return err
			}
		}
	}
}

func (s *StaticIPManager) checkIP(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if s.ip == "" {
		return fmt.Errorf("IP is empty")
	}

	if utils.GetIPType(s.ip) != utils.STATIC_IP_TYPE {
		log.WithField("ip", s.ip).Warn("IP is not a static IP, switching to dynamic IP")
		s.NextState()
		return fmt.Errorf("IP is not a static IP")
	}

	owned, err := s.client.IsIPOwnedByNode(ctx, s.ip)
	if err != nil {
		log.WithField("ip", s.ip).WithError(err).Errorf("failed to check if IP is owned by the node, switching to dynamic IP")
		s.NextState()
		return fmt.Errorf("ownership check failed: %w", err)
	}

	if !owned {
		log.WithField("ip", s.ip).Warn("IP is not owned by the node, switching to dynamic IP")
		s.NextState()
		return fmt.Errorf("IP %s is not owned by node", s.ip)
	}

	// Send IP update
	select {
	case s.ipCh <- s.ip:
	default:
		log.Warn("IP channel is full, dropping IP update")
	}

	return nil
}

func (s *StaticIPManager) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()
	return nil
}

func (s *StaticIPManager) GetType() IPManagerStateType {
	return IPManagerStateTypeStatic
}

func (s *StaticIPManager) NextState() {
	s.stateCh <- IPManagerStateTypeDynamic
}
