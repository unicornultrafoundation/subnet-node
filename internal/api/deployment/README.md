# Deployment API Server

A comprehensive API server for managing Kubernetes deployments in the subnet-node system.

## Overview

The Deployment API Server provides a clean interface for deployment management operations, wrapping the core deployment service and providing consistent response formatting. It follows the established patterns used throughout the subnet-node project.

## Features

- **Deployment Management**: Create, retrieve, and cleanup deployments
- **Service Status**: Get detailed status information for services within deployments
- **Log Streaming**: Access deployment logs
- **Command Execution**: Execute commands in deployment pods
- **Cache Management**: Manage deployment cache operations
- **Statistics**: Get deployment statistics and metrics
- **Consistent API**: Standardized response format with error handling

## Architecture

The API server uses an interface-based design for better testability and flexibility:

```
APIServer → DeploymentServiceInterface → Core Deployment Service
```

### Response Format

All API responses follow a consistent format:

```go
type Response[T any] struct {
    Data  T      `json:"data"`
    Error string `json:"error"`
}
```

## API Methods

### Core Deployment Operations

#### RequestDeployment
Creates a new deployment from a deployment request.

```go
func (api *APIServer) RequestDeployment(ctx context.Context, req *types.DeploymentRequest) *Response[*types.DeploymentResponse]
```

#### GetDeployment
Retrieves deployment information by order ID.

```go
func (api *APIServer) GetDeployment(ctx context.Context, orderID string) *Response[*types.DeploymentResponse]
```

#### CleanupDeployment
Removes a deployment and cleans up associated resources.

```go
func (api *APIServer) CleanupDeployment(ctx context.Context, orderID string) *Response[string]
```

#### GetDeployments
Lists all deployments for a specific requester.

```go
func (api *APIServer) GetDeployments(ctx context.Context, requester string) *Response[[]*types.DeploymentResponse]
```

### Service and Pod Operations

#### GetServiceStatus
Retrieves the status of a specific service within a deployment.

```go
func (api *APIServer) GetServiceStatus(ctx context.Context, orderID, serviceName string) *Response[*types.ServiceStatus]
```

#### GetDeploymentLogs
Streams logs for a deployment.

```go
func (api *APIServer) GetDeploymentLogs(ctx context.Context, orderID string) *Response[[]*types.ServiceLog]
```

#### Exec
Executes a command in a deployment's pod/service.

```go
func (api *APIServer) Exec(ctx context.Context, orderID, podName, serviceName string, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) *Response[types.ExecResult]
```

### Cache Management

#### GetDeploymentListCache
Returns the list of deployment IDs in the cache.

```go
func (api *APIServer) GetDeploymentListCache(ctx context.Context) *Response[[]string]
```

#### DeleteDeploymentListCache
Removes a deployment ID from the cache.

```go
func (api *APIServer) DeleteDeploymentListCache(ctx context.Context, deploymentID string) *Response[string]
```

#### AddDeploymentListCache
Adds a deployment ID to the cache.

```go
func (api *APIServer) AddDeploymentListCache(ctx context.Context, deploymentID string) *Response[string]
```

### Statistics and Monitoring

#### GetDeploymentStats
Returns statistics about deployments.

```go
func (api *APIServer) GetDeploymentStats(ctx context.Context) *Response[*DeploymentStats]
```

## Usage Examples

### Creating a Deployment API Server

```go
import (
    "github.com/unicornultrafoundation/subnet-node/internal/api/deployment"
    "github.com/unicornultrafoundation/subnet-node/core/deployer/deployment"
)

// Create the core deployment service
deploymentService := deployment.NewService(kubeClient, store, logger, config)

// Create the API server
apiServer := deployment.NewAPIServer(deploymentService)
```

### Requesting a Deployment

```go
ctx := context.Background()
req := &types.DeploymentRequest{
    OrderID:   "order-123",
    Requester: "0x1234567890123456789012345678901234567890",
    TTL:       60, // 60 minutes
    Manifest:  manifest.SDL{...},
    Signature: "signature",
}

response := apiServer.RequestDeployment(ctx, req)
if response.Error != "" {
    log.Printf("Deployment failed: %s", response.Error)
    return
}

log.Printf("Deployment created: %s", response.Data.ID)
```

### Getting Deployment Status

```go
ctx := context.Background()
orderID := "order-123"

response := apiServer.GetDeployment(ctx, orderID)
if response.Error != "" {
    log.Printf("Failed to get deployment: %s", response.Error)
    return
}

deployment := response.Data
log.Printf("Deployment state: %s", deployment.Status.State)
```

### Getting Service Status

```go
ctx := context.Background()
orderID := "order-123"
serviceName := "web-service"

response := apiServer.GetServiceStatus(ctx, orderID, serviceName)
if response.Error != "" {
    log.Printf("Failed to get service status: %s", response.Error)
    return
}

status := response.Data
log.Printf("Service %s: %s (%d/%d replicas ready)", 
    status.Name, status.State, status.ReadyReplicas, status.Replicas)
```

### Cleanup Deployment

```go
ctx := context.Background()
orderID := "order-123"

response := apiServer.CleanupDeployment(ctx, orderID)
if response.Error != "" {
    log.Printf("Failed to cleanup deployment: %s", response.Error)
    return
}

log.Printf("Deployment cleaned up: %s", response.Data)
```

### Getting Deployment Statistics

```go
ctx := context.Background()

response := apiServer.GetDeploymentStats(ctx)
if response.Error != "" {
    log.Printf("Failed to get deployment stats: %s", response.Error)
    return
}

stats := response.Data
log.Printf("Total deployments: %d, Cache size: %d", 
    stats.TotalDeployments, stats.CacheSize)
```

## Error Handling

The API server provides consistent error handling through the `Response` struct:

- **Success**: `Error` field is empty, `Data` contains the result
- **Failure**: `Error` field contains the error message, `Data` is empty

```go
response := apiServer.GetDeployment(ctx, orderID)
if response.Error != "" {
    // Handle error
    return fmt.Errorf("deployment error: %s", response.Error)
}

// Use response.Data safely
deployment := response.Data
```

## Testing

The API server includes comprehensive tests with mocked dependencies:

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Run specific test
go test -v -run TestAPIServer_RequestDeployment_Success
```

### Test Structure

- **MockService**: Implements `DeploymentServiceInterface` for testing
- **TestAPIServer**: Wrapper for testing with mock service
- **Comprehensive Coverage**: Tests for success and error scenarios

## Integration

The deployment API server integrates with:

- **Core Deployment Service**: Handles actual deployment operations
- **Kubernetes Client**: Manages Kubernetes resources
- **Storage Service**: Persists deployment data
- **Logging**: Provides structured logging

## Dependencies

- `github.com/unicornultrafoundation/subnet-node/core/deployer/types`
- `k8s.io/client-go/tools/remotecommand`
- Standard Go libraries (`context`, `io`, `time`)

## Contributing

When adding new functionality:

1. Add the method to `DeploymentServiceInterface`
2. Implement the method in `APIServer`
3. Add comprehensive tests
4. Update this documentation
5. Ensure consistent error handling and response formatting 