# Deployments Module

The Deployments module provides a unified interface for managing Kubernetes deployments in the subnet-node project. It supports Kubernetes-based deployments with comprehensive monitoring, logging, and management capabilities.

## Architecture

```
core/deployments/
├── types.go              # Common types and interfaces
├── interfaces.go          # Core interfaces for deployment management
├── service.go            # Main deployment service
├── factory.go            # Factory for creating deployment managers
├── api.go                # REST API endpoints
├── api_test.go           # API tests
├── example.go            # Usage examples
├── example_client.go     # Example client implementation
├── AUTHORIZATION.md      # Authorization documentation
└── README.md            # This documentation
```

## Supported Deployment Types

### 1. Kubernetes Deployment ✅
- **Status**: Fully implemented
- **Features**: Pod management, service discovery, namespace isolation, resource quotas, comprehensive monitoring
- **Location**: `core/deployments/` (integrated into main service)

## Core Components

### Service
The main `Service` struct provides a unified interface for managing Kubernetes deployments:

```go
service := deployments.NewService(cfg, logger)
service.SetManagers(eventManager, storageManager, resourceManager, factory)
service.Start(ctx)
```

### Factory
The `Factory` creates deployment managers for Kubernetes deployments:

```go
factory := deployments.NewFactory(logger)
factory.RegisterManagerCreator(deployments.DeploymentTypeKubernetes, kubernetes.CreateManager)
```

### Interfaces
Common interfaces that the deployment system implements:

- `DeploymentManager`: Core deployment operations
- `ManifestManager`: Manifest processing
- `PortManager`: Port allocation
- `NetworkManager`: Network management
- `ResourceManager`: Resource tracking
- `EventManager`: Event handling
- `StorageManager`: Data persistence

## Features

### 1. Deployment Management
- Create, update, delete Kubernetes deployments
- Start, stop, restart deployments
- Scale deployments up/down
- Update deployment images

### 2. Monitoring & Inspection
- **Deployment Inspection**: Get detailed information about deployments
- **Service Inspection**: Get detailed information about specific services
- **Resource Monitoring**: Track CPU, memory, disk, and network usage
- **Health Checks**: Monitor service health and status
- **Metrics Collection**: Collect performance metrics over time

### 3. Logs Management
- **Log Retrieval**: Get logs for deployments and services
- **Real-time Log Streaming**: Stream logs in real-time with follow mode
- **Log Filtering**: Filter logs by service, time range, log level
- **Log Aggregation**: Aggregate logs across multiple services

### 4. Console Execution
- **Command Execution**: Execute commands in deployment pods
- **Interactive Console**: Interactive terminal access to pods
- **WebSocket Support**: Real-time interactive console via WebSocket
- **Session Management**: Manage multiple execution sessions

### 5. Resource Management
- **Resource Limits**: Set and enforce resource limits via Kubernetes resource quotas
- **Port Management**: Automatic port allocation and management
- **Network Management**: Network configuration and management

## Usage Examples

### Creating a Kubernetes Deployment

```go
import (
    "github.com/unicornultrafoundation/subnet-node/core/deployments"
)

// Create deployment
deployment := &deployments.Deployment{
    ID:       "deploy-123",
    Name:     "web-app",
    Type:     deployments.DeploymentTypeKubernetes,
    Manifest: &kubernetes.Manifest{
        // Kubernetes manifest configuration
    },
}

err := service.CreateDeployment(ctx, deployment)
```

### Getting Deployment Logs

```go
// Get logs for a deployment
logs, err := service.GetDeploymentLogs(ctx, "deploy-123", "", 100)
if err != nil {
    log.Fatal(err)
}
defer logs.Close()

// Copy logs to stdout
io.Copy(os.Stdout, logs)
```

### Streaming Real-time Logs

```go
// Stream logs in real-time
logChan, err := service.StreamDeploymentLogs(ctx, "deploy-123", "web", true)
if err != nil {
    log.Fatal(err)
}

for logEntry := range logChan {
    fmt.Printf("[%s] %s: %s\n", logEntry.Timestamp, logEntry.Service, logEntry.Message)
}
```

### Executing Commands

```go
// Execute a command in a pod
session, err := service.ExecConsole(ctx, "deploy-123", "web", []string{"ls", "-la"}, false)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

result, err := session.Execute(ctx, []string{"ls", "-la"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Exit code: %d\n", result.ExitCode)
fmt.Printf("Output: %s\n", result.Stdout)
```

### Inspecting Deployments

```go
// Get detailed deployment information
inspection, err := service.InspectDeployment(ctx, "deploy-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deployment: %s\n", inspection.Name)
fmt.Printf("Status: %s\n", inspection.Status)
fmt.Printf("Services: %d\n", len(inspection.Services))

for name, service := range inspection.Services {
    fmt.Printf("  - %s: %s (%s)\n", name, service.Image, service.Status)
}
```

### Getting Metrics

```go
// Get deployment metrics for the last hour
metrics, err := service.GetDeploymentMetrics(ctx, "deploy-123", time.Hour)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("CPU Usage: %.2f%%\n", metrics.Total.CPUUsage)
fmt.Printf("Memory Usage: %d MB\n", metrics.Total.MemoryUsage/1024/1024)
```

## Configuration

The deployment module can be configured in the main configuration file:

```yaml
deployments:
  enabled: true
  
  # Kubernetes configuration
  kubernetes:
    enabled: true
    kubeconfig: "/path/to/kubeconfig"
    namespace_prefix: "subnet-"
    resource_quotas:
      enabled: true
      default_cpu: "1000m"
      default_memory: "1Gi"
      default_storage: "10Gi"
```

## API Endpoints

The deployment service provides comprehensive REST API endpoints:

### Deployment Management
- `POST /api/v1/deployments` - Create deployment
- `GET /api/v1/deployments` - List deployments
- `GET /api/v1/deployments/{id}` - Get deployment
- `PUT /api/v1/deployments/{id}` - Update deployment
- `DELETE /api/v1/deployments/{id}` - Delete deployment
- `POST /api/v1/deployments/{id}/start` - Start deployment
- `POST /api/v1/deployments/{id}/stop` - Stop deployment
- `POST /api/v1/deployments/{id}/restart` - Restart deployment

### Monitoring & Inspection
- `GET /api/v1/deployments/{id}/inspect` - Inspect deployment
- `GET /api/v1/deployments/{id}/services/{service}/inspect` - Inspect service
- `GET /api/v1/deployments/{id}/metrics` - Get deployment metrics
- `GET /api/v1/deployments/{id}/services/{service}/metrics` - Get service metrics

### Logs
- `GET /api/v1/deployments/{id}/logs` - Get deployment logs
- `GET /api/v1/deployments/{id}/services/{service}/logs` - Get service logs
- `GET /api/v1/deployments/{id}/logs/stream` - Stream deployment logs
- `GET /api/v1/deployments/{id}/services/{service}/logs/stream` - Stream service logs

### Console Execution
- `POST /api/v1/deployments/{id}/services/{service}/exec` - Execute command
- `GET /api/v1/deployments/{id}/services/{service}/exec/ws` - Interactive console (WebSocket)

### Scaling & Updates
- `POST /api/v1/deployments/{id}/services/{service}/scale` - Scale deployment
- `PUT /api/v1/deployments/{id}/services/{service}/image` - Update image

### Events
- `GET /api/v1/deployments/{id}/events` - Get deployment events
- `GET /api/v1/deployments/{id}/events/stream` - Stream deployment events

## Monitoring and Events

The deployment service emits events for all operations:

- `deployment.created`
- `deployment.started`
- `deployment.stopped`
- `deployment.failed`
- `deployment.deleted`
- `deployment.scaled`
- `deployment.updated`
- `service.health_check`
- `service.logs`
- `service.metrics`

Events can be subscribed to for monitoring and integration with external systems.

## Security

- **Network Isolation**: Prevents unauthorized communication via network policies
- **Resource Limits**: Prevents resource exhaustion via Kubernetes resource quotas
- **Audit Logging**: Logs all operations for security auditing
- **Access Control**: Role-based access control for API endpoints
- **Secure Execution**: Secure command execution with proper isolation

## Performance

- **Efficient Resource Management**: Optimized resource allocation and monitoring via Kubernetes APIs
- **Real-time Streaming**: Efficient real-time log and event streaming
- **Caching**: Intelligent caching for frequently accessed data
- **Connection Pooling**: Efficient connection management for Kubernetes API
- **Metrics Collection**: Low-overhead metrics collection and aggregation

## Contributing

The deployment system is currently focused on Kubernetes deployments. To extend functionality:

1. Implement additional Kubernetes-specific features
2. Add new monitoring and management capabilities
3. Enhance the API with new endpoints
4. Improve performance and scalability
5. Update this README with new features

## Status

- ✅ Kubernetes deployment: Complete with full monitoring, logs, and exec support

The Kubernetes deployment module is fully functional and ready for production use with comprehensive monitoring, logging, and management capabilities. The system has been optimized for Kubernetes-native deployments with full integration to Kubernetes APIs and features. 