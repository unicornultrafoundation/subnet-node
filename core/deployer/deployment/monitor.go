package deployment

import (
	"context"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// MonitorDeployments monitors the deployments running in the cluster
func (s *Service) MonitorDeployments(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.MonitorInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for deployment requests in the datastore
			deploymentIDs := s.GetDeploymentListCache()

			unstableCount := 0
			for _, deploymentID := range deploymentIDs {
				deployment, err := s.kubeClient.GetDeployment(ctx, deploymentID)
				if err != nil {
					s.logger.Error("Failed to get deployment", err)
					unstableCount++
					continue
				}

				if deployment.Status.State != types.DeploymentStateRunning {
					s.logger.WithField("deploymentID", deploymentID).WithField("state", deployment.Status.State).Warn("Deployment not running, waiting for it to be stable")
					unstableCount++
				}
			}

			if unstableCount > 0 {
				s.logger.WithField("deployments", len(deploymentIDs)).WithField("unstable", unstableCount).Warn("Deployments are not stable, waiting for them to be stable")
			} else {
				s.logger.WithField("deployments", len(deploymentIDs)).Info("All deployments are stable")
			}
		}
	}
}
