package router

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

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
	lastActivity time.Time
	mu           sync.RWMutex
}

// Start starts a packet worker
func (w *PacketWorker) Start() {
	w.running = true
	go w.run()
}

// run is the main processing loop for a worker
func (w *PacketWorker) run() {
	workerLog := log.WithField("worker_id", w.id)
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
	// Get stream for this route
	stream, err := w.router.getStreamForRoute(task.route)
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		return err
	}
	
	// Write packet to stream
	_, err = stream.Write(task.packet)
	if err != nil {
		// Handle stream error
		w.router.releaseStream(task.route.peerID, task.route.streamIndex)
		atomic.AddInt64(&w.errorCount, 1)
		
		// Try again with a new stream
		stream, err = w.router.getStreamForRoute(task.route)
		if err != nil {
			atomic.AddInt64(&w.errorCount, 1)
			return err
		}
		
		_, err = stream.Write(task.packet)
		if err != nil {
			atomic.AddInt64(&w.errorCount, 1)
			return err
		}
	}
	
	// Update metrics
	atomic.AddInt64(&w.packetCount, 1)
	w.mu.Lock()
	w.lastActivity = time.Now()
	w.mu.Unlock()
	
	// Update router metrics
	w.router.recordPacket(task.route)
	
	return nil
}
