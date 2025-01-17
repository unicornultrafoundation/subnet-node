package stats

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl/v2"
	"github.com/shirou/gopsutil/process"
	"github.com/unicornultrafoundation/subnet-node/common/networkutil"
	"github.com/unicornultrafoundation/subnet-node/common/utils"
)

// StatEntry represents the resource usage statistics for a container.
type StatEntry struct {
	UsedUploadBytes   uint64
	UsedDownloadBytes uint64
	UsedGpu           uint64
	UsedCpu           uint64
	UsedMemory        uint64
	UsedStorage       uint64
}

// Stats manages the resource usage statistics for multiple containers.
type Stats struct {
	mu               sync.Mutex
	entries          map[string]*StatEntry
	firstStats       map[string]*StatEntry
	finalStats       map[string]*StatEntry
	containerdClient *containerd.Client
	stopChan         chan struct{}
}

// NewStats creates a new Stats instance.
func NewStats(containerdClient *containerd.Client) *Stats {
	return &Stats{
		entries:          make(map[string]*StatEntry),
		firstStats:       make(map[string]*StatEntry),
		finalStats:       make(map[string]*StatEntry),
		containerdClient: containerdClient,
		stopChan:         make(chan struct{}),
	}
}

// ClearUsageData clears all usage data before the service starts tracking.
func (s *Stats) ClearUsageData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]*StatEntry)
	s.firstStats = make(map[string]*StatEntry)
	s.finalStats = make(map[string]*StatEntry)
}

// UpdateStats updates the stats for a given container ID.
func (s *Stats) updateStats(ctx context.Context, containerId string) error {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")

	// Load the container
	container, err := s.containerdClient.LoadContainer(ctx, containerId)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get the process
	pid := task.Pid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	// Get network I/O counters
	netIO, err := proc.NetIOCounters(false)
	if err != nil {
		return err
	}

	// Detect the network interfaces connected to the internet
	internetInterfaces, err := networkutil.DetectInternetInterfaces()
	if err != nil {
		return fmt.Errorf("failed to detect internet interfaces: %v", err)
	}

	// Calculate total received and transmitted bytes
	var totalRxBytes, totalTxBytes uint64
	for _, stat := range netIO {
		for _, iface := range internetInterfaces {
			if stat.Name == iface {
				totalRxBytes += stat.BytesRecv
				totalTxBytes += stat.BytesSent
			}
		}
	}

	// Get container info
	info, err := container.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container info: %v", err)
	}
	snapshotUsage, err := s.containerdClient.SnapshotService(info.Snapshotter).Usage(ctx, info.SnapshotKey)
	if err != nil {
		return fmt.Errorf("failed to get snapshot usage: %v", err)
	}

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Unmarshal metrics data
	var data interface{}
	switch {
	case typeurl.Is(metric.Data, (*v1.Metrics)(nil)):
		data = &v1.Metrics{}
	case typeurl.Is(metric.Data, (*v2.Metrics)(nil)):
		data = &v2.Metrics{}
	default:
		return fmt.Errorf("cannot convert metric data to cgroups.Metrics")
	}
	if err := typeurl.UnmarshalTo(metric.Data, data); err != nil {
		return err
	}

	// Extract CPU and memory usage
	var usedCpu, usedMemory uint64
	switch metrics := data.(type) {
	case *v1.Metrics:
		usedCpu = metrics.CPU.Usage.Total
		usedMemory = metrics.Memory.Usage.Usage
	case *v2.Metrics:
		usedCpu = metrics.CPU.UsageUsec
		usedMemory = metrics.Memory.Usage
	default:
		return fmt.Errorf("unsupported metrics type")
	}

	// Get GPU usage
	usedGpu, _ := utils.GetGpuUsageByPid(int32(pid))

	// Create current stats entry
	currentStats := &StatEntry{
		UsedUploadBytes:   totalTxBytes,
		UsedDownloadBytes: totalRxBytes,
		UsedGpu:           usedGpu,
		UsedCpu:           usedCpu,
		UsedMemory:        usedMemory,
		UsedStorage:       uint64(snapshotUsage.Size),
	}

	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the initial stats if not already stored
	if _, exists := s.firstStats[containerId]; !exists {
		s.firstStats[containerId] = currentStats
	}

	// Calculate the used stats by subtracting the initial stats
	initialStats := s.firstStats[containerId]
	usedStats := &StatEntry{
		UsedUploadBytes:   currentStats.UsedUploadBytes - initialStats.UsedUploadBytes,
		UsedDownloadBytes: currentStats.UsedDownloadBytes - initialStats.UsedDownloadBytes,
		UsedCpu:           currentStats.UsedCpu - initialStats.UsedCpu,
		UsedGpu:           currentStats.UsedGpu,
		UsedMemory:        currentStats.UsedMemory,
		UsedStorage:       currentStats.UsedStorage,
	}

	// Add final stats if they exist
	if finalStats, exists := s.finalStats[containerId]; exists {
		usedStats.UsedUploadBytes += finalStats.UsedUploadBytes
		usedStats.UsedDownloadBytes += finalStats.UsedDownloadBytes
		usedStats.UsedCpu += finalStats.UsedCpu
	}

	s.entries[containerId] = usedStats

	return nil
}

// GetStats retrieves the stats for a given container ID.
func (s *Stats) GetStats(containerId string) (*StatEntry, error) {
	// Lock the map for reading
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[containerId]
	if !exists {
		return nil, fmt.Errorf("stats not found for container ID: %s", containerId)
	}

	return entry, nil
}

// FinalizeStats finalizes the stats for a given container ID.
func (s *Stats) FinalizeStats(containerId string) error {
	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[containerId]
	if !exists {
		return fmt.Errorf("stats not found for container ID: %s", containerId)
	}

	s.finalStats[containerId] = entry
	delete(s.entries, containerId)
	delete(s.firstStats, containerId)

	return nil
}

// ClaimFinalStats claims the final stats for a given container ID.
func (s *Stats) ClaimFinalStats(containerId string) (*StatEntry, error) {
	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.finalStats[containerId]
	if !exists {
		return nil, fmt.Errorf("final stats not found for container ID: %s", containerId)
	}

	delete(s.finalStats, containerId)

	return entry, nil
}

// ClaimMultipleFinalStats claims the final stats for multiple container IDs.
func (s *Stats) ClaimMultipleFinalStats(containerIds []string) (map[string]*StatEntry, error) {
	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	claimedStats := make(map[string]*StatEntry)
	for _, containerId := range containerIds {
		entry, exists := s.finalStats[containerId]
		if !exists {
			return nil, fmt.Errorf("final stats not found for container ID: %s", containerId)
		}

		claimedStats[containerId] = entry
		delete(s.finalStats, containerId)
	}

	return claimedStats, nil
}

// Start starts the stats service and periodically updates stats for all running containers.
func (s *Stats) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan: // Exit if the stop channel is closed
				return
			case <-ticker.C: // On every tick, update stats for all running containers
				s.updateAllRunningContainersStats()
			}
		}
	}()
}

// Stop stops the stats service.
func (s *Stats) Stop() {
	close(s.stopChan)
}

// updateAllRunningContainersStats updates stats for all running containers.
func (s *Stats) updateAllRunningContainersStats() {
	ctx := namespaces.WithNamespace(context.Background(), "subnet-apps")

	// Fetch all running containers
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		fmt.Printf("failed to fetch running containers: %v\n", err)
		return
	}

	// Iterate over each container and update its stats
	for _, container := range containers {
		containerId := container.ID()
		if err := s.updateStats(ctx, containerId); err != nil {
			fmt.Printf("failed to update stats for container %s: %v\n", containerId, err)
		}
	}
}
