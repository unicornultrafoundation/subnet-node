package managers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	deployment "github.com/unicornultrafoundation/subnet-node/core/deployments/docker"
)

// TenantManagerImpl implements TenantManager
type TenantManagerImpl struct {
	mu          sync.RWMutex
	logger      *logrus.Logger
	storage     deployment.StorageManager
	resourceMgr deployment.ResourceManager
	tenants     map[string]*deployment.Tenant
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(logger *logrus.Logger, storage deployment.StorageManager, resourceMgr deployment.ResourceManager) deployment.TenantManager {
	return &TenantManagerImpl{
		logger:      logger,
		storage:     storage,
		resourceMgr: resourceMgr,
		tenants:     make(map[string]*deployment.Tenant),
	}
}

// CreateTenant creates a new tenant
func (tm *TenantManagerImpl) CreateTenant(ctx context.Context, tenant *deployment.Tenant) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}

	// Check if tenant already exists
	if _, exists := tm.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant with ID %s already exists", tenant.ID)
	}

	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	// Initialize resource usage
	usage := &deployment.TenantResourceUsage{
		TenantID:       tenant.ID,
		CPUUsage:       0,
		MemoryUsage:    0,
		DiskUsage:      0,
		NetworkUsage:   0,
		PortCount:      0,
		ContainerCount: 0,
	}

	if err := tm.resourceMgr.UpdateResourceUsage(ctx, tenant.ID, usage); err != nil {
		return fmt.Errorf("failed to initialize tenant resource usage: %w", err)
	}

	// Save tenant
	if err := tm.storage.SaveTenant(ctx, tenant); err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	tm.tenants[tenant.ID] = tenant

	tm.logger.WithField("tenant_id", tenant.ID).Info("Tenant created successfully")
	return nil
}

// GetTenant retrieves a tenant by ID
func (tm *TenantManagerImpl) GetTenant(ctx context.Context, tenantID string) (*deployment.Tenant, error) {
	tm.mu.RLock()
	tenant, exists := tm.tenants[tenantID]
	tm.mu.RUnlock()

	if exists {
		return tenant, nil
	}

	// Try to load from storage
	tenant, err := tm.storage.LoadTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Cache the tenant
	tm.mu.Lock()
	tm.tenants[tenantID] = tenant
	tm.mu.Unlock()

	return tenant, nil
}

// ListTenants lists all tenants
func (tm *TenantManagerImpl) ListTenants(ctx context.Context) ([]*deployment.Tenant, error) {
	tm.mu.RLock()
	tenants := make([]*deployment.Tenant, 0, len(tm.tenants))
	for _, tenant := range tm.tenants {
		tenants = append(tenants, tenant)
	}
	tm.mu.RUnlock()

	return tenants, nil
}

// UpdateTenant updates an existing tenant
func (tm *TenantManagerImpl) UpdateTenant(ctx context.Context, tenant *deployment.Tenant) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if tenant exists
	if _, exists := tm.tenants[tenant.ID]; !exists {
		return fmt.Errorf("tenant with ID %s does not exist", tenant.ID)
	}

	tenant.UpdatedAt = time.Now()

	// Save tenant
	if err := tm.storage.SaveTenant(ctx, tenant); err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	tm.tenants[tenant.ID] = tenant

	tm.logger.WithField("tenant_id", tenant.ID).Info("Tenant updated successfully")
	return nil
}

// DeleteTenant deletes a tenant and all its resources
func (tm *TenantManagerImpl) DeleteTenant(ctx context.Context, tenantID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if tenant exists
	if _, exists := tm.tenants[tenantID]; !exists {
		return fmt.Errorf("tenant with ID %s does not exist", tenantID)
	}

	// Delete tenant from storage
	if err := tm.storage.DeleteTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant from storage: %w", err)
	}

	// Remove from cache
	delete(tm.tenants, tenantID)

	tm.logger.WithField("tenant_id", tenantID).Info("Tenant deleted successfully")
	return nil
}

// GetTenantResourceUsage gets resource usage for a tenant
func (tm *TenantManagerImpl) GetTenantResourceUsage(ctx context.Context, tenantID string) (*deployment.TenantResourceUsage, error) {
	return tm.resourceMgr.GetTenantResourceUsage(ctx, tenantID)
}
