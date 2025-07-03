package deployment

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
	bidtypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// MonitorDeployments monitors the deployments running in the cluster
func (s *Service) MonitorDeployments(ctx context.Context) error {
	logger := s.logger.WithField("component", "deployer-monitor")
	err := s.monitorDeployments(ctx, logger.Logger)
	if err != nil {
		logger.Error("Failed to monitor deployments", err)
	}

	ticker := time.NewTicker(s.cfg.MonitorInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err = s.monitorDeployments(ctx, logger.Logger)
			if err != nil {
				logger.Error("Failed to monitor deployments", err)
			}
		}
	}
}

func (s *Service) monitorDeployments(ctx context.Context, logger *logrus.Logger) error {
	// Check for deployment requests in the datastore
	deploymentIDs := s.GetDeploymentListCache()

	unstableList := []string{}
	for _, deploymentID := range deploymentIDs {
		deployment, err := s.kubeClient.GetDeployment(ctx, deploymentID)
		if err != nil {
			logger.WithField("deploymentID", deploymentID).Error("Failed to get deployment", err)
			unstableList = append(unstableList, deploymentID)
			continue
		}

		if deployment.Status.State != types.DeploymentStateRunning {
			logger.WithField("deploymentID", deploymentID).WithField("state", deployment.Status.State).Warn("Deployment not running, waiting for it to be stable")
			unstableList = append(unstableList, deploymentID)
		}

		// Check if the deployment is expired
		expiryStatus, err := s.HandleDeploymentExpiry(ctx, deploymentID, logger)
		if err != nil {
			logger.WithField("deploymentID", deploymentID).WithField("expiryStatus", expiryStatus).Error("Failed to handle deployment expiry", err)
		}
	}

	if len(unstableList) > 0 {
		logger.WithField("deployments", len(deploymentIDs)).WithField("unstableDeploymentIDs", unstableList).Warn("Some deployments are not stable")
	} else {
		logger.WithField("deployments", len(deploymentIDs)).Info("All deployments are stable")
	}

	return nil
}

type DeploymentExpiryStatus string

const (
	DeploymentExpiryStatusExpired DeploymentExpiryStatus = "expired"
	DeploymentExpiryStatusActive  DeploymentExpiryStatus = "active"
	DeploymentExpiryStatusDeleted DeploymentExpiryStatus = "deleted"
	DeploymentExpiryStatusUnknown DeploymentExpiryStatus = "unknown"
)

func (s *Service) CheckDeploymentExpiryOnChain(ctx context.Context, deploymentID string) (DeploymentExpiryStatus, error) {
	// convert deploymentID to big.Int
	orderID, ok := new(big.Int).SetString(deploymentID, 10)
	if !ok {
		return DeploymentExpiryStatusUnknown, fmt.Errorf("failed to convert deploymentID to big.Int")
	}
	order, err := s.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return DeploymentExpiryStatusUnknown, err
	}

	if order.Status != bidtypes.OrderStatusAccepted {
		// If the order is cancelled, we should delete the deployment from the cluster
		return DeploymentExpiryStatusDeleted, nil
	}

	expiredAt := order.ExpiredAt.Int64()

	block, err := s.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return DeploymentExpiryStatusUnknown, err
	}

	now := block.Time()

	// We have 3 cases:
	// 1. expiredAt > now: the deployment is active
	// 2. expiredAt <= now:
	// 2.1. now - expiredAt < orderGracePeriod: the deployment is expired, but we should keep it in the cluster for the grace period
	// 2.2. now - expiredAt >= orderGracePeriod: the deployment should be deleted from the cluster

	timeLeft := expiredAt - int64(now)
	if timeLeft > 0 {
		return DeploymentExpiryStatusActive, nil
	}

	expiredTime := 0 - timeLeft
	orderGracePeriod := int64(s.orderGracePeriod.Seconds())

	if expiredTime < orderGracePeriod {
		// If the deployment is expired, but the time left is less than orderGracePeriod,
		// we should keep it in the cluster for the grace period
		return DeploymentExpiryStatusExpired, nil
	}

	// If the deployment is expired, and the time left is greater than orderGracePeriod,
	// we should delete it from the cluster
	return DeploymentExpiryStatusDeleted, nil
}

func (s *Service) HandleDeploymentExpiry(ctx context.Context, deploymentID string, logger *logrus.Logger) (DeploymentExpiryStatus, error) {
	status, err := s.CheckDeploymentExpiryOnChain(ctx, deploymentID)
	if err != nil {
		return DeploymentExpiryStatusUnknown, err
	}

	switch status {
	case DeploymentExpiryStatusDeleted:
		// If the deployment expired time is more than orderGracePeriod or the order is cancelled, we should delete it from the cluster
		logger.WithField("deploymentID", deploymentID).Warn("Deployment expired, cleaning up")
		err := s.CleanupDeployment(ctx, deploymentID)
		if err != nil {
			return status, fmt.Errorf("failed to cleanup expired deployment: %w", err)
		} else {
			logger.WithField("deploymentID", deploymentID).Info("Successfully cleaned up expired deployment")
		}
		return status, nil

	case DeploymentExpiryStatusExpired:
		isRunning, err := s.AddDeploymentExpiryListCache(ctx, deploymentID)
		if err != nil {
			return status, fmt.Errorf("failed to add deployment to expiry list: %w", err)
		}
		if isRunning {
			// If the deployment is running, scale it down to 0
			logger.WithField("deploymentID", deploymentID).Warn("Deployment expired, pausing")
			err := s.kubeClient.ScaleDeployment(ctx, deploymentID, 0)
			if err != nil {
				return status, fmt.Errorf("failed to scale deployment to 0: %w", err)
			} else {
				logger.WithField("deploymentID", deploymentID).Info("Successfully scaled deployment to 0")
			}
		}
		return status, nil

	case DeploymentExpiryStatusActive:
		// If the deployment is extended, scale it up to the original count in deployment request
		if s.HasDeploymentExpiryListCache(deploymentID) {
			logger.WithField("deploymentID", deploymentID).Warn("Deployment is extended, scaling up to original count")
			deploymentRequest, err := s.store.GetDeploymentRequest(ctx, deploymentID)
			if err != nil {
				return status, fmt.Errorf("failed to get deployment request: %w", err)
			}
			err = s.kubeClient.ScaleDeploymentToOriginal(ctx, deploymentRequest)
			if err != nil {
				return status, fmt.Errorf("failed to scale deployment to original count: %w", err)
			} else {
				logger.WithField("deploymentID", deploymentID).Info("Successfully scaled deployment to original count")
				if err := s.DeleteDeploymentExpiryListCache(ctx, deploymentID); err != nil {
					return status, fmt.Errorf("failed to delete deployment expiry list: %w", err)
				}
			}
		}

	default:
		return status, nil
	}

	return DeploymentExpiryStatusUnknown, nil
}
