package router

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

var workerLog = logrus.WithField("service", "vpn-packet-worker")

// PacketWorker processes packets
type PacketWorker struct {
	id           int
	packetChan   chan *PacketTask
	ctx          context.Context
	cancel       context.CancelFunc
	router       *StreamRouter
	running      bool
	packetCount  int64
	errorCount   int64
	lastActivity int64 // Unix nano timestamp, accessed atomically
	mu           sync.RWMutex

	// Local cache to reduce global lock acquisitions
	routeCache     map[string]*ConnectionRoute
	routeCacheMu   sync.Mutex
	routeCacheTTL  time.Duration
	lastCacheClean time.Time
}

// Start starts a packet worker
func (w *PacketWorker) Start() {
	w.running = true
	go w.run()
}

// run is the main processing loop for a worker
func (w *PacketWorker) run() {
	workerLog := workerLog.WithField("worker_id", w.id)
	workerLog.Debug("Worker started")

	defer func() {
		w.running = false
		workerLog.Debug("Worker stopped")
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case task, ok := <-w.packetChan:
			if !ok {
				return
			}

			err := w.processPacket(task)

			if task.doneCh != nil {
				task.doneCh <- err
				close(task.doneCh)
			}
		}
	}
}

// processPacket processes a packet using the appropriate stream
func (w *PacketWorker) processPacket(task *PacketTask) error {
	// Initialize route cache if needed
	if w.routeCache == nil {
		w.routeCache = make(map[string]*ConnectionRoute)
		w.routeCacheTTL = 5 * time.Second
		w.lastCacheClean = time.Now()
	}

	// Periodically clean the cache (every 30 seconds)
	if time.Since(w.lastCacheClean) > 30*time.Second {
		w.cleanRouteCache()
	}

	// Check if we have a cached route
	connKey := task.route.connKey
	var cachedRoute *ConnectionRoute

	w.routeCacheMu.Lock()
	cachedRoute = w.routeCache[connKey]
	if cachedRoute != nil {
		// Update last activity time
		atomic.StoreInt64(&cachedRoute.lastActivity, time.Now().UnixNano())
	} else {
		// Cache the route for future use
		w.routeCache[connKey] = task.route
	}
	w.routeCacheMu.Unlock()

	// Get stream for this route
	stream, err := w.router.getStreamForRoute(task.route)
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		return err
	}

	// Write packet to stream
	_, err = stream.Write(task.packet)
	if err != nil {
		// Log the stream error
		workerLog.WithFields(logrus.Fields{
			"worker_id": w.id,
			"peer_id":   task.route.peerID.String(),
			"index":     task.route.streamIndex,
			"error":     err,
		}).Warn("Stream write failed, closing unhealthy stream")

		// Close the unhealthy stream
		w.router.releaseStream(task.route.peerID, task.route.streamIndex)
		atomic.AddInt64(&w.errorCount, 1)

		// Try again with a new stream after a short delay to allow for stream creation
		time.Sleep(10 * time.Millisecond)
		stream, err = w.router.getStreamForRoute(task.route)
		if err != nil {
			workerLog.WithFields(logrus.Fields{
				"worker_id": w.id,
				"peer_id":   task.route.peerID.String(),
				"index":     task.route.streamIndex,
				"error":     err,
			}).Error("Failed to get replacement stream")
			atomic.AddInt64(&w.errorCount, 1)
			return err
		}

		// Try writing with the new stream
		_, err = stream.Write(task.packet)
		if err != nil {
			workerLog.WithFields(logrus.Fields{
				"worker_id": w.id,
				"peer_id":   task.route.peerID.String(),
				"index":     task.route.streamIndex,
				"error":     err,
			}).Error("Failed to write with replacement stream")
			atomic.AddInt64(&w.errorCount, 1)
			return err
		}

		workerLog.WithFields(logrus.Fields{
			"worker_id": w.id,
			"peer_id":   task.route.peerID.String(),
			"index":     task.route.streamIndex,
		}).Debug("Successfully wrote packet with replacement stream")
	}

	// Update metrics using atomic operations
	atomic.AddInt64(&w.packetCount, 1)
	atomic.StoreInt64(&w.lastActivity, time.Now().UnixNano())

	// Update router metrics with packet size
	w.router.recordPacket(task.route, len(task.packet))

	return nil
}

// cleanRouteCache removes expired entries from the route cache
func (w *PacketWorker) cleanRouteCache() {
	w.routeCacheMu.Lock()
	defer w.routeCacheMu.Unlock()

	now := time.Now()
	w.lastCacheClean = now

	// Remove expired entries
	for connKey, route := range w.routeCache {
		lastActivity := time.Unix(0, atomic.LoadInt64(&route.lastActivity))
		if now.Sub(lastActivity) > w.routeCacheTTL {
			delete(w.routeCache, connKey)
		}
	}

	// Log cache size
	workerLog.WithFields(logrus.Fields{
		"worker_id":  w.id,
		"cache_size": len(w.routeCache),
	}).Debug("Cleaned route cache")
}
