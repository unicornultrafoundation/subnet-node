package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	ctypes "github.com/docker/docker/api/types/container"
	mtypes "github.com/docker/docker/api/types/mount"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
)

var log = logrus.WithField("service", "apps")

// StatEntry represents the resource usage statistics for a container.
type StatEntry struct {
	UsedUploadBytes   uint64
	UsedDownloadBytes uint64
	UsedGpu           uint64
	UsedCpu           uint64
	UsedMemory        uint64
	UsedStorage       uint64
	Duration          int64
}

// Stats manages the resource usage statistics for multiple containers.
type Stats struct {
	mu                 sync.Mutex
	entries            map[string]*StatEntry
	firstStats         map[string]*StatEntry
	finalStats         map[string]*StatEntry
	dockerClient       docker.DockerClient
	stopChan           chan struct{}
	gpu                *GpuMonitor
	containerToPid     map[string]int32
	startTimes         map[string]time.Time
	memoryUsage        map[int32]uint64
	memorySampleCount  map[int32]uint64
	volumeSizeCache    map[string]int64
	volumeSizeCacheMux sync.RWMutex
}

// NewStats creates a new Stats instance.
func NewStats(ctx context.Context, dockerClient docker.DockerClient) *Stats {

	stats := &Stats{
		entries:           make(map[string]*StatEntry),
		firstStats:        make(map[string]*StatEntry),
		finalStats:        make(map[string]*StatEntry),
		dockerClient:      dockerClient,
		stopChan:          make(chan struct{}),
		gpu:               NewGpuMonitor(5 * time.Second),
		containerToPid:    make(map[string]int32),
		startTimes:        make(map[string]time.Time),
		memoryUsage:       make(map[int32]uint64),
		memorySampleCount: make(map[int32]uint64),
		volumeSizeCache:   make(map[string]int64),
	}

	stats.startVolumeSizeCacheJob(ctx)

	return stats
}

// ClearUsageData clears all usage data before the service starts tracking.
func (s *Stats) ClearUsageData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]*StatEntry)
	s.firstStats = make(map[string]*StatEntry)
	s.finalStats = make(map[string]*StatEntry)
	s.containerToPid = make(map[string]int32)
	s.startTimes = make(map[string]time.Time)
	s.memoryUsage = make(map[int32]uint64)
	s.memorySampleCount = make(map[int32]uint64)

	s.gpu.ClearAllGpuUsage()
}

// UpdateStats updates the stats for a given container ID.
func (s *Stats) updateStats(ctx context.Context, containerId string) error {
	// Load the container
	container, err := s.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	// Get the process
	pid := container.State.Pid

	// Check if the process ID has changed
	s.mu.Lock()
	oldPid, exists := s.containerToPid[containerId]
	if exists && oldPid != int32(pid) {
		// Finalize stats for the old process ID
		if err := s.FinalizeStats(containerId); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("failed to finalize stats for container %s: %v", containerId, err)
		}
	}
	s.containerToPid[containerId] = int32(pid)
	s.mu.Unlock()

	// Fetch real-time container stats
	stats, err := s.dockerClient.ContainerStats(context.Background(), containerId, false)
	if err != nil {
		return fmt.Errorf("failed to load container stats: %w", err)
	}
	defer stats.Body.Close()

	// Parse JSON response
	var containerStats ctypes.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
		return fmt.Errorf("failed to decode container: %w", err)
	}

	// Calculate CPU usage
	usedCpu := containerStats.CPUStats.CPUUsage.TotalUsage
	// Get memory usage in bytes
	usedMemory := containerStats.MemoryStats.Usage

	// Network IO in bytes
	var totalRxBytes, totalTxBytes uint64
	for _, v := range containerStats.Networks {
		totalRxBytes += v.RxBytes // Bytes received
		totalTxBytes += v.TxBytes // Bytes sent
	}

	// Get GPU usage
	usedGpu, _ := s.gpu.GetAverageGpuUsageByPid(int32(pid))

	// Store memory usage data
	s.mu.Lock()
	s.memoryUsage[int32(pid)] += usedMemory
	s.memorySampleCount[int32(pid)]++
	s.mu.Unlock()

	// get mount storage
	mountStorage, err := s.getTotalContainerMountVolume(ctx, containerId)
	if err != nil {
		log.Errorf("failed to get total container volume size: %v", err)
	}

	// Get Storage usage
	usedStorage := int64(0)
	if container.SizeRw != nil {
		usedStorage = *container.SizeRw
	}
	// Create current stats entry
	currentStats := &StatEntry{
		UsedUploadBytes:   totalTxBytes,
		UsedDownloadBytes: totalRxBytes,
		UsedGpu:           usedGpu,
		UsedCpu:           usedCpu,
		UsedMemory:        usedMemory,
		UsedStorage:       uint64(usedStorage) + uint64(mountStorage),
	}

	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the initial stats if not already stored
	if _, exists := s.firstStats[containerId]; !exists {
		s.firstStats[containerId] = currentStats
		s.startTimes[containerId] = time.Now()

	}
	// Calculate the used stats by subtracting the initial stats
	initialStats := s.firstStats[containerId]

	usedStats := &StatEntry{
		UsedUploadBytes:   currentStats.UsedUploadBytes - initialStats.UsedUploadBytes,
		UsedDownloadBytes: currentStats.UsedDownloadBytes - initialStats.UsedDownloadBytes,
		UsedCpu:           currentStats.UsedCpu - initialStats.UsedCpu,
		UsedGpu:           currentStats.UsedGpu,
		UsedMemory:        s.memoryUsage[int32(pid)] / s.memorySampleCount[int32(pid)],
		UsedStorage:       currentStats.UsedStorage,
		Duration:          int64(time.Since(s.startTimes[containerId]).Seconds()),
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
	delete(s.startTimes, containerId)

	// Clear GPU and memory usage for the corresponding process
	if pid, exists := s.containerToPid[containerId]; exists {
		s.gpu.ClearGpuUsageByPid(pid)
		delete(s.memoryUsage, pid)
		delete(s.memorySampleCount, pid)
		delete(s.containerToPid, containerId)
	}

	return nil
}

// ClearFinalStats clears the final stats for a given container ID.
func (s *Stats) ClearFinalStats(containerId string) error {
	// Lock the map for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.finalStats[containerId]
	if !exists {
		return fmt.Errorf("final stats not found for container ID: %s", containerId)
	}

	// Clear GPU and memory usage for the corresponding process
	if pid, exists := s.containerToPid[containerId]; exists {
		s.gpu.ClearGpuUsageByPid(pid)
		delete(s.memoryUsage, pid)
		delete(s.memorySampleCount, pid)
		delete(s.containerToPid, containerId)
	}

	delete(s.finalStats, containerId)

	return nil
}

// GetFinalStats retrieves the final stats for a given container ID.
func (s *Stats) GetFinalStats(containerId string) (*StatEntry, error) {
	// Lock the map for reading
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.finalStats[containerId]
	if !exists {
		return nil, fmt.Errorf("final stats not found for container ID: %s", containerId)
	}

	return entry, nil
}

// GetAllFinalStats retrieves the final stats for all containers.
func (s *Stats) GetAllFinalStats() (map[string]*StatEntry, error) {
	// Lock the map for reading
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.finalStats) == 0 {
		return nil, fmt.Errorf("no final stats available")
	}

	return s.finalStats, nil
}

// Start starts the stats service and periodically updates stats for all running containers.
func (s *Stats) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
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
	ctx := context.Background()
	// Fetch all running containers
	containers, err := s.dockerClient.ContainerList(ctx, ctypes.ListOptions{})
	if err != nil {
		log.Errorf("failed to fetch running containers: %v\n", err)
		return
	}

	// Iterate over each container and update its stats
	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := strings.TrimPrefix(container.Names[0], "/")
		if err := s.updateStats(ctx, containerId); err != nil {
			log.Errorf("failed to update stats for container %s: %v\n", containerId, err)
		}
	}
}

func (s *Stats) getTotalContainerMountVolume(ctx context.Context, containerId string) (int64, error) {
	// Inspect the container to get mount information
	containerInfo, err := s.dockerClient.ContainerInspect(ctx, containerId)

	if err != nil {
		return 0, fmt.Errorf("failed to inspect container: %w", err)
	}

	var totalSize int64

	for _, mount := range containerInfo.Mounts {
		if mount.Type == mtypes.TypeVolume {
			// Try to get size from cache
			if size, exists := s.getVolumeSizeFromCache(mount.Name); exists {
				totalSize += size
				continue
			}

			// If not in cache, trigger a cache update
			if err := s.updateVolumeSizeCache(ctx); err != nil {
				log.Warnf("Failed to update volume size cache: %v", err)
				continue
			}

			// Try again from cache
			if size, exists := s.getVolumeSizeFromCache(mount.Name); exists {
				totalSize += size
			}
		}
	}

	return totalSize, nil
}

// CACHE VOLUME SIZE

// updateVolumeSizeCache updates the cache with current volume sizes
func (s *Stats) updateVolumeSizeCache(ctx context.Context) error {
	dfStats, err := s.dockerClient.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	// Create new map of volume sizes
	newCache := make(map[string]int64)
	for _, v := range dfStats.Volumes {
		if v.UsageData != nil {
			newCache[v.Name] = v.UsageData.Size
		}
	}

	// Update cache thread-safely
	s.volumeSizeCacheMux.Lock()
	s.volumeSizeCache = newCache
	s.volumeSizeCacheMux.Unlock()

	return nil
}

// getVolumeSizeFromCache gets a volume size from cache
func (s *Stats) getVolumeSizeFromCache(volumeName string) (int64, bool) {
	s.volumeSizeCacheMux.RLock()
	defer s.volumeSizeCacheMux.RUnlock()
	size, exists := s.volumeSizeCache[volumeName]
	return size, exists
}

// startVolumeSizeCacheJob starts the periodic cache update job
func (s *Stats) startVolumeSizeCacheJob(ctx context.Context) {
	cacheTicker := time.NewTicker(15 * time.Minute)
	defer cacheTicker.Stop()
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-cacheTicker.C:
				if err := s.updateVolumeSizeCache(ctx); err != nil {
					log.Errorf("Failed to update volume size cache: %v", err)
				}
			}
		}
	}()
}
