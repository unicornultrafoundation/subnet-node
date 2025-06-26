package deployments

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// Factory implements DeploymentFactory
type Factory struct {
	mu              sync.RWMutex
	logger          *logrus.Logger
	managerCreators map[DeploymentType]ManagerCreator
}

// ManagerCreator is a function that creates a deployment manager
type ManagerCreator func(ctx context.Context, config *DeploymentConfig) (DeploymentManager, error)

// NewFactory creates a new deployment factory
func NewFactory(logger *logrus.Logger) *Factory {
	return &Factory{
		logger:          logger,
		managerCreators: make(map[DeploymentType]ManagerCreator),
	}
}

// RegisterManagerCreator registers a manager creator for a deployment type
func (f *Factory) RegisterManagerCreator(deploymentType DeploymentType, creator ManagerCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.managerCreators[deploymentType] = creator
	f.logger.WithField("deployment_type", deploymentType).Debug("Manager creator registered")
}

// CreateManager creates a deployment manager for a specific type
func (f *Factory) CreateManager(ctx context.Context, deploymentType DeploymentType, config *DeploymentConfig) (DeploymentManager, error) {
	f.mu.RLock()
	creator, exists := f.managerCreators[deploymentType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no manager creator registered for deployment type: %s", deploymentType)
	}

	manager, err := creator(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager for deployment type %s: %w", deploymentType, err)
	}

	f.logger.WithField("deployment_type", deploymentType).Debug("Deployment manager created successfully")
	return manager, nil
}

// GetSupportedTypes returns the list of supported deployment types
func (f *Factory) GetSupportedTypes() []DeploymentType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]DeploymentType, 0, len(f.managerCreators))
	for deploymentType := range f.managerCreators {
		types = append(types, deploymentType)
	}

	return types
}

// ValidateConfig validates configuration for a deployment type
func (f *Factory) ValidateConfig(ctx context.Context, deploymentType DeploymentType, config *DeploymentConfig) error {
	// Basic validation
	if config == nil {
		return fmt.Errorf("deployment config cannot be nil")
	}

	if config.Type != deploymentType {
		return fmt.Errorf("deployment type mismatch: expected %s, got %s", deploymentType, config.Type)
	}

	// Check if the deployment type is supported
	f.mu.RLock()
	_, exists := f.managerCreators[deploymentType]
	f.mu.RUnlock()

	if !exists {
		return fmt.Errorf("deployment type %s is not supported", deploymentType)
	}

	// Only support Kubernetes
	if deploymentType != DeploymentTypeKubernetes {
		return fmt.Errorf("only Kubernetes deployment type is supported")
	}
	return f.validateKubernetesConfig(config)
}

// validateKubernetesConfig validates Kubernetes deployment configuration
func (f *Factory) validateKubernetesConfig(config *DeploymentConfig) error {
	// Add Kubernetes-specific validation here
	if config.Config == nil {
		return fmt.Errorf("Kubernetes deployment config cannot be nil")
	}

	// Additional Kubernetes validation logic
	return nil
}
