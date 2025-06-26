package deployments

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// Service represents the main deployment service that manages multiple deployment types
type Service struct {
	mu       sync.RWMutex
	cfg      *config.C
	logger   *logrus.Logger
	stopChan chan struct{}

	// Managers
	eventManager    EventManager
	storageManager  StorageManager
	resourceManager ResourceManager

	// Deployment managers by type
	deploymentManagers map[DeploymentType]DeploymentManager

	// Factory for creating deployment managers
	factory DeploymentFactory
}

// NewService creates a new deployment service
func NewService(cfg *config.C, logger *logrus.Logger) *Service {
	return &Service{
		cfg:                cfg,
		logger:             logger,
		stopChan:           make(chan struct{}),
		deploymentManagers: make(map[DeploymentType]DeploymentManager),
	}
}

// SetManagers sets the managers for the service
func (s *Service) SetManagers(
	eventManager EventManager,
	storageManager StorageManager,
	resourceManager ResourceManager,
	factory DeploymentFactory,
) {
	s.eventManager = eventManager
	s.storageManager = storageManager
	s.resourceManager = resourceManager
	s.factory = factory
}

// RegisterDeploymentManager registers a deployment manager for a specific type
func (s *Service) RegisterDeploymentManager(deploymentType DeploymentType, manager DeploymentManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deploymentManagers[deploymentType] = manager
}

// GetDeploymentManager gets a deployment manager for a specific type
func (s *Service) GetDeploymentManager(deploymentType DeploymentType) (DeploymentManager, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	manager, exists := s.deploymentManagers[deploymentType]
	if !exists {
		return nil, fmt.Errorf("deployment manager not found for type: %s", deploymentType)
	}

	return manager, nil
}

// Start starts the deployment service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting deployment service...")

	// Initialize deployment managers for supported types
	if s.factory != nil {
		supportedTypes := s.factory.GetSupportedTypes()
		for _, deploymentType := range supportedTypes {
			config := &DeploymentConfig{
				Type: deploymentType,
			}

			manager, err := s.factory.CreateManager(ctx, deploymentType, config)
			if err != nil {
				s.logger.WithError(err).WithField("deployment_type", deploymentType).Error("Failed to create deployment manager")
				continue
			}

			s.RegisterDeploymentManager(deploymentType, manager)
			s.logger.WithField("deployment_type", deploymentType).Info("Deployment manager registered")
		}
	}

	s.logger.Info("Deployment service started successfully")
	return nil
}

// Stop stops the deployment service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping deployment service...")

	close(s.stopChan)

	// Stop all deployment managers
	s.mu.RLock()
	for deploymentType, manager := range s.deploymentManagers {
		if err := s.stopDeploymentManager(ctx, manager); err != nil {
			s.logger.WithError(err).WithField("deployment_type", deploymentType).Error("Failed to stop deployment manager")
		}
	}
	s.mu.RUnlock()

	s.logger.Info("Deployment service stopped successfully")
	return nil
}

// CreateDeployment creates a new deployment
func (s *Service) CreateDeployment(ctx context.Context, deployment *Deployment) error {
	s.logger.WithFields(logrus.Fields{
		"deployment_id": deployment.ID,
		"type":          deployment.Type,
	}).Info("Creating deployment")

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return fmt.Errorf("failed to get deployment manager: %w", err)
	}

	// Create deployment using the specific manager
	return manager.CreateDeployment(ctx, deployment)
}

// GetDeployment gets a deployment by ID
func (s *Service) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	// Try to load from storage first to get the deployment type
	deployment, err := s.storageManager.LoadDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Get deployment using the specific manager
	return manager.GetDeployment(ctx, deploymentID)
}

// ListDeployments lists deployments for a tenant
func (s *Service) ListDeployments(ctx context.Context, tenantID string) ([]*Deployment, error) {
	// For now, return all deployments from storage
	// In the future, this could be optimized to filter by tenant
	return s.storageManager.ListDeployments(ctx)
}

// StartDeployment starts a deployment
func (s *Service) StartDeployment(ctx context.Context, deploymentID string) error {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return err
	}

	// Start deployment using the specific manager
	return manager.StartDeployment(ctx, deploymentID)
}

// StopDeployment stops a deployment
func (s *Service) StopDeployment(ctx context.Context, deploymentID string) error {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return err
	}

	// Stop deployment using the specific manager
	return manager.StopDeployment(ctx, deploymentID)
}

// DeleteDeployment deletes a deployment
func (s *Service) DeleteDeployment(ctx context.Context, deploymentID string) error {
	s.logger.WithField("deployment_id", deploymentID).Info("Deleting deployment")

	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return fmt.Errorf("failed to get deployment manager: %w", err)
	}

	// Delete deployment using the specific manager
	return manager.DeleteDeployment(ctx, deploymentID)
}

// GetDeploymentLogs gets logs for a deployment
func (s *Service) GetDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, tail int) (io.ReadCloser, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Get logs using the specific manager
	return manager.GetDeploymentLogs(ctx, deploymentID, serviceName, tail)
}

// StreamDeploymentLogs streams logs for a deployment in real-time
func (s *Service) StreamDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, follow bool) (<-chan LogEntry, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Stream logs using the specific manager
	return manager.StreamDeploymentLogs(ctx, deploymentID, serviceName, follow)
}

// ExecConsole executes a command in a deployment container
func (s *Service) ExecConsole(ctx context.Context, deploymentID string, serviceName string, command []string, tty bool) (ExecSession, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Execute command using the specific manager
	return manager.ExecConsole(ctx, deploymentID, serviceName, command, tty)
}

// InspectDeployment gets detailed information about a deployment
func (s *Service) InspectDeployment(ctx context.Context, deploymentID string) (*DeploymentInspection, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Inspect deployment using the specific manager
	return manager.InspectDeployment(ctx, deploymentID)
}

// InspectService gets detailed information about a specific service
func (s *Service) InspectService(ctx context.Context, deploymentID string, serviceName string) (*ServiceInspection, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Inspect service using the specific manager
	return manager.InspectService(ctx, deploymentID, serviceName)
}

// GetDeploymentMetrics gets metrics for a deployment
func (s *Service) GetDeploymentMetrics(ctx context.Context, deploymentID string, duration time.Duration) (*DeploymentMetrics, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Get metrics using the specific manager
	return manager.GetDeploymentMetrics(ctx, deploymentID, duration)
}

// GetServiceMetrics gets metrics for a specific service
func (s *Service) GetServiceMetrics(ctx context.Context, deploymentID string, serviceName string, duration time.Duration) (*ServiceMetrics, error) {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return nil, err
	}

	// Get service metrics using the specific manager
	return manager.GetServiceMetrics(ctx, deploymentID, serviceName, duration)
}

// UpdateDeploymentImage updates the image of a service in a deployment
func (s *Service) UpdateDeploymentImage(ctx context.Context, deploymentID string, serviceName string, image string) error {
	// Get deployment to determine its type
	deployment, err := s.GetDeployment(ctx, deploymentID)
	if err != nil {
		return err
	}

	// Get the appropriate deployment manager
	manager, err := s.GetDeploymentManager(deployment.Type)
	if err != nil {
		return err
	}

	// Update image using the specific manager
	return manager.UpdateDeploymentImage(ctx, deploymentID, serviceName, image)
}

// GetSupportedDeploymentTypes returns the list of supported deployment types
func (s *Service) GetSupportedDeploymentTypes() []DeploymentType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	types := make([]DeploymentType, 0, len(s.deploymentManagers))
	for deploymentType := range s.deploymentManagers {
		types = append(types, deploymentType)
	}

	return types
}

// ValidateDeploymentConfig validates configuration for a deployment type
func (s *Service) ValidateDeploymentConfig(ctx context.Context, deploymentType DeploymentType, config *DeploymentConfig) error {
	if s.factory == nil {
		return fmt.Errorf("deployment factory not initialized")
	}

	return s.factory.ValidateConfig(ctx, deploymentType, config)
}

// Helper methods

func (s *Service) stopDeploymentManager(ctx context.Context, manager DeploymentManager) error {
	// This is a placeholder - actual implementation would depend on the manager interface
	// For now, we'll just log that we're stopping the manager
	s.logger.WithField("deployment_type", manager.GetDeploymentType()).Debug("Stopping deployment manager")
	return nil
}
