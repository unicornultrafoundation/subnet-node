package deployment

import (
	"context"
	"io"
)

// TenantManager defines the interface for tenant management
type TenantManager interface {
	// CreateTenant creates a new tenant
	CreateTenant(ctx context.Context, tenant *Tenant) error

	// GetTenant retrieves a tenant by ID
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)

	// ListTenants lists all tenants
	ListTenants(ctx context.Context) ([]*Tenant, error)

	// UpdateTenant updates an existing tenant
	UpdateTenant(ctx context.Context, tenant *Tenant) error

	// DeleteTenant deletes a tenant and all its resources
	DeleteTenant(ctx context.Context, tenantID string) error

	// GetTenantResourceUsage gets resource usage for a tenant
	GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error)
}

// DeploymentManager defines the interface for deployment management
type DeploymentManager interface {
	// CreateDeployment creates a new deployment
	CreateDeployment(ctx context.Context, deployment *Deployment) error

	// GetDeployment retrieves a deployment by ID
	GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)

	// ListDeployments lists deployments for a tenant
	ListDeployments(ctx context.Context, tenantID string) ([]*Deployment, error)

	// UpdateDeployment updates an existing deployment
	UpdateDeployment(ctx context.Context, deployment *Deployment) error

	// DeleteDeployment deletes a deployment
	DeleteDeployment(ctx context.Context, deploymentID string) error

	// StartDeployment starts a deployment
	StartDeployment(ctx context.Context, deploymentID string) error

	// StopDeployment stops a deployment
	StopDeployment(ctx context.Context, deploymentID string) error

	// RestartDeployment restarts a deployment
	RestartDeployment(ctx context.Context, deploymentID string) error

	// GetDeploymentStatus gets the current status of a deployment
	GetDeploymentStatus(ctx context.Context, deploymentID string) (DeploymentStatus, error)

	// GetDeploymentLogs gets logs for a deployment
	GetDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, tail int) (io.ReadCloser, error)
}

// ManifestManager defines the interface for manifest management
type ManifestManager interface {
	// ValidateManifest validates a compose manifest
	ValidateManifest(ctx context.Context, manifest *ComposeManifest) error

	// ProcessManifest processes and transforms a manifest for deployment
	ProcessManifest(ctx context.Context, manifest *ComposeManifest, tenantID string) (*ComposeManifest, error)

	// GenerateManifest generates a manifest from a template
	GenerateManifest(ctx context.Context, template string, params map[string]interface{}) (*ComposeManifest, error)

	// MergeManifests merges multiple manifests
	MergeManifests(ctx context.Context, manifests ...*ComposeManifest) (*ComposeManifest, error)

	// SerializeManifest serializes a manifest to YAML
	SerializeManifest(ctx context.Context, manifest *ComposeManifest) ([]byte, error)

	// DeserializeManifest deserializes a manifest from YAML
	DeserializeManifest(ctx context.Context, data []byte) (*ComposeManifest, error)
}

// PortManager defines the interface for port management
type PortManager interface {
	// AllocatePort allocates a port for a tenant
	AllocatePort(ctx context.Context, tenantID string, preferredPort int) (int, error)

	// ReleasePort releases a port
	ReleasePort(ctx context.Context, port int) error

	// GetPortPool gets the current port pool status
	GetPortPool(ctx context.Context) *PortPool

	// ReservePortRange reserves a range of ports for a tenant
	ReservePortRange(ctx context.Context, tenantID string, startPort, endPort int) error

	// ReleasePortRange releases a range of ports
	ReleasePortRange(ctx context.Context, startPort, endPort int) error

	// GetTenantPorts gets all ports allocated to a tenant
	GetTenantPorts(ctx context.Context, tenantID string) ([]int, error)
}

// NetworkManager defines the interface for network management
type NetworkManager interface {
	// CreateNetwork creates a network for a tenant
	CreateNetwork(ctx context.Context, tenantID string, config *NetworkConfig) error

	// GetNetwork gets a network by name
	GetNetwork(ctx context.Context, networkName string) (*NetworkConfig, error)

	// ListNetworks lists all networks for a tenant
	ListNetworks(ctx context.Context, tenantID string) ([]*NetworkConfig, error)

	// DeleteNetwork deletes a network
	DeleteNetwork(ctx context.Context, networkName string) error

	// EnsureNetwork ensures a network exists for a tenant
	EnsureNetwork(ctx context.Context, tenantID string) error
}

// ResourceManager defines the interface for resource management
type ResourceManager interface {
	// CheckResourceAvailability checks if resources are available for a deployment
	CheckResourceAvailability(ctx context.Context, tenantID string, manifest *ComposeManifest) error

	// AllocateResources allocates resources for a deployment
	AllocateResources(ctx context.Context, tenantID string, manifest *ComposeManifest) error

	// ReleaseResources releases resources for a deployment
	ReleaseResources(ctx context.Context, tenantID string, manifest *ComposeManifest) error

	// GetSystemResources gets the total system resources
	GetSystemResources(ctx context.Context) (*TenantResourceUsage, error)

	// GetAvailableResources gets available resources
	GetAvailableResources(ctx context.Context) (*TenantResourceUsage, error)

	// GetTenantResourceUsage gets resource usage for a specific tenant
	GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error)

	// UpdateResourceUsage updates resource usage for a tenant
	UpdateResourceUsage(ctx context.Context, tenantID string, usage *TenantResourceUsage) error
}

// EventManager defines the interface for event management
type EventManager interface {
	// EmitEvent emits a deployment event
	EmitEvent(ctx context.Context, event *DeploymentEvent) error

	// GetEvents gets events for a deployment
	GetEvents(ctx context.Context, deploymentID string, limit int) ([]*DeploymentEvent, error)

	// SubscribeToEvents subscribes to deployment events
	SubscribeToEvents(ctx context.Context, deploymentID string) (<-chan *DeploymentEvent, error)

	// UnsubscribeFromEvents unsubscribes from deployment events
	UnsubscribeFromEvents(ctx context.Context, deploymentID string) error
}

// ComposeExecutor defines the interface for docker-compose execution
type ComposeExecutor interface {
	// Up starts a deployment
	Up(ctx context.Context, deploymentID string, manifest *ComposeManifest, workDir string) error

	// Down stops a deployment
	Down(ctx context.Context, deploymentID string, workDir string) error

	// Restart restarts a deployment
	Restart(ctx context.Context, deploymentID string, workDir string) error

	// Status gets the status of a deployment
	Status(ctx context.Context, deploymentID string, workDir string) (map[string]string, error)

	// Logs gets logs for a deployment
	Logs(ctx context.Context, deploymentID string, serviceName string, workDir string, tail int) (io.ReadCloser, error)

	// Scale scales a service
	Scale(ctx context.Context, deploymentID string, serviceName string, replicas int, workDir string) error

	// Pull pulls images for a deployment
	Pull(ctx context.Context, deploymentID string, manifest *ComposeManifest, workDir string) error
}

// StorageManager defines the interface for storage management
type StorageManager interface {
	// SaveDeployment saves deployment data
	SaveDeployment(ctx context.Context, deployment *Deployment) error

	// LoadDeployment loads deployment data
	LoadDeployment(ctx context.Context, deploymentID string) (*Deployment, error)

	// DeleteDeployment deletes deployment data
	DeleteDeployment(ctx context.Context, deploymentID string) error

	// ListDeployments lists all deployments
	ListDeployments(ctx context.Context) ([]*Deployment, error)

	// SaveTenant saves tenant data
	SaveTenant(ctx context.Context, tenant *Tenant) error

	// LoadTenant loads tenant data
	LoadTenant(ctx context.Context, tenantID string) (*Tenant, error)

	// DeleteTenant deletes tenant data
	DeleteTenant(ctx context.Context, tenantID string) error

	// ListTenants lists all tenants
	ListTenants(ctx context.Context) ([]*Tenant, error)
}
