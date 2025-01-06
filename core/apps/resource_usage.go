package apps

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	wstats "github.com/Microsoft/hcsshim/cmd/containerd-shim-runhcs-v1/stats"
	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"
	"github.com/containerd/typeurl/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-datastore"
	"github.com/shirou/gopsutil/process"
	pusage "github.com/unicornultrafoundation/subnet-node/proto/subnet/usage"
)

func (s *Service) startMonitoringUsage(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		case <-ticker.C: // On every tick, update all latest running containers's resource usages
			s.updateAllRunningContainersUsage(ctx)
		}
	}
}

func (s *Service) GetUsage(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	return s.getUsageFromExternal(ctx, appId)
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

func (s *Service) updateAllRunningContainersUsage(ctx context.Context) error {
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

		usage, pid, err := s.getUsage(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to get latest resource usage for appId %s: %v", appId.String(), err)
		}

		if pid == nil {
			return fmt.Errorf("failed to get pid of appId %s: %v", appId.String(), err)
		}

		// Update usage metadata if the pid is new for appId
		err = s.syncUsageMetadata(ctx, appId, *pid)
		if err != nil {
			return fmt.Errorf("failed to sync pid %d for appId %s: %v", *pid, appId.String(), err)
		}
		err = s.updateUsage(ctx, appId, *pid, usage)

		if err != nil {
			return fmt.Errorf("failed to update resource usage for container %s: %w", containerId, err)
		}
	}

	return nil
}

// Get directly resource usage from container
func (s *Service) getUsage(ctx context.Context, appId *big.Int) (*ResourceUsage, *uint32, error) {
	// Load the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	app := App{ID: appId}
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get metrics: %w", err)
	}

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

	// Lấy thông tin Network I/O theo PID
	pid := task.Pid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, nil, err
	}

	var totalRxBytes uint64 = 0
	var totalTxBytes uint64 = 0
	netIO, err := proc.NetIOCounters(false)

	if err != nil {
		return nil, nil, err
	}

	for _, stat := range netIO {
		totalRxBytes += stat.BytesRecv
		totalTxBytes += stat.BytesSent
	}

	// Get used storage
	var storageBytes int64 = 0
	info, err := container.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container info: %v", err)
	}
	snapshotKey := info.SnapshotKey
	snapshotService := s.containerdClient.SnapshotService(info.Snapshotter)
	snapshotUsage, err := snapshotService.Usage(ctx, snapshotKey)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, nil, fmt.Errorf("snapshot %s not found: %v", snapshotKey, err)
		}
		return nil, nil, fmt.Errorf("failed to get snapshot usage: %v", err)
	}
	storageBytes = snapshotUsage.Size

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
		UsedUploadBytes:   big.NewInt(int64(totalRxBytes)), // Set if applicable
		UsedDownloadBytes: big.NewInt(int64(totalTxBytes)), // Set if applicable
		Duration:          big.NewInt(durationSeconds),     // Calculate or set duration if available
		UsedStorage:       big.NewInt(storageBytes),
	}
	switch metrics := data.(type) {
	case *v1.Metrics:
		usage.UsedCpu = big.NewInt(int64(metrics.CPU.Usage.Total))
		usage.UsedMemory = big.NewInt(int64(metrics.Memory.Usage.Usage))
	case *v2.Metrics:
		usage.UsedCpu = big.NewInt(int64(metrics.CPU.UsageUsec))
		usage.UsedMemory = big.NewInt(int64(metrics.Memory.Usage))
	case *wstats.Statistics:
		usage.UsedCpu = big.NewInt(int64(metrics.GetWindows().Processor.TotalRuntimeNS))
		usage.UsedMemory = big.NewInt(int64(metrics.GetWindows().Memory.MemoryUsageCommitBytes))
	default:
		return nil, nil, errors.New("unsupported metrics type")
	}

	// Optionally, calculate or set the duration if metrics provide a timestamp
	// For example:
	// usage.Duration = calculateDurationBasedOnMetrics(metrics)

	return usage, &pid, nil
}

func (s *Service) getUsageFromExternal(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	// Get usage from datastore first
	// Get all pids had run the appId from usage metadata
	usageMetadata, err := s.getUsageMetadata(ctx, appId)
	if err == nil {
		// Some records exist in the datastore
		// Get all usages of the appId
		totalUsage := &ResourceUsage{
			AppId:             appId,
			SubnetId:          &s.subnetID,
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
			usage, err := s.getUsageFromStorage(ctx, appId, pid)
			if err != nil {
				return nil, fmt.Errorf("failed to get resource usage from datastore for appId %s - pid %d: %v", appId.String(), pid, err)
			}
			usage = fillDefaultResourceUsage(usage)

			// Aggregate the usage
			totalUsage.UsedCpu.Add(totalUsage.UsedCpu, usage.UsedCpu)
			totalUsage.UsedGpu.Add(totalUsage.UsedGpu, usage.UsedGpu)
			totalUsage.UsedUploadBytes.Add(totalUsage.UsedUploadBytes, usage.UsedUploadBytes)
			totalUsage.UsedDownloadBytes.Add(totalUsage.UsedDownloadBytes, usage.UsedDownloadBytes)
			totalUsage.Duration.Add(totalUsage.Duration, usage.Duration)

			if index < 5 {
				// Only retrieve used memory/storage from <= 5 latest pid
				totalUsage.UsedMemory.Add(totalUsage.UsedMemory, usage.UsedMemory)
				totalUsage.UsedStorage.Add(totalUsage.UsedStorage, usage.UsedStorage)

				totalStorageRecord.Add(totalStorageRecord, big.NewInt(1))
			}
		}

		// Get average used memory/storage from <= 5 latest pid
		totalUsage.UsedMemory.Div(totalUsage.UsedMemory, totalStorageRecord)
		totalUsage.UsedStorage.Div(totalUsage.UsedStorage, totalStorageRecord)

		return totalUsage, nil
	}

	// No record found from datastore
	// Get usage from smart contract
	usage, err := s.subnetAppRegistry.GetAppNode(nil, appId, &s.subnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage from datastore & smart contract for appId %s:%s %v", appId.String(), s.subnetID.String(), err)
	}

	return convertToResourceUsage(usage, appId, &s.subnetID), nil
}

// Get resource usage created by a specific pid for appId from datastore
func (s *Service) getUsageFromStorage(ctx context.Context, appId *big.Int, pid uint32) (*ResourceUsage, error) {
	resourceUsageKey := getUsageKey(appId, pid)
	usageData, err := s.Datastore.Get(ctx, resourceUsageKey)

	if err == nil { // If the record exists in the datastore
		var usage pusage.ResourceUsage
		if err := proto.Unmarshal(usageData, &usage); err == nil {
			return convertUsageFromProto(usage), nil
		}
	}

	return nil, fmt.Errorf("failed to get resource usage from datastore for appId %s: %v", appId.String(), err)
}

// Update resource usage created by a specific pid for appId to datastore
func (s *Service) updateUsage(ctx context.Context, appId *big.Int, pid uint32, usage *ResourceUsage) error {
	resourceUsageData, err := proto.Marshal(convertUsageToProto(*usage))
	if err != nil {
		return fmt.Errorf("failed to marshal resource usage for appId %s: %v", appId.String(), err)
	}

	resourceUsageKey := getUsageKey(appId, pid)
	if err := s.Datastore.Put(ctx, resourceUsageKey, resourceUsageData); err != nil {
		return fmt.Errorf("failed to update resource usage for appId %s: %v", appId.String(), err)
	}
	return nil
}

// Get resource usage info of an appId
func (s *Service) getUsageMetadata(ctx context.Context, appId *big.Int) (*pusage.ResourceUsageMetadata, error) {
	usageMetadataKey := getUsageMetadataKey(appId)
	usageMetadataBytes, err := s.Datastore.Get(ctx, usageMetadataKey)
	if err != nil {
		return nil, err
	}

	var usageMetadata pusage.ResourceUsageMetadata
	err = proto.Unmarshal(usageMetadataBytes, &usageMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage metadata for appId %s: %w", appId.String(), err)
	}

	return &usageMetadata, nil
}

// Update resource usage info of an appId
func (s *Service) updateUsageMetadata(ctx context.Context, appId *big.Int, usageMetadata *pusage.ResourceUsageMetadata) error {
	usageMetadataBytes, err := proto.Marshal(usageMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal usage metadata for appId %s: %v", appId.String(), err)
	}

	usageMetadataKey := getUsageMetadataKey(appId)
	if err := s.Datastore.Put(ctx, usageMetadataKey, usageMetadataBytes); err != nil {
		return fmt.Errorf("failed to update usage metadata for appId %s: %v", appId.String(), err)
	}

	return nil
}

// Add new pid into usage metadata if it hadn't run appId before
func (s *Service) syncUsageMetadata(ctx context.Context, appId *big.Int, pid uint32) error {
	hasPid, err := s.hasUsage(ctx, appId, pid)
	if err != nil {
		return fmt.Errorf("failed to check pid for appId %s: %w", appId.String(), err)
	}

	if hasPid {
		return nil
	}

	hasUsageMetadata, err := s.hasUsageMetadata(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to check usage metadata for appId %s: %w", appId.String(), err)
	}

	usageMetadata := &pusage.ResourceUsageMetadata{
		Pids: *new([]uint32),
	}
	if hasUsageMetadata {
		usageMetadata, err = s.getUsageMetadata(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to get pidList for appId %s: %w", appId.String(), err)
		}
	}

	// Append new pid on the top of `Pids` list
	usageMetadata.Pids = append([]uint32{pid}, usageMetadata.Pids...)

	return s.updateUsageMetadata(ctx, appId, usageMetadata)
}

// Get key to retrieve resource usage from datastore created by a pid which had run appId
func getUsageKey(appId *big.Int, pid uint32) datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s-%s-%d", RESOURCE_USAGE_KEY, appId.String(), pid))
}

// Check if a pid creating resource usage of an `appId` has existed or not
func (s *Service) hasUsage(ctx context.Context, appId *big.Int, pid uint32) (bool, error) {
	resourceUsageKey := getUsageKey(appId, pid)
	hasUsage, err := s.Datastore.Has(ctx, resourceUsageKey)

	if err != nil {
		return false, fmt.Errorf("failed to check resource usage for appId %s: %w", appId.String(), err)
	}

	return hasUsage, nil
}

// Get key to retrieve resource usage metadata of an appId
func getUsageMetadataKey(appId *big.Int) datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s-%s", RESOURCE_USAGE_KEY, appId.String()))
}

// Check if an `appId` already has an usage metadata or not
func (s *Service) hasUsageMetadata(ctx context.Context, appId *big.Int) (bool, error) {
	usageMetadataKey := getUsageMetadataKey(appId)
	hasUsageMetadata, err := s.Datastore.Has(ctx, usageMetadataKey)

	if err != nil {
		return false, fmt.Errorf("failed to check usage metadata for appId %s: %w", appId.String(), err)
	}

	return hasUsageMetadata, nil
}
