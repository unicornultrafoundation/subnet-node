package k8scluster

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	"go.uber.org/zap"
)

// submitBid submits a bid for the deployment
func (s *Service) submitBid(ctx context.Context, requirements *manifest.ResourceRequirements) (*types.K8sBid, error) {
	// Calculate bid amount based on resource requirements
	amount := s.calculateBidAmount(requirements)

	// TODO: Submit bid to marketplace

	bid := &types.K8sBid{
		ID:        fmt.Sprintf("bid-%d", time.Now().Unix()),
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	return bid, nil
}

// calculateBidAmount calculates the bid amount based on resource requirements
func (s *Service) calculateBidAmount(requirements *manifest.ResourceRequirements) float64 {
	// Simple pricing model:
	// CPU: $0.1 per unit
	// Memory: $0.05 per GB
	// Storage: $0.01 per GB
	// GPU: $0.1 per unit
	storageSize := 0.0
	for _, storage := range requirements.Storage {
		size, err := strconv.ParseFloat(storage.Size, 64)
		if err != nil {
			return 0
		}
		storageSize += size
	}
	cpuUnits, err := strconv.ParseFloat(requirements.CPU.Request, 64)
	if err != nil {
		return 0
	}
	memoryUnits, err := strconv.ParseFloat(requirements.Memory.Request, 64)
	if err != nil {
		return 0
	}
	cpuCost := cpuUnits * 0.1
	memoryCost := memoryUnits * 0.05
	storageCost := storageSize * 0.01
	gpuCost := float64(requirements.GPU.Units) * 0.1

	return cpuCost + memoryCost + storageCost + gpuCost
}

// updateMarketplaceStatus updates the marketplace status
func (s *Service) updateMarketplaceStatus(ctx context.Context, id string, status string, errMsg string) {
	s.logger.Info("updateMarketplaceStatus called", zap.String("deployment", id), zap.String("status", status), zap.String("error", errMsg))
}
