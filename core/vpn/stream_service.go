package vpn

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

var streamLog = logrus.WithField("service", "vpn-stream")

// StreamService manages VPN streams
type StreamService struct {
	// Core components
	streamCreator api.StreamService
	ctx           context.Context
	cancel        context.CancelFunc

	// Configuration
	minStreamsPerPeer int
	streamIdleTimeout time.Duration
	cleanupInterval   time.Duration

	// Lifecycle management
	mu     sync.RWMutex
	active bool

	// Stream management
	streamsMu sync.RWMutex
	streams   map[string]api.VPNStream
}

// StreamServiceConfig contains configuration for the stream service
type StreamServiceConfig struct {
	MinStreamsPerPeer int
	StreamIdleTimeout time.Duration
	CleanupInterval   time.Duration
}

// NewStreamService creates a new stream service
func NewStreamService(streamCreator api.StreamService, config *StreamServiceConfig) *StreamService {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamService{
		streamCreator:     streamCreator,
		ctx:               ctx,
		cancel:            cancel,
		minStreamsPerPeer: config.MinStreamsPerPeer,
		streamIdleTimeout: config.StreamIdleTimeout,
		cleanupInterval:   config.CleanupInterval,
		streams:           make(map[string]api.VPNStream),
	}
}

// Start starts the stream service
func (s *StreamService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return
	}

	streamLog.Info("Starting stream service")
	s.active = true
}

// Stop stops the stream service
func (s *StreamService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	streamLog.Info("Stopping stream service")
	s.cancel()
	s.active = false
}

// CreateNewVPNStream creates a new VPN stream to a peer
func (s *StreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	return s.streamCreator.CreateNewVPNStream(ctx, peerID)
}

// GetStream gets a stream from the pool or creates a new one
func (s *StreamService) GetStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	peerIDStr := peerID.String()

	// Check if we already have a stream for this peer
	s.streamsMu.RLock()
	stream, exists := s.streams[peerIDStr]
	s.streamsMu.RUnlock()

	if exists {
		// Check if the stream is still valid
		_, err := stream.Write([]byte{})
		if err == nil {
			return stream, nil
		}

		// Stream is invalid, remove it
		s.streamsMu.Lock()
		delete(s.streams, peerIDStr)
		s.streamsMu.Unlock()
	}

	// Create a new stream
	stream, err := s.CreateNewVPNStream(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Store the stream
	s.streamsMu.Lock()
	s.streams[peerIDStr] = stream
	s.streamsMu.Unlock()

	return stream, nil
}

// ReleaseStream returns a stream to the pool
func (s *StreamService) ReleaseStream(peerID peer.ID, stream api.VPNStream, healthy bool) {
	peerIDStr := peerID.String()

	if !healthy {
		// Close the unhealthy stream
		stream.Close()

		// Remove it from the pool
		s.streamsMu.Lock()
		if s.streams[peerIDStr] == stream {
			delete(s.streams, peerIDStr)
		}
		s.streamsMu.Unlock()
	}
}

// Close implements io.Closer
func (s *StreamService) Close() error {
	s.Stop()
	return nil
}
