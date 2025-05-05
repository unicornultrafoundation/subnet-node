package dispatcher

import (
	"sort"
	"sync/atomic"
	"time"
)

// GetMetrics returns the current metrics for the dispatcher
func (d *Dispatcher) GetMetrics() map[string]uint64 {
	// Lock both metrics and streams mutex to ensure consistent data
	d.metricsMu.RLock()
	d.mu.RLock()
	defer d.mu.RUnlock()
	defer d.metricsMu.RUnlock()

	metrics := make(map[string]uint64)

	// General metrics
	metrics["active_streams_total"] = uint64(len(d.streams))
	metrics["packets_dispatched_total"] = d.packetsDispatched.Load()
	metrics["bytes_dispatched_total"] = d.bytesDispatched.Load()
	metrics["streams_created_total"] = d.streamsCreated.Load()
	metrics["streams_closed_total"] = d.streamsClosed.Load()
	metrics["stream_errors_total"] = d.streamErrors.Load()
	metrics["peer_discovery_errors_total"] = d.peerDiscoveryErrors.Load()
	metrics["packet_parse_errors_total"] = d.packetParseErrors.Load()

	// Calculate average bytes per packet
	packetsCount := d.packetsDispatched.Load()
	if packetsCount > 0 {
		bytesCount := d.bytesDispatched.Load()
		metrics["avg_bytes_per_packet"] = bytesCount / packetsCount
	}

	// Calculate stream age metrics
	now := time.Now()
	if !d.oldestStreamCreationTime.IsZero() {
		metrics["oldest_stream_age_seconds"] = uint64(now.Sub(d.oldestStreamCreationTime).Seconds())
	}
	if !d.newestStreamCreationTime.IsZero() {
		metrics["newest_stream_age_seconds"] = uint64(now.Sub(d.newestStreamCreationTime).Seconds())
	}

	// Calculate average stream age using sync.Map
	var totalAge uint64
	var streamCount uint64
	d.streamCreationTimes.Range(func(_, value interface{}) bool {
		creationTime := value.(time.Time)
		totalAge += uint64(now.Sub(creationTime).Seconds())
		streamCount++
		return true
	})

	if streamCount > 0 {
		metrics["avg_stream_age_seconds"] = totalAge / streamCount
	}

	// Add per-peer metrics
	var uniquePeers uint64
	d.packetsDispatchedByPeer.Range(func(_, _ interface{}) bool {
		uniquePeers++
		return true
	})
	metrics["unique_peers"] = uniquePeers

	// Add top peers by packet count (up to 5)
	peersByPackets := make([]struct {
		peer    string
		packets uint64
	}, 0, int(uniquePeers))

	// Collect peer packet counts using sync.Map
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

		// Add packet count
		metrics["top_peer_"+shortPeerID+"_packets"] = p.packets

		// Add byte count
		if byteVal, ok := d.bytesDispatchedByPeer.Load(p.peer); ok {
			byteCounter := byteVal.(*atomic.Uint64)
			metrics["top_peer_"+shortPeerID+"_bytes"] = byteCounter.Load()
		}

		// Add active streams count
		if activeVal, ok := d.activeStreamsByPeer.Load(p.peer); ok {
			activeCounter := activeVal.(*atomic.Uint64)
			metrics["top_peer_"+shortPeerID+"_active_streams"] = activeCounter.Load()
		}

		// Add error count
		if errorVal, ok := d.streamErrorsByPeer.Load(p.peer); ok {
			errorCounter := errorVal.(*atomic.Uint64)
			metrics["top_peer_"+shortPeerID+"_errors"] = errorCounter.Load()
		}
	}

	return metrics
}
