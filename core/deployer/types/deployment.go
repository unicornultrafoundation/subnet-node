package types

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
)

// DeploymentStatus represents a deployment status
type DeploymentStatus string

const (
	// DeploymentStatusPending represents a pending deployment
	DeploymentStatusPending DeploymentStatus = "pending"
	// DeploymentStatusRunning represents a running deployment
	DeploymentStatusRunning DeploymentStatus = "running"
	// DeploymentStatusCompleted represents a completed deployment
	DeploymentStatusCompleted DeploymentStatus = "completed"
	// DeploymentStatusFailed represents a failed deployment
	DeploymentStatusFailed DeploymentStatus = "failed"
	// DeploymentStatusTerminated represents a terminated deployment
	DeploymentStatusTerminated DeploymentStatus = "terminated"
)

// ManagedDeployment represents a managed deployment
type ManagedDeployment struct {
	ID           string
	LeaseID      string
	Name         string
	Namespace    string
	Requester    common.Address
	Status       DeploymentStatus
	HealthStatus HealthStatus
	Version      int64
	Error        error
	Manifest     *manifest.Manifest
	SDL          *manifest.SDL
	Resources    manifest.ResourceRequirements
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastHealth   time.Time
}

// EnhancedDeploymentStatus represents an enhanced deployment status
type EnhancedDeploymentStatus struct {
	ID            string
	LeaseID       string
	Name          string
	Namespace     string
	Requester     common.Address
	Status        DeploymentStatus
	HealthStatus  HealthStatus
	Version       int64
	Error         error
	Manifest      *manifest.SDL
	Resources     manifest.ResourceRequirements
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ErrorCount    int64
	WarningCount  int64
	CriticalCount int64
	ResponseTime  time.Duration
	ResourceUsage map[string]float64
}

// ServiceStatus represents a service status
type ServiceStatus struct {
	Name      string
	Namespace string
	Status    DeploymentStatus
	Error     error
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Service represents a service in the SDL
type DeploymentService struct {
	Name      string
	Image     string
	Command   []string
	Args      []string
	Env       []string
	Ports     []int32
	Resources manifest.ResourceRequirements
	Hostname  string // Optional custom hostname for the service
}

// DeploymentManagerInterface defines the interface for deployment management
type DeploymentManagerInterface interface {
	Start(ctx context.Context) error
	Stop()
	CreateDeployment(ctx context.Context, id string, requester common.Address, sdl *manifest.SDL) error
	StopDeployment(ctx context.Context, id string) error
	GetDeployment(id string) (*ManagedDeployment, error)
	UpdateDeployment(ctx context.Context, dep *ManagedDeployment) error
	ListDeployments() ([]*ManagedDeployment, error)
	IsHealthy() bool
	GetErrorCount() int64
	GetWarningCount() int64
	GetCriticalCount() int64
	GetResponseTime() time.Duration
	GetResourceUsage() map[string]float64
	GetDeploymentVersion(id string) (int64, error)
}

// PodStatus represents the status of a pod
type PodStatus struct {
	Name   string
	Status string
	Ready  bool
}

// DeploymentManagerConfig defines the configuration for a deployment manager
type DeploymentManagerConfig struct {
	DeploymentDir string
	MaxRetries    int
}
