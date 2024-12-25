package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/containerd/containerd/namespaces"
)

// ...existing code...

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
		// Extract appId from container ID
		containerID := container.ID()
		appIDStr := strings.TrimPrefix(containerID, "subnet-app-")
		appID, ok := new(big.Int).SetString(appIDStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid container ID: %s", containerID)
		}

		usage, err := s.GetUsage(ctx, appID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource usage for container %s: %w", containerID, err)
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

// ...existing code...
