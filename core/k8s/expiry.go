package k8s

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	btypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
)

type expiryService struct {
	bidMarket        btypes.BidMarketContract
	ethClient        *ethclient.Client
	orderGracePeriod time.Duration
}

func newExpiryService(bidMarket btypes.BidMarketContract, ethClient *ethclient.Client) *expiryService {
	return &expiryService{
		bidMarket:        bidMarket,
		ethClient:        ethClient,
		orderGracePeriod: time.Hour * 24, // TODO: get from Smart Contract
	}
}

func (s *expiryService) CheckDeploymentExpiry(ctx context.Context, deploymentID string) (etypes.DeploymentExpiryInfo, error) {
	info := etypes.DeploymentExpiryInfo{
		DeploymentID: deploymentID,
		Status:       etypes.DeploymentExpiryStatusUnknown,
		TimeLeft:     0,
	}

	// convert deploymentID to big.Int
	orderID, ok := new(big.Int).SetString(deploymentID, 10)
	if !ok {
		return info, fmt.Errorf("failed to convert leaseID to big.Int")
	}
	order, err := s.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return info, err
	}

	if order.Status != btypes.OrderStatusAccepted {
		// If the order is cancelled, we should delete the deployment from the cluster
		info.Status = etypes.DeploymentExpiryStatusDeleted
		return info, nil
	}

	expiredAt := order.ExpiredAt.Int64()
	block, err := s.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return info, fmt.Errorf("failed to get block by number: %w", err)
	}

	now := block.Time()

	// We have 3 cases:
	// 1. expiredAt > now: the deployment is active
	// 2. expiredAt <= now:
	// 2.1. now - expiredAt < orderGracePeriod: the deployment is expired, but we should keep it in the cluster for the grace period
	// 2.2. now - expiredAt >= orderGracePeriod: the deployment should be deleted from the cluster

	timeLeft := expiredAt - int64(now)
	if timeLeft > 0 {
		info.Status = etypes.DeploymentExpiryStatusActive
		info.TimeLeft = timeLeft
		return info, nil
	}

	expiredTime := 0 - timeLeft
	orderGracePeriod := int64(s.orderGracePeriod.Seconds())

	if expiredTime < orderGracePeriod {
		// If the deployment is expired, but the time left is less than orderGracePeriod,
		// we should keep it in the cluster for the grace period
		info.Status = etypes.DeploymentExpiryStatusExpired
		info.TimeLeft = orderGracePeriod - expiredTime
		return info, nil
	}

	// If the deployment is expired, and the time left is greater than orderGracePeriod,
	// we should delete it from the cluster
	info.Status = etypes.DeploymentExpiryStatusDeleted
	return info, nil
}
