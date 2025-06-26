package deployments

import (
	"time"
)

// DeploymentType represents the type of deployment
type DeploymentType string

const (
	DeploymentTypeDocker     DeploymentType = "docker"
	DeploymentTypeKubernetes DeploymentType = "kubernetes"
	DeploymentTypeNomad      DeploymentType = "nomad"
	DeploymentTypeTerraform  DeploymentType = "terraform"
)

// Deployment represents a deployment configuration
type Deployment struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Type        DeploymentType    `json:"type" yaml:"type"`
	Manifest    interface{}       `json:"manifest" yaml:"manifest"` // Type-specific manifest
	Status      string            `json:"status" yaml:"status"`
	Owner       string            `json:"owner" yaml:"owner"` // Ethereum address of deployment owner
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "pending"
	DeploymentStatusRunning  DeploymentStatus = "running"
	DeploymentStatusStopped  DeploymentStatus = "stopped"
	DeploymentStatusFailed   DeploymentStatus = "failed"
	DeploymentStatusRemoving DeploymentStatus = "removing"
	DeploymentStatusUpdating DeploymentStatus = "updating"
)

// DeploymentEvent represents an event in the deployment lifecycle
type DeploymentEvent struct {
	ID           string                 `json:"id" yaml:"id"`
	DeploymentID string                 `json:"deployment_id" yaml:"deployment_id"`
	Type         string                 `json:"type" yaml:"type"`
	Message      string                 `json:"message" yaml:"message"`
	Timestamp    time.Time              `json:"timestamp" yaml:"timestamp"`
	Data         map[string]interface{} `json:"data" yaml:"data"`
}

// DeploymentConfig represents configuration for a specific deployment type
type DeploymentConfig struct {
	Type    DeploymentType         `json:"type" yaml:"type"`
	Config  interface{}            `json:"config" yaml:"config"`
	Options map[string]interface{} `json:"options" yaml:"options"`
}

// ResourceLimits represents resource limits for a deployment
type ResourceLimits struct {
	CPUShares  int64  `json:"cpu_shares" yaml:"cpu_shares"`
	Memory     int64  `json:"memory" yaml:"memory"`           // in bytes
	MemorySwap int64  `json:"memory_swap" yaml:"memory_swap"` // in bytes
	CPUs       string `json:"cpus" yaml:"cpus"`               // e.g., "0.5", "2"
	Pids       int64  `json:"pids" yaml:"pids"`
	DiskQuota  int64  `json:"disk_quota" yaml:"disk_quota"` // in bytes
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Name       string            `json:"name" yaml:"name"`
	Driver     string            `json:"driver" yaml:"driver"`
	Subnet     string            `json:"subnet" yaml:"subnet"`
	Gateway    string            `json:"gateway" yaml:"gateway"`
	Labels     map[string]string `json:"labels" yaml:"labels"`
	Internal   bool              `json:"internal" yaml:"internal"`
	EnableIPv6 bool              `json:"enable_ipv6" yaml:"enable_ipv6"`
}

// PortMapping represents port mapping configuration
type PortMapping struct {
	HostPort      int    `json:"host_port" yaml:"host_port"`
	ContainerPort int    `json:"container_port" yaml:"container_port"`
	Protocol      string `json:"protocol" yaml:"protocol"` // tcp, udp
	HostIP        string `json:"host_ip" yaml:"host_ip"`
}

// VolumeMapping represents volume mapping configuration
type VolumeMapping struct {
	HostPath      string `json:"host_path" yaml:"host_path"`
	ContainerPath string `json:"container_path" yaml:"container_path"`
	ReadOnly      bool   `json:"read_only" yaml:"read_only"`
	Type          string `json:"type" yaml:"type"` // bind, volume, tmpfs
}

// EnvironmentVariable represents environment variable
type EnvironmentVariable struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test        []string      `json:"test" yaml:"test"`
	Interval    time.Duration `json:"interval" yaml:"interval"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	Retries     int           `json:"retries" yaml:"retries"`
	StartPeriod time.Duration `json:"start_period" yaml:"start_period"`
	Disable     bool          `json:"disable" yaml:"disable"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Driver  string            `json:"driver" yaml:"driver"`
	Options map[string]string `json:"options" yaml:"options"`
}
