package apps

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	wstats "github.com/Microsoft/hcsshim/cmd/containerd-shim-runhcs-v1/stats"
	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl/v2"
	"github.com/shirou/gopsutil/process"
	"github.com/unicornultrafoundation/subnet-node/common/networkutil"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

type UsageEntry = ResourceUsage
type UsageMetadata = pbapp.ResourceUsageMetadata

// UsageService tracks the resource usage for multiple containers.
type UsageService struct {
	accountService      *account.AccountService
	mu                  sync.Mutex
	entries             map[string]*UsageEntry
	finalUsages         map[string]*UsageEntry
	containerdClient    *containerd.Client
	containerToMetadata map[string]*UsageMetadata
	stopChan            chan struct{}
}

// NewUsages creates a new UsageService instance.
func NewUsages(containerdClient *containerd.Client, accountService *account.AccountService) *UsageService {
	return &UsageService{
		accountService:      accountService,
		entries:             make(map[string]*UsageEntry),
		finalUsages:         make(map[string]*UsageEntry),
		containerdClient:    containerdClient,
		stopChan:            make(chan struct{}),
		containerToMetadata: make(map[string]*UsageMetadata),
	}
}

// Start starts the UsageService and periodically updates usages for all running containers.
func (s *UsageService) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan: // Exit if the stop channel is closed
			return
		case <-ticker.C: // On every tick, update resource usages for all running containers
			s.updateAllRunningContainersUsage(ctx)
		}
	}
}

// Stop stops the UsageService.
func (s *UsageService) Stop() {
	close(s.stopChan)
	s.ClearUsageData()
}

// ClearUsageData clears all usage data.
func (s *UsageService) ClearUsageData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]*UsageEntry)
	s.finalUsages = make(map[string]*UsageEntry)
	s.containerToMetadata = make(map[string]*UsageMetadata)
}

func (s *Service) GetAllRunningContainersUsage(ctx context.Context) (*ResourceUsage, error) {
	// Set the namespace for the containers
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Fetch all running containers
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch running containers: %w", err)
	}

	// Initialize a single ResourceUsage to aggregate all usage
	totalUsage := &ResourceUsage{
		UsedCpu:           big.NewInt(0),
		UsedGpu:           big.NewInt(0),
		UsedMemory:        big.NewInt(0),
		UsedStorage:       big.NewInt(0),
		UsedUploadBytes:   big.NewInt(0),
		UsedDownloadBytes: big.NewInt(0),
		Duration:          big.NewInt(0),
	}

	providerId := big.NewInt(s.accountService.ProviderID())

	// Iterate over each container and aggregate its resource usage
	for _, container := range containers {
		containerId := container.ID()
		appId, err := getAppIdFromContainerId(containerId)

		if err != nil {
			return nil, err
		}

		usage, err := s.GetUsage(ctx, appId)
		if err != nil {
			usage = &ResourceUsage{
				AppId:             appId,
				ProviderId:        providerId,
				UsedCpu:           big.NewInt(0),
				UsedGpu:           big.NewInt(0),
				UsedMemory:        big.NewInt(0),
				UsedStorage:       big.NewInt(0),
				UsedUploadBytes:   big.NewInt(0),
				UsedDownloadBytes: big.NewInt(0),
				Duration:          big.NewInt(0),
			}
		}

		// Aggregate the usage
		totalUsage.UsedCpu.Add(totalUsage.UsedCpu, usage.UsedCpu)
		totalUsage.UsedGpu.Add(totalUsage.UsedGpu, usage.UsedGpu)
		totalUsage.UsedMemory.Add(totalUsage.UsedMemory, usage.UsedMemory)
		totalUsage.UsedStorage.Add(totalUsage.UsedStorage, usage.UsedStorage)
		totalUsage.UsedUploadBytes.Add(totalUsage.UsedUploadBytes, usage.UsedUploadBytes)
		totalUsage.UsedDownloadBytes.Add(totalUsage.UsedDownloadBytes, usage.UsedDownloadBytes)
		totalUsage.Duration.Add(totalUsage.Duration, usage.Duration)
	}

	return totalUsage, nil
}

func (s *UsageService) updateAllRunningContainersUsage(ctx context.Context) error {
	// Set the namespace for the containers
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Fetch all running containers
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %w", err)
	}

	// Iterate over each container and update its latest resource usage
	for _, container := range containers {
		// Extract appId from container ID
		containerId := container.ID()
		appId, err := getAppIdFromContainerId(containerId)

		if err != nil {
			return err
		}

		usage, pid, err := s.getUsageFromContainer(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to get latest resource usage for appId %s: %v", appId.String(), err)
		}

		if pid == nil {
			return fmt.Errorf("failed to get pid of appId %s: %v", appId.String(), err)
		}

		// Update usage metadata if the pid is new for appId
		s.syncUsageMetadata(appId, *pid)
		s.mu.Lock()
		s.updateUsage(appId, *pid, usage)
		s.mu.Unlock()
	}

	return nil
}

// Get directly resource usage from container
func (s *UsageService) getUsageFromContainer(ctx context.Context, appId *big.Int) (*ResourceUsage, *uint32, error) {
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	app := App{ID: appId}

	// Load the container
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get the process
	pid := task.Pid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, nil, err
	}

	// Get network I/O counters
	netIO, err := proc.NetIOCounters(false)
	if err != nil {
		return nil, nil, err
	}

	// Detect the network interfaces connected to the internet
	internetInterfaces, err := networkutil.DetectInternetInterfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to detect internet interfaces: %v", err)
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

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Unmarshal metrics data
	var data interface{}
	switch {
	case typeurl.Is(metric.Data, (*v1.Metrics)(nil)):
		data = &v1.Metrics{}
	case typeurl.Is(metric.Data, (*v2.Metrics)(nil)):
		data = &v2.Metrics{}
	case typeurl.Is(metric.Data, (*wstats.Statistics)(nil)):
		data = &wstats.Statistics{}
	default:
		return nil, nil, errors.New("cannot convert metric data to cgroups.Metrics or windows.Statistics")
	}
	if err := typeurl.UnmarshalTo(metric.Data, data); err != nil {
		return nil, nil, err
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
	case *wstats.Statistics:
		usedCpu = metrics.GetWindows().Processor.TotalRuntimeNS
		usedMemory = metrics.GetWindows().Memory.MemoryUsageCommitBytes
	default:
		return nil, nil, fmt.Errorf("unsupported metrics type")
	}

	// Get used storage
	var storageBytes int64
	info, err := container.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container info: %v", err)
	}
	snapshotUsage, err := s.containerdClient.SnapshotService(info.Snapshotter).Usage(ctx, info.SnapshotKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get snapshot usage: %v", err)
	}
	storageBytes = snapshotUsage.Size

	// Get duration of the pid running this app
	createTime, err := proc.CreateTime()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get process create time: %w", err)
	}
	now := time.Now().UnixMilli()
	duration := now - createTime
	durationSeconds := duration / 1000

	// Extract resource usage details from the metrics
	usage := &ResourceUsage{
		AppId:             appId,
		UsedCpu:           big.NewInt(int64(usedCpu)),
		UsedMemory:        big.NewInt(int64(usedMemory)),
		UsedUploadBytes:   big.NewInt(int64(totalRxBytes)), // Set if applicable
		UsedDownloadBytes: big.NewInt(int64(totalTxBytes)), // Set if applicable
		UsedStorage:       big.NewInt(storageBytes),
		Duration:          big.NewInt(durationSeconds), // Calculate or set duration if available
	}

	return usage, &pid, nil
}

func (s *UsageService) getUsageFromExternal(appId *big.Int) (*ResourceUsage, error) {
	providerId := big.NewInt(s.accountService.ProviderID())

	// Get usage from memory first
	// Get all pids had run the appId from usage metadata
	s.mu.Lock()
	usageMetadata, hasUsageMetadata := s.getUsageMetadata(appId)
	if hasUsageMetadata {
		// Some records exist in the memory
		// Get all usages of the appId
		totalUsage := &ResourceUsage{
			AppId:             appId,
			ProviderId:        providerId,
			UsedCpu:           big.NewInt(0),
			UsedGpu:           big.NewInt(0),
			UsedMemory:        big.NewInt(0),
			UsedStorage:       big.NewInt(0),
			UsedUploadBytes:   big.NewInt(0),
			UsedDownloadBytes: big.NewInt(0),
			Duration:          big.NewInt(0),
		}

		totalStorageRecord := big.NewInt(0)
		for index, pid := range usageMetadata.Pids {
			usage, hasUsage := s.getUsage(appId, pid)
			if !hasUsage {
				s.mu.Unlock()
				return nil, fmt.Errorf("failed to get resource usage from memory for appId %s - pid %d: usage should not exist", appId.String(), pid)
			}
			usage = fillDefaultResourceUsage(usage)

			// Aggregate the usage
			totalUsage.UsedCpu.Add(totalUsage.UsedCpu, usage.UsedCpu)
			totalUsage.UsedGpu.Add(totalUsage.UsedGpu, usage.UsedGpu)
			totalUsage.UsedUploadBytes.Add(totalUsage.UsedUploadBytes, usage.UsedUploadBytes)
			totalUsage.UsedDownloadBytes.Add(totalUsage.UsedDownloadBytes, usage.UsedDownloadBytes)
			totalUsage.Duration.Add(totalUsage.Duration, usage.Duration)

			if index < 5 { // Only retrieve used memory/storage from <= 5 latest pid
				totalUsage.UsedMemory.Add(totalUsage.UsedMemory, usage.UsedMemory)
				totalUsage.UsedStorage.Add(totalUsage.UsedStorage, usage.UsedStorage)
				totalStorageRecord.Add(totalStorageRecord, big.NewInt(1))
			}
		}
		s.mu.Unlock()

		// Get average used memory/storage from <= 5 latest pid
		totalUsage.UsedMemory.Div(totalUsage.UsedMemory, totalStorageRecord)
		totalUsage.UsedStorage.Div(totalUsage.UsedStorage, totalStorageRecord)

		return totalUsage, nil
	}

	// Try get usage from last finalized record
	cachedFinalUsage, exist := s.finalUsages[appId.String()]
	s.mu.Unlock()
	if exist {
		return cachedFinalUsage, nil
	}

	// No record found from memory
	// Get usage from smart contract
	usage, err := s.accountService.AppStore().GetDeployment(nil, appId, providerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage from cache & smart contract for appId %s, providerId %s: %v", appId.String(), providerId.String(), err)
	}

	return convertToResourceUsage(usage, appId, providerId), nil
}

// Add new pid into usage metadata if it hadn't run appId before
func (s *UsageService) syncUsageMetadata(appId *big.Int, pid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, hasUsage := s.getUsage(appId, pid)

	if hasUsage {
		return
	}

	usageMetadata := &pbapp.ResourceUsageMetadata{
		Pids: *new([]uint32),
	}
	existingUsageMetadata, hasUsageMetadata := s.getUsageMetadata(appId)

	if hasUsageMetadata {
		usageMetadata = existingUsageMetadata
	}

	// Append new pid on the top of `Pids` list
	usageMetadata.Pids = append([]uint32{pid}, usageMetadata.Pids...)

	s.updateUsageMetadata(appId, usageMetadata)
}

// FinalizeUsages finalizes a given usage after being submitted to smart contract
func (s *UsageService) finalizeUsages(usage *ResourceUsage) error {
	appId := usage.AppId

	s.mu.Lock()
	defer s.mu.Unlock()
	// Save latest submitted usage data
	s.finalUsages[appId.String()] = usage

	// Clear previous resource usage cached data
	usageMetadata, hasUsageMetadata := s.getUsageMetadata(appId)
	if !hasUsageMetadata {
		return fmt.Errorf("resource usage metadata not found for appID: %s", appId)
	}

	for _, pid := range usageMetadata.Pids {
		_, hasUsage := s.getUsage(appId, pid)
		if !hasUsage {
			return fmt.Errorf("failed to get resource usage from memory for appId %s - pid %d: usage should not exist", appId.String(), pid)
		}

		delete(s.entries, getUsageKey(appId, pid))
	}

	delete(s.containerToMetadata, getUsageMetadataKey(appId))

	return nil
}

// Get resource usage metadata of an appId
func (s *UsageService) getUsageMetadata(appId *big.Int) (*UsageMetadata, bool) {
	usageMetadataKey := getUsageMetadataKey(appId)
	usageMetadata, exist := s.containerToMetadata[usageMetadataKey]

	return usageMetadata, exist
}

// Update resource usage metadata of an appId
func (s *UsageService) updateUsageMetadata(appId *big.Int, usageMetadata *pbapp.ResourceUsageMetadata) {
	usageMetadataKey := getUsageMetadataKey(appId)
	s.containerToMetadata[usageMetadataKey] = usageMetadata
}

// Get resource usage created by a pid of an appId from memory
func (s *UsageService) getUsage(appId *big.Int, pid uint32) (*UsageEntry, bool) {
	resourceUsageKey := getUsageKey(appId, pid)
	usage, exist := s.entries[resourceUsageKey]

	return usage, exist
}

// Update resource usage created by a specific pid for appId
func (s *UsageService) updateUsage(appId *big.Int, pid uint32, usage *ResourceUsage) {
	resourceUsageKey := getUsageKey(appId, pid)
	s.entries[resourceUsageKey] = usage
}

// Get key to retrieve resource usage created by a pid which had run appId
func getUsageKey(appId *big.Int, pid uint32) string {
	return fmt.Sprintf("%s-%s-%d", RESOURCE_USAGE_KEY, appId.String(), pid)
}

// Get key to retrieve resource usage metadata of an appId
func getUsageMetadataKey(appId *big.Int) string {
	return fmt.Sprintf("%s-%s", RESOURCE_USAGE_KEY, appId.String())
}
