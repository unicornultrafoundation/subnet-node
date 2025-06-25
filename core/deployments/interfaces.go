package deployments

import (
	"context"
	"io"
	"time"
)

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

	// StreamDeploymentLogs streams logs for a deployment in real-time
	StreamDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, follow bool) (<-chan LogEntry, error)

	// ExecConsole executes a command in a deployment container
	ExecConsole(ctx context.Context, deploymentID string, serviceName string, command []string, tty bool) (ExecSession, error)

	// InspectDeployment gets detailed information about a deployment
	InspectDeployment(ctx context.Context, deploymentID string) (*DeploymentInspection, error)

	// InspectService gets detailed information about a specific service
	InspectService(ctx context.Context, deploymentID string, serviceName string) (*ServiceInspection, error)

	// GetDeploymentMetrics gets metrics for a deployment
	GetDeploymentMetrics(ctx context.Context, deploymentID string, duration time.Duration) (*DeploymentMetrics, error)

	// GetServiceMetrics gets metrics for a specific service
	GetServiceMetrics(ctx context.Context, deploymentID string, serviceName string, duration time.Duration) (*ServiceMetrics, error)

	// ScaleDeployment scales a deployment
	ScaleDeployment(ctx context.Context, deploymentID string, serviceName string, replicas int) error

	// UpdateDeploymentImage updates the image of a service in a deployment
	UpdateDeploymentImage(ctx context.Context, deploymentID string, serviceName string, image string) error

	// GetDeploymentType returns the type of deployment this manager handles
	GetDeploymentType() DeploymentType
}

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

// ManifestManager defines the interface for manifest management
type ManifestManager interface {
	// ValidateManifest validates a manifest
	ValidateManifest(ctx context.Context, manifest interface{}) error

	// ProcessManifest processes and transforms a manifest for deployment
	ProcessManifest(ctx context.Context, manifest interface{}, tenantID string) (interface{}, error)

	// GenerateManifest generates a manifest from a template
	GenerateManifest(ctx context.Context, template string, params map[string]interface{}) (interface{}, error)

	// MergeManifests merges multiple manifests
	MergeManifests(ctx context.Context, manifests ...interface{}) (interface{}, error)

	// SerializeManifest serializes a manifest
	SerializeManifest(ctx context.Context, manifest interface{}) ([]byte, error)

	// DeserializeManifest deserializes a manifest
	DeserializeManifest(ctx context.Context, data []byte) (interface{}, error)

	// GetManifestType returns the type of manifest this manager handles
	GetManifestType() DeploymentType
}

// PortManager defines the interface for port management
type PortManager interface {
	// AllocatePort allocates a port for a tenant
	AllocatePort(ctx context.Context, tenantID string, preferredPort int) (int, error)

	// ReleasePort releases a port
	ReleasePort(ctx context.Context, port int) error

	// GetPortPool gets the current port pool status
	GetPortPool(ctx context.Context) interface{}

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
	CheckResourceAvailability(ctx context.Context, tenantID string, manifest interface{}) error

	// AllocateResources allocates resources for a deployment
	AllocateResources(ctx context.Context, tenantID string, manifest interface{}) error

	// ReleaseResources releases resources for a deployment
	ReleaseResources(ctx context.Context, tenantID string, manifest interface{}) error

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

// DeploymentFactory defines the interface for creating deployment managers
type DeploymentFactory interface {
	// CreateManager creates a deployment manager for a specific type
	CreateManager(ctx context.Context, deploymentType DeploymentType, config *DeploymentConfig) (DeploymentManager, error)

	// GetSupportedTypes returns the list of supported deployment types
	GetSupportedTypes() []DeploymentType

	// ValidateConfig validates configuration for a deployment type
	ValidateConfig(ctx context.Context, deploymentType DeploymentType, config *DeploymentConfig) error
}

// ServiceManager defines the interface for the main deployment service
type ServiceManager interface {
	// Start starts the deployment service
	Start(ctx context.Context) error

	// Stop stops the deployment service
	Stop(ctx context.Context) error

	// CreateTenant creates a new tenant
	CreateTenant(ctx context.Context, tenant *Tenant) error

	// CreateDeployment creates a new deployment
	CreateDeployment(ctx context.Context, deployment *Deployment) error

	// GetDeployment gets a deployment by ID
	GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)

	// ListDeployments lists deployments for a tenant
	ListDeployments(ctx context.Context, tenantID string) ([]*Deployment, error)

	// StartDeployment starts a deployment
	StartDeployment(ctx context.Context, deploymentID string) error

	// StopDeployment stops a deployment
	StopDeployment(ctx context.Context, deploymentID string) error

	// DeleteDeployment deletes a deployment
	DeleteDeployment(ctx context.Context, deploymentID string) error

	// GetTenantResourceUsage gets resource usage for a tenant
	GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error)

	// GetDeploymentLogs gets logs for a deployment
	GetDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, tail int) (io.ReadCloser, error)

	// StreamDeploymentLogs streams logs for a deployment in real-time
	StreamDeploymentLogs(ctx context.Context, deploymentID string, serviceName string, follow bool) (<-chan LogEntry, error)

	// ExecConsole executes a command in a deployment container
	ExecConsole(ctx context.Context, deploymentID string, serviceName string, command []string, tty bool) (ExecSession, error)

	// InspectDeployment gets detailed information about a deployment
	InspectDeployment(ctx context.Context, deploymentID string) (*DeploymentInspection, error)

	// InspectService gets detailed information about a specific service
	InspectService(ctx context.Context, deploymentID string, serviceName string) (*ServiceInspection, error)

	// GetDeploymentMetrics gets metrics for a deployment
	GetDeploymentMetrics(ctx context.Context, deploymentID string, duration time.Duration) (*DeploymentMetrics, error)

	// GetServiceMetrics gets metrics for a specific service
	GetServiceMetrics(ctx context.Context, deploymentID string, serviceName string, duration time.Duration) (*ServiceMetrics, error)

	// ScaleDeployment scales a deployment
	ScaleDeployment(ctx context.Context, deploymentID string, serviceName string, replicas int) error

	// UpdateDeploymentImage updates the image of a service in a deployment
	UpdateDeploymentImage(ctx context.Context, deploymentID string, serviceName string, image string) error
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// ExecSession represents an execution session
type ExecSession interface {
	// Execute executes a command and returns the result
	Execute(ctx context.Context, command []string) (ExecResult, error)

	// ExecuteInteractive executes a command in interactive mode
	ExecuteInteractive(ctx context.Context, command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error

	// Close closes the session
	Close() error
}

// ExecResult represents the result of an executed command
type ExecResult struct {
	ExitCode int           `json:"exit_code"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
}

// DeploymentInspection represents detailed deployment information
type DeploymentInspection struct {
	ID        string                  `json:"id"`
	TenantID  string                  `json:"tenant_id"`
	Name      string                  `json:"name"`
	Status    DeploymentStatus        `json:"status"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	Services  map[string]*ServiceInfo `json:"services"`
	Networks  []*NetworkInfo          `json:"networks"`
	Volumes   []*VolumeInfo           `json:"volumes"`
	Resources *ResourceUsage          `json:"resources"`
	Health    *HealthStatus           `json:"health"`
	Config    map[string]interface{}  `json:"config"`
}

// ServiceInfo represents information about a service
type ServiceInfo struct {
	Name        string         `json:"name"`
	Image       string         `json:"image"`
	Status      string         `json:"status"`
	Replicas    int            `json:"replicas"`
	Ports       []*PortInfo    `json:"ports"`
	Volumes     []*VolumeInfo  `json:"volumes"`
	Environment []*EnvVar      `json:"environment"`
	Resources   *ResourceUsage `json:"resources"`
	Health      *HealthStatus  `json:"health"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ServiceInspection represents detailed service information
type ServiceInspection struct {
	*ServiceInfo
	Logs      []*LogEntry            `json:"logs"`
	Metrics   *ServiceMetrics        `json:"metrics"`
	Processes []*ProcessInfo         `json:"processes"`
	Network   *NetworkInfo           `json:"network"`
	Config    map[string]interface{} `json:"config"`
}

// PortInfo represents port information
type PortInfo struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"host_ip"`
}

// VolumeInfo represents volume information
type VolumeInfo struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ReadOnly    bool   `json:"read_only"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  int64     `json:"memory_usage"`
	DiskUsage    int64     `json:"disk_usage"`
	NetworkUsage int64     `json:"network_usage"`
	Timestamp    time.Time `json:"timestamp"`
}

// HealthStatus represents health status information
type HealthStatus struct {
	Status    string               `json:"status"`
	Message   string               `json:"message"`
	LastCheck time.Time            `json:"last_check"`
	Checks    []*HealthCheckStatus `json:"checks"`
}

// HealthCheckStatus represents a health check status
type HealthCheckStatus struct {
	Type      string        `json:"type"`
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	LastCheck time.Time     `json:"last_check"`
	Duration  time.Duration `json:"duration"`
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	Internal   bool              `json:"internal"`
	EnableIPv6 bool              `json:"enable_ipv6"`
	Labels     map[string]string `json:"labels"`
}

// ProcessInfo represents process information
type ProcessInfo struct {
	PID         int       `json:"pid"`
	Command     string    `json:"command"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage int64     `json:"memory_usage"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
}

// DeploymentMetrics represents deployment metrics
type DeploymentMetrics struct {
	DeploymentID string                     `json:"deployment_id"`
	Timestamp    time.Time                  `json:"timestamp"`
	Duration     time.Duration              `json:"duration"`
	Services     map[string]*ServiceMetrics `json:"services"`
	Total        *ResourceUsage             `json:"total"`
}

// ServiceMetrics represents service metrics
type ServiceMetrics struct {
	ServiceName string          `json:"service_name"`
	Timestamp   time.Time       `json:"timestamp"`
	Duration    time.Duration   `json:"duration"`
	Resources   *ResourceUsage  `json:"resources"`
	Requests    *RequestMetrics `json:"requests"`
	Errors      *ErrorMetrics   `json:"errors"`
}

// RequestMetrics represents request metrics
type RequestMetrics struct {
	Total       int64   `json:"total"`
	Successful  int64   `json:"successful"`
	Failed      int64   `json:"failed"`
	AvgResponse float64 `json:"avg_response"`
	MaxResponse float64 `json:"max_response"`
	MinResponse float64 `json:"min_response"`
}

// ErrorMetrics represents error metrics
type ErrorMetrics struct {
	Total     int64            `json:"total"`
	ByType    map[string]int64 `json:"by_type"`
	ByCode    map[int]int64    `json:"by_code"`
	LastError *ErrorInfo       `json:"last_error"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Type      string    `json:"type"`
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Stack     string    `json:"stack"`
}
