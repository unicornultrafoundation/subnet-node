package dispatcher

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// SampledMetrics holds a snapshot of metrics that is periodically updated
type SampledMetrics struct {
	// General metrics
	ActiveStreamsTotal       uint64
	PacketsDispatchedTotal   uint64
	BytesDispatchedTotal     uint64
	StreamsCreatedTotal      uint64
	StreamsClosedTotal       uint64
	StreamErrorsTotal        uint64
	PeerDiscoveryErrorsTotal uint64
	PacketParseErrorsTotal   uint64
	AvgBytesPerPacket        uint64

	// Stream age metrics
	OldestStreamAgeSeconds uint64
	NewestStreamAgeSeconds uint64
	AvgStreamAgeSeconds    uint64

	// Peer metrics
	UniquePeers uint64
	TopPeers    []PeerMetrics

	// Last update time
	LastUpdated time.Time
}

// PeerMetrics holds metrics for a specific peer
type PeerMetrics struct {
	PeerID        string
	ShortPeerID   string
	Packets       uint64
	Bytes         uint64
	ActiveStreams uint64
	Errors        uint64
}

// startMetricsSampling starts a goroutine that periodically samples metrics
func (d *Dispatcher) startMetricsSampling() {
	// Create initial metrics
	d.sampledMetrics = &SampledMetrics{
		LastUpdated: time.Now(),
		TopPeers:    make([]PeerMetrics, 0, 5),
	}

	// Start sampling goroutine
	samplingInterval := d.configService.GetMetricsSamplingInterval()
	if samplingInterval == 0 {
		// Default to 1 second if not configured
		samplingInterval = 1 * time.Second
	}

	d.metricsSamplingTicker = time.NewTicker(samplingInterval)

	go func() {
		// Add panic recovery
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic in metrics sampling goroutine: %v", r)
				// Restart the goroutine after a short delay
				time.Sleep(time.Second)
				d.startMetricsSampling()
			}
		}()

		for {
			select {
			case <-d.stopChan:
				return
			case <-d.metricsSamplingTicker.C:
				// Wrap the sampling in another recovery block
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Errorf("Recovered from panic in sampleMetrics: %v", r)
						}
					}()
					d.sampleMetrics()
				}()
			}
		}
	}()

	log.Infof("Metrics sampling scheduled every %s", samplingInterval)
}

// sampleMetrics collects current metrics and updates the sampled metrics
func (d *Dispatcher) sampleMetrics() {
	now := time.Now()

	// Create a new metrics snapshot
	metrics := &SampledMetrics{
		LastUpdated: now,
		TopPeers:    make([]PeerMetrics, 0, 5),
	}

	// Sample general metrics using atomic operations (no locks needed)
	metrics.PacketsDispatchedTotal = d.packetsDispatched.Load()
	metrics.BytesDispatchedTotal = d.bytesDispatched.Load()
	metrics.StreamsCreatedTotal = d.streamsCreated.Load()
	metrics.StreamsClosedTotal = d.streamsClosed.Load()
	metrics.StreamErrorsTotal = d.streamErrors.Load()
	metrics.PeerDiscoveryErrorsTotal = d.peerDiscoveryErrors.Load()
	metrics.PacketParseErrorsTotal = d.packetParseErrors.Load()

	// Calculate average bytes per packet
	packetsCount := metrics.PacketsDispatchedTotal
	if packetsCount > 0 {
		bytesCount := metrics.BytesDispatchedTotal
		metrics.AvgBytesPerPacket = bytesCount / packetsCount
	}

	// Count active streams using sync.Map
	var activeStreamsCount uint64
	d.streams.Range(func(_, _ interface{}) bool {
		activeStreamsCount++
		return true
	})
	metrics.ActiveStreamsTotal = activeStreamsCount

	// Sample stream age metrics with proper locking
	func() {
		streamTimeMutex.Lock()
		defer streamTimeMutex.Unlock()

		if !d.oldestStreamCreationTime.IsZero() {
			metrics.OldestStreamAgeSeconds = uint64(now.Sub(d.oldestStreamCreationTime).Seconds())
		}
		if !d.newestStreamCreationTime.IsZero() {
			metrics.NewestStreamAgeSeconds = uint64(now.Sub(d.newestStreamCreationTime).Seconds())
		}
	}()

	// Calculate average stream age using sync.Map (no locks needed)
	var totalAge uint64
	var streamCount uint64
	d.streamCreationTimes.Range(func(_, value interface{}) bool {
		creationTime := value.(time.Time)
		totalAge += uint64(now.Sub(creationTime).Seconds())
		streamCount++
		return true
	})

	if streamCount > 0 {
		metrics.AvgStreamAgeSeconds = totalAge / streamCount
	}

	// Count unique peers (no locks needed)
	var uniquePeers uint64
	d.packetsDispatchedByPeer.Range(func(_, _ interface{}) bool {
		uniquePeers++
		return true
	})
	metrics.UniquePeers = uniquePeers

	// Collect top peers by packet count (no locks needed)
	peersByPackets := make([]struct {
		peer    string
		packets uint64
	}, 0, int(uniquePeers))

	d.packetsDispatchedByPeer.Range(func(key, value interface{}) bool {
		peer := key.(string)
		counter := value.(*atomic.Uint64)
		packets := counter.Load()

		peersByPackets = append(peersByPackets, struct {
			peer    string
			packets uint64
		}{peer, packets})

		return true
	})

	// Sort peers by packet count (descending)
	sort.Slice(peersByPackets, func(i, j int) bool {
		return peersByPackets[i].packets > peersByPackets[j].packets
	})

	// Add top 5 peers to metrics
	for i, p := range peersByPackets {
		if i >= 5 {
			break
		}

		shortPeerID := p.peer
		if len(shortPeerID) > 8 {
			shortPeerID = shortPeerID[:8]
		}

		peerMetrics := PeerMetrics{
			PeerID:      p.peer,
			ShortPeerID: shortPeerID,
			Packets:     p.packets,
		}

		// Add byte count
		if byteVal, ok := d.bytesDispatchedByPeer.Load(p.peer); ok {
			byteCounter := byteVal.(*atomic.Uint64)
			peerMetrics.Bytes = byteCounter.Load()
		}

		// Add active streams count
		if activeVal, ok := d.activeStreamsByPeer.Load(p.peer); ok {
			activeCounter := activeVal.(*atomic.Uint64)
			peerMetrics.ActiveStreams = activeCounter.Load()
		}

		// Add error count
		if errorVal, ok := d.streamErrorsByPeer.Load(p.peer); ok {
			errorCounter := errorVal.(*atomic.Uint64)
			peerMetrics.Errors = errorCounter.Load()
		}

		metrics.TopPeers = append(metrics.TopPeers, peerMetrics)
	}

	// Update the sampled metrics atomically
	d.sampledMetricsMu.Lock()
	d.sampledMetrics = metrics
	d.sampledMetricsMu.Unlock()
}

// GetSampledMetrics returns the current sampled metrics
func (d *Dispatcher) GetSampledMetrics() *SampledMetrics {
	d.sampledMetricsMu.RLock()
	defer d.sampledMetricsMu.RUnlock()

	// Check if metrics are initialized
	if d.sampledMetrics == nil {
		// Return an empty metrics object with current timestamp
		return &SampledMetrics{
			LastUpdated: time.Now(),
			TopPeers:    make([]PeerMetrics, 0),
		}
	}

	// Return a copy to prevent race conditions
	metrics := *d.sampledMetrics
	metrics.TopPeers = make([]PeerMetrics, len(d.sampledMetrics.TopPeers))
	copy(metrics.TopPeers, d.sampledMetrics.TopPeers)

	return &metrics
}

// logMetricsPeriodically periodically logs the dispatcher metrics
func (d *Dispatcher) logMetricsPeriodically() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in metrics logging goroutine: %v", r)
			// Restart the goroutine after a short delay
			time.Sleep(time.Second)
			go d.logMetricsPeriodically()
		}
	}()

	for {
		select {
		case <-d.stopChan:
			return
		case <-d.metricsLogTicker.C:
			// Wrap the logging in another recovery block
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("Recovered from panic in metrics logging: %v", r)
					}
				}()

				// Use sampled metrics instead of calculating on the fly
				metrics := d.GetSampledMetrics()
				if metrics == nil {
					log.Warn("Sampled metrics is nil, skipping metrics logging")
					return
				}

				// Log the most important metrics
				log.WithFields(logrus.Fields{
					"active_streams":       metrics.ActiveStreamsTotal,
					"packets_dispatched":   metrics.PacketsDispatchedTotal,
					"bytes_dispatched":     metrics.BytesDispatchedTotal,
					"streams_created":      metrics.StreamsCreatedTotal,
					"streams_closed":       metrics.StreamsClosedTotal,
					"stream_errors":        metrics.StreamErrorsTotal,
					"unique_peers":         metrics.UniquePeers,
					"avg_bytes_per_packet": metrics.AvgBytesPerPacket,
				}).Info("VPN dispatcher metrics")

				// Log per-peer metrics if there are any peers
				if metrics.UniquePeers > 0 && len(metrics.TopPeers) > 0 {
					for _, peer := range metrics.TopPeers {
						fields := logrus.Fields{
							"peer_id":        peer.ShortPeerID,
							"packets":        peer.Packets,
							"bytes":          peer.Bytes,
							"active_streams": peer.ActiveStreams,
							"errors":         peer.Errors,
						}

						log.WithFields(fields).Info("VPN peer metrics")
					}
				}
			}()
		}
	}
}
