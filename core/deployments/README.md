# Deployments Module

The Deployments module provides a unified interface for managing different types of deployments in the subnet-node project. It supports multiple deployment technologies including Docker, Kubernetes, Nomad, and Terraform.

## Architecture

```
core/deployments/
├── types.go              # Common types and interfaces
├── interfaces.go          # Core interfaces for all deployment types
├── service.go            # Main deployment service
├── factory.go            # Factory for creating deployment managers
├── api.go                # REST API endpoints
├── docker/               # Docker deployment implementation
│   ├── service.go        # Docker deployment service
│   ├── types.go          # Docker-specific types
│   ├── interfaces.go     # Docker-specific interfaces
│   ├── managers/         # Docker managers (tenant, port, etc.)
│   └── templates/        # Docker Compose templates
├── kubernetes/           # Kubernetes deployment (planned)
├── nomad/               # Nomad deployment (planned)
└── terraform/           # Terraform deployment (planned)
```

## Supported Deployment Types

### 1. Docker Deployment ✅
- **Status**: Fully implemented
- **Features**: Docker Compose support, tenant isolation, port management, resource limits
- **Location**: `core/deployments/docker/`

### 2. Kubernetes Deployment 🚧
- **Status**: Planned
- **Features**: Pod management, service discovery, namespace isolation, resource quotas
- **Location**: `core/deployments/kubernetes/`

### 3. Nomad Deployment 🚧
- **Status**: Planned
- **Features**: Job scheduling, service discovery, multi-region deployment
- **Location**: `core/deployments/nomad/`

### 4. Terraform Deployment 🚧
- **Status**: Planned
- **Features**: Infrastructure as Code, multi-cloud support, state management
- **Location**: `core/deployments/terraform/`

## Core Components

### Service
The main `Service` struct provides a unified interface for managing deployments across different technologies:

```go
service := deployments.NewService(cfg, logger)
service.SetManagers(tenantManager, eventManager, storageManager, resourceManager, factory)
service.Start(ctx)
```

### Factory
The `Factory` creates deployment managers for different types:

```go
factory := deployments.NewFactory(logger)
factory.RegisterManagerCreator(deployments.DeploymentTypeDocker, docker.CreateManager)
```

### Interfaces
Common interfaces that all deployment types must implement:

- `DeploymentManager`: Core deployment operations
- `TenantManager`: Tenant management
- `ManifestManager`: Manifest processing
- `PortManager`: Port allocation
- `NetworkManager`: Network management
- `ResourceManager`: Resource tracking
- `EventManager`: Event handling
- `StorageManager`: Data persistence

## Features

### 1. Deployment Management
- Create, update, delete deployments
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
- **Command Execution**: Execute commands in deployment containers
- **Interactive Console**: Interactive terminal access to containers
- **WebSocket Support**: Real-time interactive console via WebSocket
- **Session Management**: Manage multiple execution sessions

### 5. Resource Management
- **Tenant Isolation**: Complete isolation between tenants
- **Resource Limits**: Set and enforce resource limits
- **Port Management**: Automatic port allocation and management
- **Network Isolation**: Separate networks per tenant

## Usage Examples

### Creating a Docker Deployment

```go
import (
    "github.com/unicornultrafoundation/subnet-node/core/deployments"
    "github.com/unicornultrafoundation/subnet-node/core/deployments/docker"
)

// Create deployment
deployment := &deployments.Deployment{
    ID:       "deploy-123",
    TenantID: "tenant-456",
    Name:     "web-app",
    Type:     deployments.DeploymentTypeDocker,
    Manifest: &docker.ComposeManifest{
        Services: map[string]docker.Service{
            "web": {
                Image: "nginx:latest",
                Ports: []string{"8080:80"},
            },
        },
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
// Execute a command in a container
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
  
  # Docker configuration
  docker:
    enabled: true
    compose_path: "/usr/local/bin/docker-compose"
    port_pool:
      start: 30000
      end: 40000
    
  # Kubernetes configuration (future)
  kubernetes:
    enabled: false
    kubeconfig: "/path/to/kubeconfig"
    
  # Nomad configuration (future)
  nomad:
    enabled: false
    address: "http://localhost:4646"
    
  # Terraform configuration (future)
  terraform:
    enabled: false
    version: "1.5.0"
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

### Tenant Management
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List tenants
- `GET /api/v1/tenants/{id}` - Get tenant
- `PUT /api/v1/tenants/{id}` - Update tenant
- `DELETE /api/v1/tenants/{id}` - Delete tenant
- `GET /api/v1/tenants/{id}/resources` - Get tenant resource usage

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

- **Tenant Isolation**: Ensures deployments are completely separated
- **Network Isolation**: Prevents cross-tenant communication
- **Resource Limits**: Prevents resource exhaustion
- **Audit Logging**: Logs all operations for security auditing
- **Access Control**: Role-based access control for API endpoints
- **Secure Execution**: Secure command execution with proper isolation

## Performance

- **Efficient Resource Management**: Optimized resource allocation and monitoring
- **Real-time Streaming**: Efficient real-time log and event streaming
- **Caching**: Intelligent caching for frequently accessed data
- **Connection Pooling**: Efficient connection management for external services
- **Metrics Collection**: Low-overhead metrics collection and aggregation

## Contributing

To add a new deployment type:

1. Create a new directory under `core/deployments/`
2. Implement the required interfaces
3. Register the manager creator with the factory
4. Add configuration options
5. Update this README

## Status

- ✅ Docker deployment: Complete with full monitoring, logs, and exec support
- 🚧 Kubernetes deployment: Planned
- 🚧 Nomad deployment: Planned  
- 🚧 Terraform deployment: Planned

The Docker deployment module is fully functional and ready for production use with comprehensive monitoring, logging, and management capabilities. Other deployment types are planned for future releases. 