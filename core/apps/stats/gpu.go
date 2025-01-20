package stats

import (
	"fmt"
	"sync"
	"time"

	"github.com/unicornultrafoundation/subnet-node/common/utils"
)

// GpuMonitor struct to hold monitoring configuration and data
type GpuMonitor struct {
	Interval    time.Duration
	UsageData   map[int32]uint64
	SampleCount map[int32]uint64
	DataMutex   sync.Mutex
	stopChan    chan struct{}
	stopWait    sync.WaitGroup
}

// NewGpuMonitor creates a new GpuMonitor with the given interval
func NewGpuMonitor(interval time.Duration) *GpuMonitor {
	return &GpuMonitor{
		Interval:    interval,
		UsageData:   make(map[int32]uint64),
		SampleCount: make(map[int32]uint64),
		stopChan:    make(chan struct{}),
	}
}

// Start begins monitoring GPU usage at regular intervals.
func (gm *GpuMonitor) Start() {
	gm.stopWait.Add(1)
	go gm.monitor()
}

// Stop stops the monitoring process.
func (gm *GpuMonitor) Stop() {
	close(gm.stopChan)
	gm.stopWait.Wait()
}

// monitor is the internal method that performs the monitoring.
func (gm *GpuMonitor) monitor() {
	defer gm.stopWait.Done()

	for {
		select {
		case <-gm.stopChan:
			return
		default:
			// Get GPU usage for all processes
			gpuUsage, err := utils.GetGpuUsage()
			if err != nil {
				fmt.Printf("failed to get GPU usage: %v\n", err)
				time.Sleep(gm.Interval)
				continue
			}

			// Store the GPU usage data
			gm.DataMutex.Lock()
			for pid, usage := range gpuUsage {
				gm.UsageData[pid] += usage
				gm.SampleCount[pid]++
			}
			gm.DataMutex.Unlock()

			// Print GPU usage for each process
			for pid, usage := range gpuUsage {
				fmt.Printf("PID: %d, GPU Memory Used: %d MiB\n", pid, usage)
			}

			time.Sleep(gm.Interval)
		}
	}
}

// GetAverageGpuUsage calculates the average GPU memory usage for each process over the last 30 minutes.
func (gm *GpuMonitor) GetAverageGpuUsage() map[int32]uint64 {
	gm.DataMutex.Lock()
	defer gm.DataMutex.Unlock()

	averageUsage := make(map[int32]uint64)
	for pid, totalUsage := range gm.UsageData {
		averageUsage[pid] = totalUsage / gm.SampleCount[pid]
	}

	return averageUsage
}

// GetAverageGpuUsageByPid calculates the average GPU memory usage for a specific process over the last 30 minutes.
func (gm *GpuMonitor) GetAverageGpuUsageByPid(pid int32) (uint64, error) {
	gm.DataMutex.Lock()
	defer gm.DataMutex.Unlock()

	totalUsage, exists := gm.UsageData[pid]
	if !exists {
		return 0, fmt.Errorf("no data for PID: %d", pid)
	}

	return totalUsage / gm.SampleCount[pid], nil
}

// ClearGpuUsageByPid clears the GPU usage data for a specific process.
func (gm *GpuMonitor) ClearGpuUsageByPid(pid int32) {
	gm.DataMutex.Lock()
	defer gm.DataMutex.Unlock()

	delete(gm.UsageData, pid)
	delete(gm.SampleCount, pid)
}

// ClearAllGpuUsage clears the GPU usage data for all processes.
func (gm *GpuMonitor) ClearAllGpuUsage() {
	gm.DataMutex.Lock()
	defer gm.DataMutex.Unlock()

	gm.UsageData = make(map[int32]uint64)
	gm.SampleCount = make(map[int32]uint64)
}
