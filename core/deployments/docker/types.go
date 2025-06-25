package deployment

import (
	"time"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

// NetworkConfig represents network configuration for a tenant
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

// ResourceLimits represents resource limits for a container
type ResourceLimits struct {
	CPUShares  int64  `json:"cpu_shares" yaml:"cpu_shares"`
	Memory     int64  `json:"memory" yaml:"memory"`           // in bytes
	MemorySwap int64  `json:"memory_swap" yaml:"memory_swap"` // in bytes
	CPUs       string `json:"cpus" yaml:"cpus"`               // e.g., "0.5", "2"
	Pids       int64  `json:"pids" yaml:"pids"`
	DiskQuota  int64  `json:"disk_quota" yaml:"disk_quota"` // in bytes
}

// ServiceConfig represents a service configuration in docker-compose
type ServiceConfig struct {
	Name          string                `json:"name" yaml:"name"`
	Image         string                `json:"image" yaml:"image"`
	ImageTag      string                `json:"image_tag" yaml:"image_tag"`
	Command       []string              `json:"command" yaml:"command"`
	Entrypoint    []string              `json:"entrypoint" yaml:"entrypoint"`
	WorkingDir    string                `json:"working_dir" yaml:"working_dir"`
	User          string                `json:"user" yaml:"user"`
	Ports         []PortMapping         `json:"ports" yaml:"ports"`
	Volumes       []VolumeMapping       `json:"volumes" yaml:"volumes"`
	Environment   []EnvironmentVariable `json:"environment" yaml:"environment"`
	EnvFile       []string              `json:"env_file" yaml:"env_file"`
	DependsOn     []string              `json:"depends_on" yaml:"depends_on"`
	RestartPolicy string                `json:"restart_policy" yaml:"restart_policy"`
	HealthCheck   *HealthCheck          `json:"health_check" yaml:"health_check"`
	Resources     *ResourceLimits       `json:"resources" yaml:"resources"`
	Networks      []string              `json:"networks" yaml:"networks"`
	Labels        map[string]string     `json:"labels" yaml:"labels"`
	Secrets       []string              `json:"secrets" yaml:"secrets"`
	Configs       []string              `json:"configs" yaml:"configs"`
	Deploy        *DeployConfig         `json:"deploy" yaml:"deploy"`
	ExtraHosts    []string              `json:"extra_hosts" yaml:"extra_hosts"`
	CapAdd        []string              `json:"cap_add" yaml:"cap_add"`
	CapDrop       []string              `json:"cap_drop" yaml:"cap_drop"`
	SecurityOpt   []string              `json:"security_opt" yaml:"security_opt"`
	Privileged    bool                  `json:"privileged" yaml:"privileged"`
	ReadOnly      bool                  `json:"read_only" yaml:"read_only"`
	Tmpfs         []string              `json:"tmpfs" yaml:"tmpfs"`
	Logging       *LoggingConfig        `json:"logging" yaml:"logging"`
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

// DeployConfig represents deployment configuration
type DeployConfig struct {
	Replicas       int              `json:"replicas" yaml:"replicas"`
	UpdateConfig   *UpdateConfig    `json:"update_config" yaml:"update_config"`
	RollbackConfig *UpdateConfig    `json:"rollback_config" yaml:"rollback_config"`
	RestartPolicy  *RestartPolicy   `json:"restart_policy" yaml:"restart_policy"`
	Resources      *ResourceLimits  `json:"resources" yaml:"resources"`
	Placement      *PlacementConfig `json:"placement" yaml:"placement"`
}

// UpdateConfig represents update configuration
type UpdateConfig struct {
	Parallelism     uint64        `json:"parallelism" yaml:"parallelism"`
	Delay           time.Duration `json:"delay" yaml:"delay"`
	FailureAction   string        `json:"failure_action" yaml:"failure_action"`
	Monitor         time.Duration `json:"monitor" yaml:"monitor"`
	MaxFailureRatio float64       `json:"max_failure_ratio" yaml:"max_failure_ratio"`
	Order           string        `json:"order" yaml:"order"`
}

// RestartPolicy represents restart policy configuration
type RestartPolicy struct {
	Condition   string        `json:"condition" yaml:"condition"`
	Delay       time.Duration `json:"delay" yaml:"delay"`
	MaxAttempts uint64        `json:"max_attempts" yaml:"max_attempts"`
	Window      time.Duration `json:"window" yaml:"window"`
}

// PlacementConfig represents placement configuration
type PlacementConfig struct {
	Constraints        []string              `json:"constraints" yaml:"constraints"`
	Preferences        []PlacementPreference `json:"preferences" yaml:"preferences"`
	MaxReplicasPerNode *uint64               `json:"max_replicas_per_node" yaml:"max_replicas_per_node"`
}

// PlacementPreference represents placement preference
type PlacementPreference struct {
	Spread string `json:"spread" yaml:"spread"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Driver  string            `json:"driver" yaml:"driver"`
	Options map[string]string `json:"options" yaml:"options"`
}

// ComposeManifest represents a docker-compose manifest
type ComposeManifest struct {
	Version    string                   `json:"version" yaml:"version"`
	Services   map[string]ServiceConfig `json:"services" yaml:"services"`
	Networks   map[string]NetworkConfig `json:"networks" yaml:"networks"`
	Volumes    map[string]interface{}   `json:"volumes" yaml:"volumes"`
	Secrets    map[string]interface{}   `json:"secrets" yaml:"secrets"`
	Configs    map[string]interface{}   `json:"configs" yaml:"configs"`
	Extensions map[string]interface{}   `json:"extensions" yaml:"extensions"`
}

// Deployment represents a deployment configuration
type Deployment struct {
	ID          string            `json:"id" yaml:"id"`
	TenantID    string            `json:"tenant_id" yaml:"tenant_id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Manifest    ComposeManifest   `json:"manifest" yaml:"manifest"`
	Status      string            `json:"status" yaml:"status"`
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

// PortPool represents a pool of available ports
type PortPool struct {
	StartPort int            `json:"start_port" yaml:"start_port"`
	EndPort   int            `json:"end_port" yaml:"end_port"`
	UsedPorts map[int]string `json:"used_ports" yaml:"used_ports"` // port -> tenant_id
}

// TenantResourceUsage represents resource usage for a tenant
type TenantResourceUsage struct {
	TenantID       string  `json:"tenant_id" yaml:"tenant_id"`
	CPUUsage       float64 `json:"cpu_usage" yaml:"cpu_usage"`
	MemoryUsage    int64   `json:"memory_usage" yaml:"memory_usage"`
	DiskUsage      int64   `json:"disk_usage" yaml:"disk_usage"`
	NetworkUsage   int64   `json:"network_usage" yaml:"network_usage"`
	PortCount      int     `json:"port_count" yaml:"port_count"`
	ContainerCount int     `json:"container_count" yaml:"container_count"`
}

// DeploymentEvent represents an event in the deployment lifecycle
type DeploymentEvent struct {
	ID           string                 `json:"id" yaml:"id"`
	DeploymentID string                 `json:"deployment_id" yaml:"deployment_id"`
	TenantID     string                 `json:"tenant_id" yaml:"tenant_id"`
	Type         string                 `json:"type" yaml:"type"`
	Message      string                 `json:"message" yaml:"message"`
	Timestamp    time.Time              `json:"timestamp" yaml:"timestamp"`
	Data         map[string]interface{} `json:"data" yaml:"data"`
}
