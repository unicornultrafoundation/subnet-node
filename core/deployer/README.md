# Core Deployer Package

The `core/deployer` package provides a comprehensive Kubernetes-based deployment system for containerized applications. It handles the lifecycle management of deployments, including creation, monitoring, cleanup, and status tracking.

## Overview

The deployer package is designed to manage application deployments in Kubernetes clusters with the following key features:

- **Service Definition Language (SDL) Support**: Parse and validate SDL manifests for application deployment
- **Kubernetes Integration**: Direct integration with Kubernetes clusters for deployment management
- **Persistent Storage**: Store deployment requests and metadata using IPFS datastore
- **Monitoring & Health Checks**: Continuous monitoring of deployment health and status
- **TTL Management**: Automatic cleanup of expired deployments
- **Multi-tenant Support**: Support for multiple requesters with isolated deployments
- **GPU Resource Management (WIP)**: GPU support for NVIDIA, AMD, Intel, and Apple Silicon GPUs
- **Advanced Resource Monitoring**: Real-time CPU, memory, storage, network, and GPU usage tracking
- **Enhanced Error Handling**: Improved validation, rollback, and error recovery mechanisms
- **Dependency Management**: Intelligent service dependency resolution and deployment ordering
- **JSON-RPC API**: Full JSON-RPC 2.0 API for deployment operations
- **WebSocket Support**: Real-time log streaming and command execution

## Architecture

The package is organized into several key components:

```
core/deployer/
├── config.go          # Service configuration management
├── service.go         # Main deployer service orchestration
├── deployment.go      # Public API for deployment operations
├── types/             # Data structures and types
├── deployment/        # Deployment lifecycle management
├── kube/              # Kubernetes client and operations
├── store/             # Persistent storage layer
└── manifest/          # SDL parsing and validation
```

## Components

### 1. Service (`service.go`)

The main orchestrator that coordinates all deployer operations:

```go
type Service struct {
    config            *ServiceConfig
    deploymentService *deployment.Service
    storeService      *store.Service
    logger            *logrus.Logger
}
```

**Key Responsibilities:**
- Initialize and start all sub-services
- Manage service lifecycle (start/stop)
- Coordinate between Kubernetes client, storage, and deployment management

### 2. Configuration (`config.go`)

Manages deployer service configuration:

```go
type ServiceConfig struct {
    KubeConfigPath        string
    MonitorInterval       time.Duration
    DefaultServiceType    string
    LocalhostEnabled      bool
    DeploymentWaitTimeout time.Duration
}
```

**Configuration Options:**
- `kubeconfig_path`: Path to Kubernetes configuration file
- `monitor_interval`: Interval for deployment monitoring (default: 30s)
- `default_service_type`: Default Kubernetes service type (default: "NodePort")
- `localhost_enabled`: Enable localhost access (default: true)
- `deployment_wait_timeout`: Timeout for deployment readiness (default: 30s)

### 3. Deployment Management (`deployment/`)

Handles the core deployment lifecycle operations:

#### Service (`deployment/service.go`)
- Manages deployment requests and responses
- Maintains deployment cache for monitoring
- Coordinates with Kubernetes client and storage
- Thread-safe deployment cache with mutex protection
- Enhanced deployment list management for monitoring

#### Handler (`deployment/handler.go`)
- Processes deployment requests
- Validates SDL manifests
- Deploys applications to Kubernetes
- Handles deployment cleanup

#### Monitor (`deployment/monitor.go`)
- Continuously monitors deployment health
- Checks TTL expiration
- Triggers automatic cleanup for expired deployments
- Reports deployment stability status
- Enhanced stability monitoring with detailed logging
- Improved TTL enforcement and cleanup tracking

#### Cache (`deployment/cache.go`)
- Thread-safe deployment ID cache
- Tracks active deployments for monitoring

### 4. Kubernetes Integration (`kube/`)

Provides Kubernetes cluster operations with enhanced capabilities:

#### Client (`kube/client.go`)
- Kubernetes client initialization
- Connection management
- Metrics client integration for resource monitoring
- Enhanced configuration options for service types

#### Deploy (`kube/deploy.go`)
- Deployment creation and management
- Service and ingress configuration
- Resource allocation and scheduling
- GPU resource support with automatic limit setting
- Enhanced dependency resolution for service deployment
- Improved resource validation and compute profile handling
- Automatic rollback on deployment failures
- Enhanced namespace and resource cleanup
- Better error handling with detailed validation messages

#### Status (`kube/status.go`)
- Deployment status retrieval
- Service endpoint information
- Resource usage monitoring
- Enhanced pod and container status tracking
- Detailed resource usage reporting
- Improved service status aggregation
- Better error handling for missing deployments

#### Stats (`kube/stats.go`)
- Resource usage monitoring
- GPU usage tracking for multiple GPU types
- Network statistics collection
- Storage usage monitoring
- Optimized metrics collection with fallback mechanisms
- Support for Apple Silicon GPU monitoring
- Enhanced error handling and logging

#### Cleanup (`kube/cleanup.go`)
- Resource cleanup operations
- Namespace termination
- Resource deletion

#### Wait (`kube/wait.go`)
- Deployment readiness waiting
- Resource availability checking

### 5. Storage Layer (`store/`)

Persistent storage for deployment metadata:

```go
type Service struct {
    Datastore datastore.Datastore
    logger    *logrus.Logger
}
```

**Operations:**
- Store deployment requests
- Retrieve deployment metadata
- Delete deployment records
- Query deployment history

### 6. Manifest Processing (`manifest/`)

SDL (Service Definition Language) parsing and validation:

- **SDL Structure**: Parse YAML/JSON service definitions
- **Validation**: Validate manifest syntax and requirements
- **Conversion**: Convert SDL to Kubernetes manifests
- **Resource Management**: Handle CPU, memory, storage, and GPU requirements

See [manifest/README.md](manifest/README.md) for detailed SDL documentation.

### 7. Types (`types/`)

Data structures for deployment operations:

#### DeploymentRequest
```go
type DeploymentRequest struct {
    OrderID   string       `json:"order_id"`
    Manifest  manifest.SDL `json:"manifest"`
    Signature string       `json:"signature"`
    Requester string       `json:"requester"`
    TTL       int64        `json:"ttl"` // in minutes
}
```

#### DeploymentResponse
```go
type DeploymentResponse struct {
    ID        string
    Requester common.Address
    Status    *DeploymentStatus
}
```

#### DeploymentStatus
```go
type DeploymentStatus struct {
    State      DeploymentState `json:"state"`
    Services   []ServiceStatus `json:"services"`
    Endpoints  []EndpointInfo  `json:"endpoints"`
    CreatedAt  string          `json:"createdAt"`
    UpdatedAt  string          `json:"updatedAt"`
    DeployedAt string          `json:"deployedAt"`
    TTL        int64           `json:"ttl"`
    TimeLeft   int64           `json:"timeLeft"`
}
```

#### DeploymentStats
```go
type DeploymentStats struct {
    UsedCpu           uint64 `json:"usedCpu"`           // millicores
    UsedMemory        uint64 `json:"usedMemory"`        // bytes
    UsedStorage       uint64 `json:"usedStorage"`       // bytes
    UsedUploadBytes   uint64 `json:"usedUploadBytes"`   // bytes
    UsedDownloadBytes uint64 `json:"usedDownloadBytes"` // bytes
    UsedGpu           uint64 `json:"usedGpu"`           // GPU count
    Duration          int64  `json:"duration"`          // seconds
}
```

## API Documentation

The deployer package exposes a comprehensive JSON-RPC 2.0 API for deployment operations, along with WebSocket endpoints for real-time functionality.

For complete API documentation including all methods, examples, and WebSocket message formats, see [API.md](API.md).

## New Features and Improvements

### GPU Resource Support

The deployer now provides comprehensive GPU support:

- **Multiple GPU Types**: Support for NVIDIA, AMD, Intel, and Apple Silicon GPUs
- **Automatic Limit Setting**: GPU limits are automatically set equal to requests for non-overcommitable resources
- **GPU Usage Monitoring**: Real-time GPU utilization tracking
- **Apple Silicon Support**: Native support for Apple Silicon GPU monitoring
- **Fallback Mechanisms**: Graceful handling when GPU monitoring tools are unavailable

### Enhanced Resource Monitoring

- **Real-time Metrics**: Integration with Kubernetes Metrics API for live resource usage
- **Network Statistics**: Advanced network usage tracking with multiple collection methods
- **Storage Monitoring**: Comprehensive storage usage tracking including persistent volumes
- **Optimized Collection**: Reduced timeouts and improved performance for metrics collection

### Improved Error Handling

- **Validation Enhancements**: More detailed validation with specific error messages
- **Rollback Support**: Automatic cleanup of resources on deployment failures
- **Resource Tracking**: Comprehensive tracking of created resources for cleanup
- **Dependency Resolution**: Intelligent handling of service dependencies with circular dependency detection

### Enhanced Deployment Management

- **Thread-safe Operations**: Mutex-protected deployment cache for concurrent access
- **Stability Monitoring**: Improved deployment stability tracking and reporting
- **TTL Enforcement**: Enhanced TTL management with automatic cleanup
- **Service Dependencies**: Intelligent deployment ordering based on service dependencies

### Better Kubernetes Integration

- **Metrics Client**: Integration with Kubernetes Metrics API for resource monitoring
- **Enhanced Validation**: Comprehensive validation of compute profiles and resources
- **Improved Resource Management**: Better handling of CPU, memory, and storage resources
- **Service Type Configuration**: Flexible service type configuration with localhost support

## Usage

### Initialization

```go
import "github.com/unicornultrafoundation/subnet-node/core/deployer"

// Create deployer service
config := config.NewConfig()
ds := datastore.NewMapDatastore()
deployerService, err := deployer.NewService(config, ds)
if err != nil {
    log.Fatal(err)
}

// Start the service
ctx := context.Background()
if err := deployerService.Start(ctx); err != nil {
    log.Fatal(err)
}
defer deployerService.Stop(ctx)
```

### Deployment Operations

#### Request Deployment
```go
deploymentRequest := &types.DeploymentRequest{
    OrderID:   "order-123",
    Manifest:  sdlManifest,
    Signature: "signature",
    Requester: "0x123...",
    TTL:       60, // 60 minutes
}

response, err := deployerService.RequestDeployment(ctx, deploymentRequest)
if err != nil {
    log.Error("Deployment failed:", err)
}
```

#### Get Deployment Status
```go
deployment, err := deployerService.GetDeployment(ctx, "order-123")
if err != nil {
    log.Error("Failed to get deployment:", err)
}
```

#### Get Deployment Statistics
```go
stats, err := deployerService.GetDeploymentStats(ctx, "order-123")
if err != nil {
    log.Error("Failed to get deployment stats:", err)
} else {
    log.Printf("CPU: %d mCPU, Memory: %d bytes, GPU: %d", 
        stats.UsedCpu, stats.UsedMemory, stats.UsedGpu)
}
```

#### List Deployments
```go
deployments, err := deployerService.GetDeployments(ctx, "0x123...")
if err != nil {
    log.Error("Failed to get deployments:", err)
}
```

#### Cleanup Deployment
```go
err := deployerService.CleanupDeployment(ctx, "order-123")
if err != nil {
    log.Error("Failed to cleanup deployment:", err)
}
```

## Configuration

The deployer service can be configured through the main application configuration:

```yaml
deployer:
  kubeconfig_path: "/path/to/kubeconfig"
  monitor_interval: "30s"
  default_service_type: "NodePort"
  localhost_enabled: true
  deployment_wait_timeout: "30s"
```

## Monitoring

The deployer service provides comprehensive monitoring capabilities:

- **Health Monitoring**: Continuous monitoring of deployment health
- **TTL Management**: Automatic cleanup of expired deployments
- **Status Tracking**: Real-time deployment status updates
- **Resource Monitoring**: CPU, memory, storage, network, and GPU usage tracking
- **Stability Reporting**: Deployment stability status reporting
- **Performance Metrics**: Real-time performance metrics collection

## Error Handling

The package includes comprehensive error handling:

- **Validation Errors**: SDL manifest validation failures with detailed messages
- **Kubernetes Errors**: Cluster operation failures with rollback support
- **Storage Errors**: Datastore operation failures
- **Timeout Errors**: Deployment readiness timeout
- **Resource Errors**: Resource allocation failures with cleanup
- **GPU Errors**: GPU resource errors with fallback mechanisms

## Dependencies

- **Kubernetes Client**: `k8s.io/client-go`
- **Kubernetes Metrics**: `k8s.io/metrics/pkg/client/clientset/versioned`
- **IPFS Datastore**: `github.com/ipfs/go-datastore`
- **Logging**: `github.com/sirupsen/logrus`
- **Ethereum**: `github.com/ethereum/go-ethereum/common`

## Testing

The package includes comprehensive test coverage:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end deployment testing
- **Mock Testing**: Mocked dependencies for isolated testing
- **GPU Testing**: GPU resource testing with various configurations

Run tests with:
```bash
go test ./core/deployer/...
```

## Security Considerations

- **Authentication**: Requester signature validation
- **Isolation**: Multi-tenant deployment isolation
- **Resource Limits**: Kubernetes resource constraints with GPU limits
- **TTL Enforcement**: Automatic cleanup to prevent resource exhaustion
- **Network Security**: Service exposure controls
- **Resource Validation**: Comprehensive resource validation and sanitization

## Performance

- **Caching**: In-memory deployment cache for fast lookups
- **Async Operations**: Non-blocking deployment operations
- **Resource Optimization**: Efficient Kubernetes resource management
- **Monitoring Efficiency**: Configurable monitoring intervals
- **Metrics Optimization**: Optimized metrics collection with reduced timeouts
- **GPU Optimization**: Efficient GPU resource management and monitoring

## Contributing

When contributing to the deployer package:

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new functionality
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Validate SDL manifest compatibility
6. Test with different Kubernetes configurations
7. Test GPU functionality with various GPU types
8. Validate resource monitoring accuracy
9. Ensure thread safety for concurrent operations 