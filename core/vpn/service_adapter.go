package vpn

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ServiceAdapter combines multiple service interfaces into a single VPNService
// This allows components to use the specific interfaces they need while maintaining compatibility
// with the existing VPNService interface
type ServiceAdapter struct {
	PeerDiscovery PeerDiscoveryService
	Stream        StreamService
	Retry         RetryService
	Metrics       MetricsService
	Config        ConfigService
}

// Ensure ServiceAdapter implements VPNService
var _ VPNService = (*ServiceAdapter)(nil)

// GetPeerID delegates to the PeerDiscoveryService
func (a *ServiceAdapter) GetPeerID(ctx context.Context, destIP string) (string, error) {
	return a.PeerDiscovery.GetPeerID(ctx, destIP)
}

// CreateNewVPNStream delegates to the StreamService
func (a *ServiceAdapter) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error) {
	return a.Stream.CreateNewVPNStream(ctx, peerID)
}

// RetryOperation delegates to the RetryService
func (a *ServiceAdapter) RetryOperation(ctx context.Context, operation func() error) error {
	return a.Retry.RetryOperation(ctx, operation)
}

// IncrementStreamErrors delegates to the MetricsService
func (a *ServiceAdapter) IncrementStreamErrors() {
	a.Metrics.IncrementStreamErrors()
}

// IncrementPacketsSent delegates to the MetricsService
func (a *ServiceAdapter) IncrementPacketsSent(bytes int) {
	a.Metrics.IncrementPacketsSent(bytes)
}

// IncrementPacketsDropped delegates to the MetricsService
func (a *ServiceAdapter) IncrementPacketsDropped() {
	a.Metrics.IncrementPacketsDropped()
}

// GetWorkerBufferSize delegates to the ConfigService
func (a *ServiceAdapter) GetWorkerBufferSize() int {
	return a.Config.GetWorkerBufferSize()
}

// GetWorkerIdleTimeout delegates to the ConfigService
func (a *ServiceAdapter) GetWorkerIdleTimeout() int {
	return a.Config.GetWorkerIdleTimeout()
}
