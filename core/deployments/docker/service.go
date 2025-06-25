package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// Service represents the main deployment service
type Service struct {
	mu       sync.RWMutex
	cfg      *config.C
	logger   *logrus.Logger
	stopChan chan struct{}

	// Managers
	tenantManager     TenantManager
	deploymentManager DeploymentManager
	manifestManager   ManifestManager
	portManager       PortManager
	networkManager    NetworkManager
	resourceManager   ResourceManager
	eventManager      EventManager
	composeExecutor   ComposeExecutor
	storageManager    StorageManager

	// Configuration
	workDir           string
	portPool          *PortPool
	tenantNetworks    map[string]string      // tenantID -> networkName
	activeDeployments map[string]*Deployment // deploymentID -> deployment
}

// NewService creates a new deployment service
func NewService(cfg *config.C, logger *logrus.Logger) *Service {
	return &Service{
		cfg:               cfg,
		logger:            logger,
		stopChan:          make(chan struct{}),
		tenantNetworks:    make(map[string]string),
		activeDeployments: make(map[string]*Deployment),
	}
}

// SetManagers sets the managers for the service
func (s *Service) SetManagers(
	tenantManager TenantManager,
	deploymentManager DeploymentManager,
	manifestManager ManifestManager,
	portManager PortManager,
	networkManager NetworkManager,
	resourceManager ResourceManager,
	eventManager EventManager,
	composeExecutor ComposeExecutor,
	storageManager StorageManager,
) {
	s.tenantManager = tenantManager
	s.deploymentManager = deploymentManager
	s.manifestManager = manifestManager
	s.portManager = portManager
	s.networkManager = networkManager
	s.resourceManager = resourceManager
	s.eventManager = eventManager
	s.composeExecutor = composeExecutor
	s.storageManager = storageManager
}

// Start starts the deployment service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting deployment service...")

	// Initialize work directory
	s.workDir = s.cfg.GetString("deployment.work_dir", "./deployments")
	if err := os.MkdirAll(s.workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Initialize port pool
	s.portPool = &PortPool{
		StartPort: s.cfg.GetInt("deployment.port_pool.start", 10000),
		EndPort:   s.cfg.GetInt("deployment.port_pool.end", 20000),
		UsedPorts: make(map[int]string),
	}

	// Load existing deployments
	if err := s.loadExistingDeployments(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to load existing deployments")
	}

	// Start background tasks
	go s.startBackgroundTasks(ctx)

	s.logger.Info("Deployment service started successfully")
	return nil
}

// Stop stops the deployment service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping deployment service...")

	close(s.stopChan)

	// Stop all active deployments
	s.mu.Lock()
	for deploymentID, deployment := range s.activeDeployments {
		if deployment.Status == string(DeploymentStatusRunning) {
			if err := s.stopDeploymentInternal(ctx, deploymentID); err != nil {
				s.logger.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to stop deployment")
			}
		}
	}
	s.mu.Unlock()

	s.logger.Info("Deployment service stopped successfully")
	return nil
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(ctx context.Context, tenant *Tenant) error {
	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}

	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	// Create tenant directory
	tenantDir := filepath.Join(s.workDir, tenant.ID)
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Create tenant network
	if err := s.networkManager.EnsureNetwork(ctx, tenant.ID); err != nil {
		return fmt.Errorf("failed to create tenant network: %w", err)
	}

	// Save tenant
	if err := s.storageManager.SaveTenant(ctx, tenant); err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	s.logger.WithField("tenant_id", tenant.ID).Info("Tenant created successfully")
	return nil
}

// CreateDeployment creates a new deployment
func (s *Service) CreateDeployment(ctx context.Context, deployment *Deployment) error {
	if deployment.ID == "" {
		deployment.ID = uuid.New().String()
	}

	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()
	deployment.Status = string(DeploymentStatusPending)

	// Validate manifest
	if err := s.manifestManager.ValidateManifest(ctx, &deployment.Manifest); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Process manifest for tenant
	processedManifest, err := s.manifestManager.ProcessManifest(ctx, &deployment.Manifest, deployment.TenantID)
	if err != nil {
		return fmt.Errorf("failed to process manifest: %w", err)
	}
	deployment.Manifest = *processedManifest

	// Check resource availability
	if err := s.resourceManager.CheckResourceAvailability(ctx, deployment.TenantID, &deployment.Manifest); err != nil {
		return fmt.Errorf("insufficient resources: %w", err)
	}

	// Allocate ports
	if err := s.allocatePortsForDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to allocate ports: %w", err)
	}

	// Create deployment directory
	deploymentDir := filepath.Join(s.workDir, deployment.TenantID, deployment.ID)
	if err := os.MkdirAll(deploymentDir, 0755); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Save deployment
	if err := s.storageManager.SaveDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to save deployment: %w", err)
	}

	// Add to active deployments
	s.mu.Lock()
	s.activeDeployments[deployment.ID] = deployment
	s.mu.Unlock()

	// Emit event
	s.eventManager.EmitEvent(ctx, &DeploymentEvent{
		ID:           uuid.New().String(),
		DeploymentID: deployment.ID,
		TenantID:     deployment.TenantID,
		Type:         "deployment_created",
		Message:      "Deployment created successfully",
		Timestamp:    time.Now(),
	})

	s.logger.WithField("deployment_id", deployment.ID).Info("Deployment created successfully")
	return nil
}

// StartDeployment starts a deployment
func (s *Service) StartDeployment(ctx context.Context, deploymentID string) error {
	s.mu.Lock()
	deployment, exists := s.activeDeployments[deploymentID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	if deployment.Status == string(DeploymentStatusRunning) {
		return fmt.Errorf("deployment is already running")
	}

	// Allocate resources
	if err := s.resourceManager.AllocateResources(ctx, deployment.TenantID, &deployment.Manifest); err != nil {
		return fmt.Errorf("failed to allocate resources: %w", err)
	}

	// Ensure network exists
	if err := s.networkManager.EnsureNetwork(ctx, deployment.TenantID); err != nil {
		return fmt.Errorf("failed to ensure network: %w", err)
	}

	// Start deployment
	deploymentDir := filepath.Join(s.workDir, deployment.TenantID, deployment.ID)
	if err := s.composeExecutor.Up(ctx, deploymentID, &deployment.Manifest, deploymentDir); err != nil {
		return fmt.Errorf("failed to start deployment: %w", err)
	}

	// Update status
	deployment.Status = string(DeploymentStatusRunning)
	deployment.UpdatedAt = time.Now()

	if err := s.storageManager.SaveDeployment(ctx, deployment); err != nil {
		s.logger.WithError(err).Error("Failed to save deployment status")
	}

	// Emit event
	s.eventManager.EmitEvent(ctx, &DeploymentEvent{
		ID:           uuid.New().String(),
		DeploymentID: deploymentID,
		TenantID:     deployment.TenantID,
		Type:         "deployment_started",
		Message:      "Deployment started successfully",
		Timestamp:    time.Now(),
	})

	s.logger.WithField("deployment_id", deploymentID).Info("Deployment started successfully")
	return nil
}

// StopDeployment stops a deployment
func (s *Service) StopDeployment(ctx context.Context, deploymentID string) error {
	return s.stopDeploymentInternal(ctx, deploymentID)
}

// DeleteDeployment deletes a deployment
func (s *Service) DeleteDeployment(ctx context.Context, deploymentID string) error {
	s.mu.Lock()
	deployment, exists := s.activeDeployments[deploymentID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	// Stop deployment if running
	if deployment.Status == string(DeploymentStatusRunning) {
		if err := s.stopDeploymentInternal(ctx, deploymentID); err != nil {
			s.logger.WithError(err).Error("Failed to stop deployment before deletion")
		}
	}

	// Release ports
	if err := s.releasePortsForDeployment(ctx, deployment); err != nil {
		s.logger.WithError(err).Error("Failed to release ports")
	}

	// Release resources
	if err := s.resourceManager.ReleaseResources(ctx, deployment.TenantID, &deployment.Manifest); err != nil {
		s.logger.WithError(err).Error("Failed to release resources")
	}

	// Remove from active deployments
	s.mu.Lock()
	delete(s.activeDeployments, deploymentID)
	s.mu.Unlock()

	// Delete deployment data
	if err := s.storageManager.DeleteDeployment(ctx, deploymentID); err != nil {
		return fmt.Errorf("failed to delete deployment data: %w", err)
	}

	// Remove deployment directory
	deploymentDir := filepath.Join(s.workDir, deployment.TenantID, deploymentID)
	if err := os.RemoveAll(deploymentDir); err != nil {
		s.logger.WithError(err).Error("Failed to remove deployment directory")
	}

	// Emit event
	s.eventManager.EmitEvent(ctx, &DeploymentEvent{
		ID:           uuid.New().String(),
		DeploymentID: deploymentID,
		TenantID:     deployment.TenantID,
		Type:         "deployment_deleted",
		Message:      "Deployment deleted successfully",
		Timestamp:    time.Now(),
	})

	s.logger.WithField("deployment_id", deploymentID).Info("Deployment deleted successfully")
	return nil
}

// GetDeployment gets a deployment by ID
func (s *Service) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	s.mu.RLock()
	deployment, exists := s.activeDeployments[deploymentID]
	s.mu.RUnlock()

	if exists {
		return deployment, nil
	}

	// Try to load from storage
	return s.storageManager.LoadDeployment(ctx, deploymentID)
}

// ListDeployments lists deployments for a tenant
func (s *Service) ListDeployments(ctx context.Context, tenantID string) ([]*Deployment, error) {
	s.mu.RLock()
	var deployments []*Deployment
	for _, deployment := range s.activeDeployments {
		if deployment.TenantID == tenantID {
			deployments = append(deployments, deployment)
		}
	}
	s.mu.RUnlock()

	return deployments, nil
}

// GetTenantResourceUsage gets resource usage for a tenant
func (s *Service) GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error) {
	return s.resourceManager.GetTenantResourceUsage(ctx, tenantID)
}

// Helper methods

func (s *Service) stopDeploymentInternal(ctx context.Context, deploymentID string) error {
	s.mu.Lock()
	deployment, exists := s.activeDeployments[deploymentID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	// Stop deployment
	deploymentDir := filepath.Join(s.workDir, deployment.TenantID, deploymentID)
	if err := s.composeExecutor.Down(ctx, deploymentID, deploymentDir); err != nil {
		return fmt.Errorf("failed to stop deployment: %w", err)
	}

	// Update status
	deployment.Status = string(DeploymentStatusStopped)
	deployment.UpdatedAt = time.Now()

	if err := s.storageManager.SaveDeployment(ctx, deployment); err != nil {
		s.logger.WithError(err).Error("Failed to save deployment status")
	}

	// Emit event
	s.eventManager.EmitEvent(ctx, &DeploymentEvent{
		ID:           uuid.New().String(),
		DeploymentID: deploymentID,
		TenantID:     deployment.TenantID,
		Type:         "deployment_stopped",
		Message:      "Deployment stopped successfully",
		Timestamp:    time.Now(),
	})

	return nil
}

func (s *Service) allocatePortsForDeployment(ctx context.Context, deployment *Deployment) error {
	for serviceName, service := range deployment.Manifest.Services {
		for i, portMapping := range service.Ports {
			if portMapping.HostPort == 0 {
				// Allocate a port
				allocatedPort, err := s.portManager.AllocatePort(ctx, deployment.TenantID, 0)
				if err != nil {
					return fmt.Errorf("failed to allocate port for service %s: %w", serviceName, err)
				}

				// Update the port mapping
				deployment.Manifest.Services[serviceName].Ports[i].HostPort = allocatedPort
			}
		}
	}
	return nil
}

func (s *Service) releasePortsForDeployment(ctx context.Context, deployment *Deployment) error {
	for _, service := range deployment.Manifest.Services {
		for _, portMapping := range service.Ports {
			if portMapping.HostPort != 0 {
				if err := s.portManager.ReleasePort(ctx, portMapping.HostPort); err != nil {
					s.logger.WithError(err).WithField("port", portMapping.HostPort).Error("Failed to release port")
				}
			}
		}
	}
	return nil
}

func (s *Service) loadExistingDeployments(ctx context.Context) error {
	deployments, err := s.storageManager.ListDeployments(ctx)
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		s.activeDeployments[deployment.ID] = deployment
	}

	return nil
}

func (s *Service) startBackgroundTasks(ctx context.Context) {
	// Start resource monitoring
	go s.monitorResources(ctx)

	// Start deployment health checks
	go s.healthCheckDeployments(ctx)
}

func (s *Service) monitorResources(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.updateResourceUsage(ctx)
		}
	}
}

func (s *Service) updateResourceUsage(ctx context.Context) {
	// Update resource usage for all tenants
	tenants, err := s.storageManager.ListTenants(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list tenants for resource monitoring")
		return
	}

	for _, tenant := range tenants {
		usage, err := s.resourceManager.GetTenantResourceUsage(ctx, tenant.ID)
		if err != nil {
			s.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to get tenant resource usage")
			continue
		}

		if err := s.resourceManager.UpdateResourceUsage(ctx, tenant.ID, usage); err != nil {
			s.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to update tenant resource usage")
		}
	}
}

func (s *Service) healthCheckDeployments(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkDeploymentHealth(ctx)
		}
	}
}

func (s *Service) checkDeploymentHealth(ctx context.Context) {
	s.mu.RLock()
	deployments := make([]*Deployment, 0, len(s.activeDeployments))
	for _, deployment := range s.activeDeployments {
		if deployment.Status == string(DeploymentStatusRunning) {
			deployments = append(deployments, deployment)
		}
	}
	s.mu.RUnlock()

	for _, deployment := range deployments {
		deploymentDir := filepath.Join(s.workDir, deployment.TenantID, deployment.ID)
		status, err := s.composeExecutor.Status(ctx, deployment.ID, deploymentDir)
		if err != nil {
			s.logger.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to check deployment status")
			continue
		}

		// Check if any service is not running
		allRunning := true
		for serviceName, serviceStatus := range status {
			if serviceStatus != "running" {
				allRunning = false
				s.logger.WithFields(logrus.Fields{
					"deployment_id": deployment.ID,
					"service_name":  serviceName,
					"status":        serviceStatus,
				}).Warn("Service not running")
			}
		}

		if !allRunning {
			// Emit health check event
			s.eventManager.EmitEvent(ctx, &DeploymentEvent{
				ID:           uuid.New().String(),
				DeploymentID: deployment.ID,
				TenantID:     deployment.TenantID,
				Type:         "health_check_failed",
				Message:      "Deployment health check failed",
				Timestamp:    time.Now(),
				Data: map[string]interface{}{
					"status": status,
				},
			})
		}
	}
}
