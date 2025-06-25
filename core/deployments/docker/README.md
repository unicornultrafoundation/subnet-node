# Deployment Module

A module for managing Docker deployments using Docker Compose with customizable manifest capabilities and tenant isolation.

## Key Features

### 1. Tenant Management
- Create, update, delete tenants
- Complete isolation between tenants
- Resource usage management for each tenant

### 2. Deployment Management
- Create deployments from customizable manifests
- Start, stop, restart deployments
- Health checks and monitoring
- Log management

### 3. Port Management
- Dynamic port pool for tenants
- Avoid port conflicts between tenants
- Reserve port ranges for tenants

### 4. Network Management
- Create dedicated networks for each tenant
- Network isolation between tenants
- Custom network configuration

### 5. Resource Management
- Resource allocation and monitoring
- Resource limits per tenant
- Resource usage tracking

### 6. Manifest Management
- Validate manifests
- Process manifests for tenants
- Template-based manifest generation
- Merge multiple manifests

## Directory Structure

```
core/deployment/
├── types.go              # Data type definitions
├── interfaces.go         # Interface definitions
├── service.go           # Main service
├── managers/            # Manager implementations
│   ├── tenant_manager.go
│   ├── port_manager.go
│   ├── manifest_manager.go
│   ├── network_manager.go
│   ├── resource_manager.go
│   ├── event_manager.go
│   ├── compose_executor.go
│   └── storage_manager.go
├── templates/           # Sample templates
│   ├── nginx.yaml
│   ├── postgres.yaml
│   ├── redis.yaml
│   └── wordpress.yaml
└── README.md
```

## Configuration

### Main configuration file
```yaml
deployment:
  work_dir: "./deployments"
  port_pool:
    start: 10000
    end: 20000
  resource_limits:
    max_cpu_cores: 8
    max_memory_gb: 16
    max_disk_gb: 100
    max_containers: 10
    max_ports: 50
  network:
    default_driver: "bridge"
    subnet_pool:
      start: "172.16.0.0/12"
      end: "172.31.0.0/12"
```

## Usage

### 1. Initialize Service

```go
import (
    "github.com/unicornultrafoundation/subnet-node/config"
    "github.com/unicornultrafoundation/subnet-node/core/deployment"
    "github.com/unicornultrafoundation/subnet-node/core/deployment/managers"
)

// Initialize service
cfg := config.NewC(logger)
deploymentService := deployment.NewService(cfg, logger)

// Initialize managers
tenantManager := managers.NewTenantManager(logger, storage, resourceMgr)
portManager := managers.NewPortManager(logger, 10000, 20000)
manifestManager := managers.NewManifestManager(logger)
// ... initialize other managers

// Set managers for service
deploymentService.SetManagers(
    tenantManager,
    deploymentManager,
    manifestManager,
    portManager,
    networkManager,
    resourceManager,
    eventManager,
    composeExecutor,
    storageManager,
)

// Start service
deploymentService.Start(ctx)
```

### 2. Create Tenant

```go
tenant := &deployment.Tenant{
    ID:          "tenant-001",
    Name:        "Production Tenant",
    Description: "Production environment for customer A",
    Labels: map[string]string{
        "environment": "production",
        "customer":    "customer-a",
    },
}

err := deploymentService.CreateTenant(ctx, tenant)
```

### 3. Create Deployment

```go
// Create manifest
manifest := &deployment.ComposeManifest{
    Version: "3.8",
    Services: map[string]deployment.ServiceConfig{
        "web": {
            Image: "nginx:alpine",
            Ports: []deployment.PortMapping{
                {ContainerPort: 80, Protocol: "tcp"},
            },
            RestartPolicy: "unless-stopped",
        },
    },
}

// Create deployment
deployment := &deployment.Deployment{
    TenantID:    "tenant-001",
    Name:        "web-app",
    Description: "Web application deployment",
    Manifest:    *manifest,
}

err := deploymentService.CreateDeployment(ctx, deployment)
```

### 4. Start Deployment

```go
err := deploymentService.StartDeployment(ctx, deployment.ID)
```

### 5. Using Templates

```go
// Load template
templateData, err := os.ReadFile("templates/nginx.yaml")
if err != nil {
    return err
}

// Generate manifest from template
params := map[string]interface{}{
    "TenantID": "tenant-001",
    "Port":     "8080:80",
    "Domain":   "example.com",
    "Version":  "1.0.0",
}

manifest, err := manifestManager.GenerateManifest(ctx, string(templateData), params)
if err != nil {
    return err
}

// Create deployment with generated manifest
deployment := &deployment.Deployment{
    TenantID: "tenant-001",
    Name:     "nginx-app",
    Manifest: *manifest,
}

err = deploymentService.CreateDeployment(ctx, deployment)
```

## Templates

### Nginx Template
```yaml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "{{.Port}}"
    environment:
      - NGINX_HOST={{.Domain}}
    labels:
      tenant.id: "{{.TenantID}}"
```

### PostgreSQL Template
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    ports:
      - "{{.Port}}:5432"
    environment:
      - POSTGRES_DB={{.Database}}
      - POSTGRES_USER={{.Username}}
      - POSTGRES_PASSWORD={{.Password}}
```

## API Endpoints

### Tenant Management
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List tenants
- `GET /api/v1/tenants/{id}` - Get tenant details
- `PUT /api/v1/tenants/{id}` - Update tenant
- `DELETE /api/v1/tenants/{id}` - Delete tenant

### Deployment Management
- `POST /api/v1/tenants/{tenantId}/deployments` - Create deployment
- `GET /api/v1/tenants/{tenantId}/deployments` - List deployments
- `GET /api/v1/deployments/{id}` - Get deployment details
- `PUT /api/v1/deployments/{id}` - Update deployment
- `DELETE /api/v1/deployments/{id}` - Delete deployment
- `POST /api/v1/deployments/{id}/start` - Start deployment
- `POST /api/v1/deployments/{id}/stop` - Stop deployment
- `POST /api/v1/deployments/{id}/restart` - Restart deployment

### Resource Management
- `GET /api/v1/tenants/{id}/resources` - Get tenant resource usage
- `GET /api/v1/system/resources` - Get system resources
- `GET /api/v1/system/available` - Get available resources

## Monitoring

### Metrics
- `deployment_tenant_count` - Number of tenants
- `deployment_active_count` - Number of running deployments
- `deployment_resource_usage` - Resource usage per tenant
- `deployment_port_usage` - Port usage

### Health Checks
- Container health checks
- Service availability checks
- Resource usage monitoring
- Network connectivity checks

## Security

### Tenant Isolation
- Network isolation between tenants
- Resource limits per tenant
- Port isolation
- Volume isolation

### Access Control
- Tenant-based access control
- Resource quota enforcement
- Image registry restrictions
- Security policy enforcement

## Backup & Recovery

### Backup
- Automatic backup scheduling
- Configuration backup
- Data backup
- Template backup

### Recovery
- Deployment restoration
- Configuration recovery
- Data recovery
- Disaster recovery

## Troubleshooting

### Common Issues

1. **Port Conflict**
   - Check port pool configuration
   - Verify port allocation logic
   - Check for port leaks

2. **Resource Exhaustion**
   - Monitor resource usage
   - Check resource limits
   - Optimize resource allocation

3. **Network Issues**
   - Verify network configuration
   - Check network isolation
   - Validate subnet allocation

4. **Template Issues**
   - Validate template syntax
   - Check parameter substitution
   - Verify template variables

### Logs
- Service logs: `./logs/deployment.log`
- Container logs: Available via API
- System logs: System journal

### Debug Mode
```yaml
deployment:
  logging:
    level: "debug"
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Implement changes
4. Add tests
5. Submit pull request

## License

This module is part of the subnet-node project and follows the same license. 