package dispatcher

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
)

var log = logrus.WithField("service", "vpn-dispatcher")

// Config holds configuration for the dispatcher
type Config struct {
	// Stream cleanup configuration
	StreamCleanupInterval time.Duration
	StreamIdleTimeout     time.Duration
}

// Dispatcher implements the DispatcherService interface
type Dispatcher struct {
	// Peer discovery service
	peerDiscovery *discovery.PeerDiscovery
	// Stream service for creating new streams
	streamService api.StreamService
	// Configuration service
	configService vpnconfig.ConfigService
	// Map of peer ID to stream
	streams sync.Map // string -> api.VPNStream
	// Stream cleanup ticker
	cleanupTicker *time.Ticker
	// Metrics logging ticker
	metricsLogTicker *time.Ticker
	// Metrics sampling ticker
	metricsSamplingTicker *time.Ticker
	// Stop channel for background goroutines
	stopChan chan struct{}
	// Last used time for each stream
	lastUsed sync.Map // string -> time.Time

	// Metrics
	packetsDispatched   atomic.Uint64
	bytesDispatched     atomic.Uint64
	streamsCreated      atomic.Uint64
	streamsClosed       atomic.Uint64
	streamErrors        atomic.Uint64
	peerDiscoveryErrors atomic.Uint64
	packetParseErrors   atomic.Uint64

	packetsDispatchedByPeer sync.Map // string -> *atomic.Uint64
	bytesDispatchedByPeer   sync.Map // string -> *atomic.Uint64
	streamErrorsByPeer      sync.Map // string -> *atomic.Uint64
	activeStreamsByPeer     sync.Map // string -> *atomic.Uint64
	streamCreationTimes     sync.Map // string -> time.Time

	oldestStreamCreationTime time.Time
	newestStreamCreationTime time.Time

	// Sampled metrics
	sampledMetricsMu sync.RWMutex
	sampledMetrics   *SampledMetrics
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(
	peerDiscovery *discovery.PeerDiscovery,
	streamService api.StreamService,
	configService vpnconfig.ConfigService,
) DispatcherService {
	return &Dispatcher{
		peerDiscovery: peerDiscovery,
		streamService: streamService,
		configService: configService,
		streams:       sync.Map{},
		stopChan:      make(chan struct{}),
		lastUsed:      sync.Map{},

		// Initialize time values
		oldestStreamCreationTime: time.Now(),
		newestStreamCreationTime: time.Now(),
	}
}

// Start starts the dispatcher
func (d *Dispatcher) Start() {
	log.Info("Starting VPN dispatcher")

	// Start the cleanup ticker
	cleanupInterval := d.configService.GetStreamCleanupInterval()
	d.cleanupTicker = time.NewTicker(cleanupInterval)

	// Start the cleanup goroutine
	go d.cleanupStreams()

	log.Infof("Stream cleanup scheduled every %s", cleanupInterval)

	// Start the metrics sampling
	d.startMetricsSampling()

	// Start the metrics logging ticker
	metricsInterval := d.configService.GetMetricsLogInterval()
	d.metricsLogTicker = time.NewTicker(metricsInterval)

	// Start the metrics logging goroutine
	go d.logMetricsPeriodically()

	log.Infof("Metrics logging scheduled every %s", metricsInterval)
}

// Close stops the dispatcher
func (d *Dispatcher) Close() error {
	log.Info("Stopping VPN dispatcher")

	// Stop the cleanup ticker if it exists
	if d.cleanupTicker != nil {
		d.cleanupTicker.Stop()
	}

	// Stop the metrics logging ticker if it exists
	if d.metricsLogTicker != nil {
		d.metricsLogTicker.Stop()
	}

	// Stop the metrics sampling ticker if it exists
	if d.metricsSamplingTicker != nil {
		d.metricsSamplingTicker.Stop()
	}

	// Signal the cleanup goroutine to stop
	close(d.stopChan)

	// Close all streams
	// Use a slice to collect all streams to close to avoid holding locks during close operations
	var streamsToClose []struct {
		streamID string
		stream   api.VPNStream
	}

	// Collect all streams that need to be closed
	d.streams.Range(func(key, value interface{}) bool {
		streamID := key.(string)
		stream := value.(api.VPNStream)
		streamsToClose = append(streamsToClose, struct {
			streamID string
			stream   api.VPNStream
		}{streamID, stream})
		return true
	})

	// Close each stream
	for _, item := range streamsToClose {
		// Extract peer ID from stream ID
		peerInfo := extractPeerID(item.streamID)

		if err := item.stream.Close(); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"streamID": item.streamID,
				"peer":     peerInfo,
			}).Error("Error closing stream")
		}

		// Delete from the maps
		d.streams.Delete(item.streamID)
		d.lastUsed.Delete(item.streamID)
	}

	return nil
}
