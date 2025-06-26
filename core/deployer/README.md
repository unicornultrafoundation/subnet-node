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

#### Cache (`deployment/cache.go`)
- Thread-safe deployment ID cache
- Tracks active deployments for monitoring

### 4. Kubernetes Integration (`kube/`)

Provides Kubernetes cluster operations:

#### Client (`kube/client.go`)
- Kubernetes client initialization
- Connection management

#### Deploy (`kube/deploy.go`)
- Deployment creation and management
- Service and ingress configuration
- Resource allocation and scheduling

#### Status (`kube/status.go`)
- Deployment status retrieval
- Service endpoint information
- Resource usage monitoring

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
- **Resource Monitoring**: CPU, memory, and storage usage tracking
- **Stability Reporting**: Deployment stability status reporting

## Error Handling

The package includes comprehensive error handling:

- **Validation Errors**: SDL manifest validation failures
- **Kubernetes Errors**: Cluster operation failures
- **Storage Errors**: Datastore operation failures
- **Timeout Errors**: Deployment readiness timeout
- **Resource Errors**: Resource allocation failures

## Dependencies

- **Kubernetes Client**: `k8s.io/client-go`
- **IPFS Datastore**: `github.com/ipfs/go-datastore`
- **Logging**: `github.com/sirupsen/logrus`
- **Ethereum**: `github.com/ethereum/go-ethereum/common`

## Testing

The package includes comprehensive test coverage:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end deployment testing
- **Mock Testing**: Mocked dependencies for isolated testing

Run tests with:
```bash
go test ./core/deployer/...
```

## Security Considerations

- **Authentication**: Requester signature validation
- **Isolation**: Multi-tenant deployment isolation
- **Resource Limits**: Kubernetes resource constraints
- **TTL Enforcement**: Automatic cleanup to prevent resource exhaustion
- **Network Security**: Service exposure controls

## Performance

- **Caching**: In-memory deployment cache for fast lookups
- **Async Operations**: Non-blocking deployment operations
- **Resource Optimization**: Efficient Kubernetes resource management
- **Monitoring Efficiency**: Configurable monitoring intervals

## Contributing

When contributing to the deployer package:

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new functionality
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Validate SDL manifest compatibility
6. Test with different Kubernetes configurations 