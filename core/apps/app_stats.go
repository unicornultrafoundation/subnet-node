package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/unicornultrafoundation/subnet-node/core/apps/stats"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) GetUsage(ctx context.Context, appId *big.Int) (*atypes.ResourceUsage, error) {
	containerId := atypes.GetContainerIdFromAppId(appId)
	providerId := big.NewInt(s.accountService.ProviderID())

	// Get resource usage from stats
	statUsage, _ := s.statService.GetStats(containerId)
	if statUsage == nil {
		statUsage = &stats.StatEntry{}
	}

	usage := &atypes.ResourceUsage{
		AppId:             appId,
		ProviderId:        providerId,
		PeerId:            s.peerId.String(),
		UsedCpu:           big.NewInt(int64(statUsage.UsedCpu)),
		UsedGpu:           big.NewInt(int64(statUsage.UsedGpu)),
		UsedMemory:        big.NewInt(int64(statUsage.UsedMemory)),
		UsedStorage:       big.NewInt(int64(statUsage.UsedStorage)),
		UsedUploadBytes:   big.NewInt(int64(statUsage.UsedUploadBytes)),
		UsedDownloadBytes: big.NewInt(int64(statUsage.UsedDownloadBytes)),
		Duration:          big.NewInt(int64(statUsage.Duration)),
	}

	return usage, nil
}

func (s *Service) GetAllRunningContainersUsage(ctx context.Context) (*atypes.ResourceUsage, error) {
	// Fetch all running containers
	containers, err := s.dockerClient.ContainerList(ctx, ctypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch running containers: %w", err)
	}

	// Initialize a single ResourceUsage to aggregate all usage
	totalUsage := &atypes.ResourceUsage{
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
		containerId := strings.TrimPrefix(container.Names[0], "/")
		appId, err := atypes.GetAppIdFromContainerId(containerId)

		if err != nil {
			return nil, err
		}

		usage, err := s.GetUsage(ctx, appId)
		if err != nil {
			usage = &atypes.ResourceUsage{
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
