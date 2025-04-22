package stream

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// Ensure StreamService implements the PoolServiceExtension interface
var _ pool.PoolServiceExtension = (*StreamService)(nil)

// streamIndexMap tracks streams by index for each peer
type streamIndexMap struct {
	streams map[int]types.VPNStream
	mutex   sync.RWMutex
}

// streamRegistry tracks streams by index for all peers
type streamRegistry struct {
	peerStreams map[string]*streamIndexMap
	mutex       sync.RWMutex
}

// newStreamRegistry creates a new stream registry
func newStreamRegistry() *streamRegistry {
	return &streamRegistry{
		peerStreams: make(map[string]*streamIndexMap),
	}
}

// getOrCreatePeerMap gets or creates a stream map for a peer
func (r *streamRegistry) getOrCreatePeerMap(peerID peer.ID) *streamIndexMap {
	peerIDStr := peerID.String()
	r.mutex.RLock()
	peerMap, exists := r.peerStreams[peerIDStr]
	r.mutex.RUnlock()

	if !exists {
		r.mutex.Lock()
		peerMap = &streamIndexMap{
			streams: make(map[int]types.VPNStream),
		}
		r.peerStreams[peerIDStr] = peerMap
		r.mutex.Unlock()
	}

	return peerMap
}

// registerStream registers a stream for a peer at a specific index
func (r *streamRegistry) registerStream(peerID peer.ID, index int, stream types.VPNStream) {
	peerMap := r.getOrCreatePeerMap(peerID)
	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	// If there's an existing stream at this index, close it
	if existingStream, exists := peerMap.streams[index]; exists && existingStream != stream {
		existingStream.Close()
	}

	peerMap.streams[index] = stream
}

// getStream gets a stream for a peer at a specific index
func (r *streamRegistry) getStream(peerID peer.ID, index int) (types.VPNStream, bool) {
	peerIDStr := peerID.String()
	r.mutex.RLock()
	peerMap, exists := r.peerStreams[peerIDStr]
	r.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	peerMap.mutex.RLock()
	defer peerMap.mutex.RUnlock()

	stream, exists := peerMap.streams[index]
	return stream, exists
}

// removeStream removes a stream for a peer at a specific index
func (r *streamRegistry) removeStream(peerID peer.ID, index int, closeStream bool) bool {
	peerIDStr := peerID.String()
	r.mutex.RLock()
	peerMap, exists := r.peerStreams[peerIDStr]
	r.mutex.RUnlock()

	if !exists {
		return false
	}

	peerMap.mutex.Lock()
	defer peerMap.mutex.Unlock()

	stream, exists := peerMap.streams[index]
	if !exists {
		return false
	}

	if closeStream {
		stream.Close()
	}

	delete(peerMap.streams, index)
	return true
}

// getStreamCount returns the number of streams for a peer
func (r *streamRegistry) getStreamCount(peerID peer.ID) int {
	peerIDStr := peerID.String()
	r.mutex.RLock()
	peerMap, exists := r.peerStreams[peerIDStr]
	r.mutex.RUnlock()

	if !exists {
		return 0
	}

	peerMap.mutex.RLock()
	defer peerMap.mutex.RUnlock()

	return len(peerMap.streams)
}

// getNextAvailableIndex returns the next available index for a peer
func (r *streamRegistry) getNextAvailableIndex(peerID peer.ID) int {
	peerMap := r.getOrCreatePeerMap(peerID)
	peerMap.mutex.RLock()
	defer peerMap.mutex.RUnlock()

	// Find the lowest unused index
	index := 0
	for {
		_, exists := peerMap.streams[index]
		if !exists {
			break
		}
		index++
	}

	return index
}

// Initialize the stream registry in the StreamService
func (s *StreamService) initStreamRegistry() {
	if s.streamRegistry == nil {
		s.streamRegistry = newStreamRegistry()
	}
}

// GetStreamByIndex gets a stream by index from the pool
func (s *StreamService) GetStreamByIndex(ctx context.Context, peerID peer.ID, index int) (types.VPNStream, error) {
	s.mu.Lock()
	s.initStreamRegistry()
	s.mu.Unlock()

	// Add the peer to the stream warmer
	s.streamWarmer.AddPeer(peerID)

	// Check if we already have a stream at this index
	if stream, exists := s.streamRegistry.getStream(peerID, index); exists {
		log.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"index":   index,
		}).Debug("Found existing stream at index")

		s.metrics.IncrementStreamReuse(peerID, index)
		return stream, nil
	}

	// Get a stream from the pool manager
	stream, err := s.poolManager.GetStream(ctx, peerID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"index":   index,
			"error":   err,
		}).Warn("Failed to get stream from pool")

		s.metrics.IncrementStreamAcquisitionFailure(peerID, index)
		return nil, err
	}

	// Register the stream at the specified index
	s.streamRegistry.registerStream(peerID, index, stream)

	log.WithFields(logrus.Fields{
		"peer_id": peerID.String(),
		"index":   index,
	}).Debug("Registered new stream at index")

	s.metrics.IncrementStreamRegistration(peerID, index)
	return stream, nil
}

// ReleaseStreamByIndex releases a stream by index
func (s *StreamService) ReleaseStreamByIndex(peerID peer.ID, index int, close bool) {
	s.mu.Lock()
	s.initStreamRegistry()
	s.mu.Unlock()

	// Remove the stream from the registry
	removed := s.streamRegistry.removeStream(peerID, index, close)

	if removed {
		s.metrics.IncrementStreamRelease(peerID, index)

		log.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"index":   index,
			"close":   close,
		}).Debug("Released stream at index")
	} else {
		log.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"index":   index,
			"close":   close,
		}).Debug("No stream found at index to release")
	}
}

// SetTargetStreamsForPeer sets the target number of streams for a peer
func (s *StreamService) SetTargetStreamsForPeer(peerID peer.ID, targetStreams int) {
	s.mu.Lock()
	s.initStreamRegistry()
	s.mu.Unlock()

	// Get the current stream count for this peer
	currentCount := s.streamRegistry.getStreamCount(peerID)

	log.WithFields(logrus.Fields{
		"peer_id":        peerID.String(),
		"current_count":  currentCount,
		"target_streams": targetStreams,
	}).Debug("Setting target streams for peer")

	// If we need more streams, create them
	if targetStreams > currentCount {
		go s.ensureStreamsForPeer(peerID, targetStreams)
	} else if targetStreams < currentCount {
		// If we have too many streams, we could implement logic to reduce them
		// For now, we'll just log it and let the idle cleanup handle it
		log.WithFields(logrus.Fields{
			"peer_id":        peerID.String(),
			"current_count":  currentCount,
			"target_streams": targetStreams,
		}).Debug("Have more streams than target, will let idle cleanup handle reduction")
	}
}

// ensureStreamsForPeer ensures that we have at least the specified number of streams for a peer
func (s *StreamService) ensureStreamsForPeer(peerID peer.ID, targetStreams int) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.poolManager.GetCleanupInterval())
	defer cancel()

	// Get the current stream count
	currentCount := s.streamRegistry.getStreamCount(peerID)

	// If we already have enough streams, we're done
	if currentCount >= targetStreams {
		log.WithFields(logrus.Fields{
			"peer_id":        peerID.String(),
			"current_count":  currentCount,
			"target_streams": targetStreams,
		}).Debug("Already have enough streams for peer")
		return
	}

	log.WithFields(logrus.Fields{
		"peer_id":        peerID.String(),
		"current_count":  currentCount,
		"target_streams": targetStreams,
		"to_create":      targetStreams - currentCount,
	}).Info("Creating additional streams for peer")

	// Create a wait group to wait for all streams to be created
	var wg sync.WaitGroup
	// Track success/failure counts
	var successCount, failureCount int32
	// Use a mutex to protect the counts
	var countMutex sync.Mutex

	// Create the additional streams
	for i := currentCount; i < targetStreams; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Get the next available index
			index := s.streamRegistry.getNextAvailableIndex(peerID)

			// Create a new stream
			stream, err := s.CreateNewVPNStream(ctx, peerID)
			if err != nil {
				log.WithFields(logrus.Fields{
					"peer_id": peerID.String(),
					"index":   index,
					"error":   err,
				}).Warn("Failed to create stream for peer")

				countMutex.Lock()
				failureCount++
				countMutex.Unlock()
				return
			}

			// Register the stream at the specified index
			s.streamRegistry.registerStream(peerID, index, stream)

			log.WithFields(logrus.Fields{
				"peer_id": peerID.String(),
				"index":   index,
			}).Debug("Created and registered new stream")

			countMutex.Lock()
			successCount++
			countMutex.Unlock()
		}()

		// Add a small delay between stream creation attempts to avoid overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all streams to be created
	wg.Wait()

	// Log the result
	newCount := s.streamRegistry.getStreamCount(peerID)
	log.WithFields(logrus.Fields{
		"peer_id":        peerID.String(),
		"previous_count": currentCount,
		"new_count":      newCount,
		"target_streams": targetStreams,
		"success_count":  successCount,
		"failure_count":  failureCount,
	}).Info("Finished creating streams for peer")
}
