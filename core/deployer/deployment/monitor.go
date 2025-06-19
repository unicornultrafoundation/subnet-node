package deployment

import (
	"context"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// MonitorDeployments monitors the deployments running in the cluster
func (s *Service) MonitorDeployments(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.MonitorInterval)

	logger := s.logger.WithField("component", "deployer-monitor")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
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
			}

			if len(unstableList) > 0 {
				logger.WithField("deployments", len(deploymentIDs)).WithField("unstableDeploymentIDs", unstableList).Warn("Some deployments are not stable")
			} else {
				logger.WithField("deployments", len(deploymentIDs)).Info("All deployments are stable")
			}
		}
	}
}
