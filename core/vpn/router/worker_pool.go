package router

import (
	"context"
	"hash/fnv"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool manages a pool of packet processing workers
type WorkerPool struct {
	workers     []*PacketWorker
	workerCount int32
	minWorkers  int
	maxWorkers  int
	workerMu    sync.RWMutex
	router      *StreamRouter
	queueSize   int

	// Worker metrics
	workerMetrics map[int]*WorkerMetrics
	metricsMu     sync.RWMutex

	// Scaling parameters
	scaleInterval      time.Duration
	scaleUpThreshold   float64 // e.g., 0.75 (75% queue utilization)
	scaleDownThreshold float64 // e.g., 0.25 (25% queue utilization)
}

// WorkerMetrics tracks performance metrics for a worker
type WorkerMetrics struct {
	queueSize        int
	processedPackets int64
	errors           int64
	lastUpdateTime   time.Time
}

// WorkerPoolConfig contains configuration for the worker pool
type WorkerPoolConfig struct {
	MinWorkers         int
	MaxWorkers         int
	InitialWorkers     int
	QueueSize          int
	ScaleInterval      time.Duration
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
}

// NewWorkerPool creates a new worker pool with dynamic scaling
func NewWorkerPool(config *WorkerPoolConfig) *WorkerPool {
	return &WorkerPool{
		workers:            make([]*PacketWorker, 0, config.MaxWorkers),
		workerCount:        0,
		minWorkers:         config.MinWorkers,
		maxWorkers:         config.MaxWorkers,
		queueSize:          config.QueueSize,
		workerMetrics:      make(map[int]*WorkerMetrics),
		scaleInterval:      config.ScaleInterval,
		scaleUpThreshold:   config.ScaleUpThreshold,
		scaleDownThreshold: config.ScaleDownThreshold,
	}
}

// Initialize initializes the worker pool with a reference to the router
func (p *WorkerPool) Initialize(router *StreamRouter) {
	p.router = router

	// Create initial workers
	initialWorkers := max(p.minWorkers, min(p.router.config.InitialWorkers, p.maxWorkers))

	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	for i := 0; i < initialWorkers; i++ {
		p.createWorker()
	}

	// Start scaling routine
	go p.startScalingRoutine()
}

// createWorker creates a new worker and adds it to the pool
func (p *WorkerPool) createWorker() *PacketWorker {
	workerID := int(p.workerCount)
	ctx, cancel := context.WithCancel(p.router.ctx)

	worker := &PacketWorker{
		id:             workerID,
		packetChan:     make(chan *PacketTask, p.queueSize),
		ctx:            ctx,
		cancel:         cancel,
		router:         p.router,
		lastActivity:   time.Now().UnixNano(),
		routeCache:     make(map[string]*ConnectionRoute),
		routeCacheTTL:  5 * time.Second,
		lastCacheClean: time.Now(),
	}

	p.workers = append(p.workers, worker)
	atomic.AddInt32(&p.workerCount, 1)

	// Initialize metrics
	p.metricsMu.Lock()
	p.workerMetrics[workerID] = &WorkerMetrics{
		lastUpdateTime: time.Now(),
	}
	p.metricsMu.Unlock()

	// Start the worker
	worker.Start()

	log.Infof("Created worker %d, total workers: %d", workerID, p.workerCount)
	return worker
}

// GetWorkerCount returns the current number of workers
func (p *WorkerPool) GetWorkerCount() int {
	return int(atomic.LoadInt32(&p.workerCount))
}

// GetWorker returns a worker by ID
func (p *WorkerPool) GetWorker(id int) *PacketWorker {
	p.workerMu.RLock()
	defer p.workerMu.RUnlock()

	count := len(p.workers)
	if count == 0 {
		return nil
	}

	// Ensure ID is within bounds
	safeID := id % count
	return p.workers[safeID]
}

// startScalingRoutine starts the worker scaling routine
func (p *WorkerPool) startScalingRoutine() {
	ticker := time.NewTicker(p.scaleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.router.ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
			p.evaluateScaling()
		}
	}
}

// updateMetrics updates worker metrics
func (p *WorkerPool) updateMetrics() {
	p.workerMu.RLock()
	defer p.workerMu.RUnlock()

	p.metricsMu.Lock()
	defer p.metricsMu.Unlock()

	for _, worker := range p.workers {
		if !worker.running {
			continue
		}

		metrics := p.workerMetrics[worker.id]
		if metrics == nil {
			metrics = &WorkerMetrics{
				lastUpdateTime: time.Now(),
			}
			p.workerMetrics[worker.id] = metrics
		}

		// Update queue size
		metrics.queueSize = len(worker.packetChan)

		// Update processed packets (from worker's internal counter)
		metrics.processedPackets = atomic.LoadInt64(&worker.packetCount)

		// Update errors
		metrics.errors = atomic.LoadInt64(&worker.errorCount)
	}
}

// evaluateScaling evaluates whether to scale workers up or down
func (p *WorkerPool) evaluateScaling() {
	currentWorkers := p.GetWorkerCount()

	// Calculate average queue utilization
	p.metricsMu.RLock()
	totalQueueSize := 0
	activeWorkers := 0

	for _, metrics := range p.workerMetrics {
		totalQueueSize += metrics.queueSize
		activeWorkers++
	}

	avgQueueUtilization := 0.0
	if activeWorkers > 0 {
		avgQueueUtilization = float64(totalQueueSize) / float64(activeWorkers*p.queueSize)
	}
	p.metricsMu.RUnlock()

	// Scale up if average queue utilization is high
	if avgQueueUtilization > p.scaleUpThreshold && currentWorkers < p.maxWorkers {
		p.scaleUp()
		log.Infof("Scaling up workers from %d to %d (utilization: %.2f)",
			currentWorkers, currentWorkers+1, avgQueueUtilization)
	}

	// Scale down if average queue utilization is low and we have more than minimum workers
	if avgQueueUtilization < p.scaleDownThreshold && currentWorkers > p.minWorkers {
		p.scaleDown()
		log.Infof("Scaling down workers from %d to %d (utilization: %.2f)",
			currentWorkers, currentWorkers-1, avgQueueUtilization)
	}
}

// scaleUp adds a new worker to the pool
func (p *WorkerPool) scaleUp() {
	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	if int(p.workerCount) >= p.maxWorkers {
		return
	}

	p.createWorker()
}

// scaleDown removes a worker from the pool
func (p *WorkerPool) scaleDown() {
	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	if int(p.workerCount) <= p.minWorkers {
		return
	}

	// Find the least busy worker
	leastBusyWorkerIdx := -1
	minQueueSize := math.MaxInt32

	p.metricsMu.RLock()
	for i, worker := range p.workers {
		if !worker.running {
			continue
		}

		metrics := p.workerMetrics[worker.id]
		if metrics != nil && metrics.queueSize < minQueueSize {
			minQueueSize = metrics.queueSize
			leastBusyWorkerIdx = i
		}
	}
	p.metricsMu.RUnlock()

	if leastBusyWorkerIdx == -1 {
		return
	}

	// Stop the worker
	worker := p.workers[leastBusyWorkerIdx]
	p.stopWorker(worker)

	// Remove from slice
	p.workers = append(p.workers[:leastBusyWorkerIdx], p.workers[leastBusyWorkerIdx+1:]...)
	atomic.AddInt32(&p.workerCount, -1)

	log.Infof("Removed worker %d, total workers: %d", worker.id, p.workerCount)
}

// stopWorker stops a worker and redistributes its connections
func (p *WorkerPool) stopWorker(worker *PacketWorker) {
	// Signal worker to stop
	worker.cancel()

	// Wait for worker to finish current tasks
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	// Wait for worker to empty its queue or timeout
	for {
		if len(worker.packetChan) == 0 || !worker.running {
			break
		}

		select {
		case <-timeout.C:
			// Force close if timeout
			close(worker.packetChan)
			return
		case <-time.After(100 * time.Millisecond):
			// Check again
			continue
		}
	}

	// Redistribute connections from this worker
	p.redistributeConnections(worker.id)

	// Clean up metrics
	p.metricsMu.Lock()
	delete(p.workerMetrics, worker.id)
	p.metricsMu.Unlock()
}

// redistributeConnections redistributes connections from a stopped worker
func (p *WorkerPool) redistributeConnections(workerID int) {
	// Get all connections assigned to this worker
	p.router.connectionMapMu.RLock()
	affectedConnections := make([]string, 0)

	for connKey, assignedWorkerID := range p.router.connectionToWorker {
		if assignedWorkerID == workerID {
			affectedConnections = append(affectedConnections, connKey)
		}
	}
	p.router.connectionMapMu.RUnlock()

	if len(affectedConnections) == 0 {
		return
	}

	log.Infof("Redistributing %d connections from worker %d", len(affectedConnections), workerID)

	// Get active worker IDs
	p.workerMu.RLock()
	activeWorkerIDs := make([]int, 0, len(p.workers))
	for _, w := range p.workers {
		if w.id != workerID && w.running {
			activeWorkerIDs = append(activeWorkerIDs, w.id)
		}
	}
	p.workerMu.RUnlock()

	if len(activeWorkerIDs) == 0 {
		log.Warn("No active workers to redistribute connections to")
		return
	}

	// Redistribute connections
	p.router.connectionMapMu.Lock()
	for _, connKey := range affectedConnections {
		// Use consistent hashing to assign to a new worker
		h := fnv.New32a()
		h.Write([]byte(connKey))
		newWorkerIdx := int(h.Sum32() % uint32(len(activeWorkerIDs)))
		newWorkerID := activeWorkerIDs[newWorkerIdx]

		// Update the mapping
		p.router.connectionToWorker[connKey] = newWorkerID

		// Update the route in cache if it exists
		if routeObj, found := p.router.connectionCache.Get(connKey); found {
			route := routeObj.(*ConnectionRoute)
			route.workerID = newWorkerID
		}
	}
	p.router.connectionMapMu.Unlock()
}

// Shutdown gracefully shuts down the worker pool
func (p *WorkerPool) Shutdown() {
	log.Info("Shutting down WorkerPool")

	p.workerMu.Lock()
	workers := p.workers
	p.workerMu.Unlock()

	// Stop all workers
	for _, worker := range workers {
		worker.cancel()
	}

	// Wait for workers to finish (with timeout)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		allStopped := true

		p.workerMu.RLock()
		for _, worker := range p.workers {
			if worker.running {
				allStopped = false
				break
			}
		}
		p.workerMu.RUnlock()

		if allStopped {
			break
		}

		select {
		case <-timeout:
			log.Warn("Timeout waiting for workers to stop")
			return
		case <-ticker.C:
			continue
		}
	}

	log.Info("WorkerPool shutdown complete")
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
