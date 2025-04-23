package router

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

var streamServiceLog = logrus.WithField("service", "vpn-router-stream-service")

// StreamCreator is a function type that creates a new VPN stream to a peer
type StreamCreator func(ctx context.Context, peerID peer.ID) (api.VPNStream, error)

// RouterStreamService implements the api.StreamService interface
// It's a simple wrapper around a stream creation function
type RouterStreamService struct {
	createStreamFn StreamCreator
}

// NewRouterStreamService creates a new RouterStreamService
func NewRouterStreamService(createStreamFn StreamCreator) *RouterStreamService {
	return &RouterStreamService{
		createStreamFn: createStreamFn,
	}
}

// CreateNewVPNStream implements the api.StreamService interface
func (s *RouterStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	streamServiceLog.WithField("peer_id", peerID.String()).Debug("Creating new VPN stream")

	stream, err := s.createStreamFn(ctx, peerID)
	if err != nil {
		streamServiceLog.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"error":   err,
		}).Warn("Failed to create new VPN stream")
		return nil, err
	}

	streamServiceLog.WithField("peer_id", peerID.String()).Debug("Successfully created new VPN stream")
	return stream, nil
}

// Ensure RouterStreamService implements the api.StreamService interface
var _ api.StreamService = (*RouterStreamService)(nil)
