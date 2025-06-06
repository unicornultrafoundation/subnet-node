package session

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// Session represents a provider session
type Session struct {
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	providerAddr common.Address
}

// New creates a new session
func New() *Session {
	ctx, cancel := context.WithCancel(context.Background())
	return &Session{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the session context
func (s *Session) Context() context.Context {
	return s.ctx
}

// Cancel cancels the session
func (s *Session) Cancel() {
	s.cancel()
}

// SetProviderAddress sets the provider address
func (s *Session) SetProviderAddress(addr common.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerAddr = addr
}

// GetProviderAddress gets the provider address
func (s *Session) GetProviderAddress() common.Address {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.providerAddr
}
