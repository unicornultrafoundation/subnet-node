package dispatcher

import (
	"strings"
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
	// Mutex for protecting the streams map
	mu sync.RWMutex
	// Map of peer ID to stream
	streams map[string]api.VPNStream
	// Stream cleanup ticker
	cleanupTicker *time.Ticker
	// Metrics logging ticker
	metricsLogTicker *time.Ticker
	// Stop channel for background goroutines
	stopChan chan struct{}
	// Last used time for each stream
	lastUsed map[string]time.Time

	// Metrics
	metricsMu           sync.RWMutex
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
		streams:       make(map[string]api.VPNStream),
		stopChan:      make(chan struct{}),
		lastUsed:      make(map[string]time.Time),

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

	// Start the metrics logging ticker
	metricsInterval := d.configService.GetMetricsLogInterval()
	d.metricsLogTicker = time.NewTicker(metricsInterval)

	// Start the metrics logging goroutine
	go d.logMetricsPeriodically()

	log.Infof("Metrics logging scheduled every %s", metricsInterval)
}

// Stop stops the dispatcher
func (d *Dispatcher) Stop() {
	log.Info("Stopping VPN dispatcher")

	// Stop the cleanup ticker if it exists
	if d.cleanupTicker != nil {
		d.cleanupTicker.Stop()
	}

	// Stop the metrics logging ticker if it exists
	if d.metricsLogTicker != nil {
		d.metricsLogTicker.Stop()
	}

	// Signal the cleanup goroutine to stop
	close(d.stopChan)

	// Close all streams
	d.mu.Lock()
	defer d.mu.Unlock()

	for streamID, stream := range d.streams {
		// Extract peer ID from stream ID (format: peerID/queueID)
		parts := strings.Split(streamID, "/")
		peerInfo := streamID
		if len(parts) > 0 {
			peerInfo = parts[0]
		}

		if err := stream.Close(); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"streamID": streamID,
				"peer":     peerInfo,
			}).Error("Error closing stream")
		}
	}

	// Clear the streams map
	d.streams = make(map[string]api.VPNStream)
	d.lastUsed = make(map[string]time.Time)
}
