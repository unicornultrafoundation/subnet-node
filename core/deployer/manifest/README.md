# Deployer Manifest Package

The `core/deployer/manifest` package provides a Service Definition Language (SDL) for defining containerized applications and their deployment configurations. This package handles parsing, validation, and conversion of SDL files to Kubernetes deployment manifests.

## Overview

The SDL is a declarative language that allows you to define:
- **Services**: Containerized applications with their configurations
- **Profiles**: Resource requirements and compute specifications
- **Deployments**: How services should be deployed
- **Endpoints**: External access points for your services

## SDL Structure

### Basic SDL Format

```yaml
version: "2.0"
services:
  # Service definitions
profiles:
  # Resource profiles
deployment:
  # Deployment configurations
endpoints:
  # External endpoints
```

### Version

The SDL version must be specified. Currently supports version `"2.0"`.

```yaml
version: "2.0"
```

## Services

Services define containerized applications that will be deployed.

### Service Structure

```yaml
services:
  service-name:
    image: "container-image:tag"
    command: ["command", "args"]
    args: ["additional", "arguments"]
    env:
      - "KEY=value"
      - "ANOTHER_KEY=another_value"
    count: 1
    expose:
      - port: 8080
        as: 80
        proto: "TCP"
        to:
          - global: true
    depends-on:
      - "other-service"
    params:
      health:
        readiness:
          http:
            path: "/health"
            port: 8080
          initial_delay_seconds: 5
          period_seconds: 10
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
        liveness:
          http:
            path: "/health"
            port: 8080
          initial_delay_seconds: 15
          period_seconds: 10
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
    resources:
      cpu:
        units:
          value: 100
          unit: "m"
      memory:
        size:
          value: 128
          unit: "Mi"
      storage:
        size:
          value: 1
          unit: "Gi"
      gpu:
        units: 1
        attributes:
          vendor:
            nvidia:
              - model: "A100"
                ram: "40Gi"
                interface: "PCIe"
    credentials:
      host: "registry.example.com"
      username: "user"
      password: "password"
    imagePullSecrets:
      - name: "registry-secret"
    scheduler_params:
      runtime_class: "runsc"
```

### Service Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | Container image to deploy |
| `command` | []string | No | Command to run in container |
| `args` | []string | No | Arguments for the command |
| `env` | []string | No | Environment variables (KEY=value format) |
| `count` | int32 | Yes | Number of replicas to deploy |
| `expose` | []Expose | No | Port exposure configuration |
| `depends-on` | []string | No | Service dependencies |
| `params` | ServiceParams | No | Service parameters (health checks, storage) |
| `resources` | ServiceResources | No | Resource requirements |
| `credentials` | RegistryAuth | No | Container registry authentication |
| `imagePullSecrets` | []ImagePullSecret | No | Kubernetes image pull secrets |
| `scheduler_params` | SchedulerParams | No | Scheduler-specific parameters |

### Expose Configuration

```yaml
expose:
  - port: 8080        # Container port
    as: 80            # Service port
    proto: "TCP"      # Protocol (TCP/UDP)
    to:
      - global: true  # Expose globally
      - service: "other-service"  # Expose to specific service
    accept:
      - "192.168.1.0/24"  # IP whitelist
    http_options:
      max_body_size: 1048576  # Max request body size
      next_cases: ["case1", "case2"]  # Additional HTTP options
```

### Health Checks

```yaml
params:
  health:
    readiness:
      http:
        path: "/health"
        port: 8080
      initial_delay_seconds: 5
      period_seconds: 10
      timeout_seconds: 5
      success_threshold: 1
      failure_threshold: 3
    liveness:
      http:
        path: "/health"
        port: 8080
      initial_delay_seconds: 15
      period_seconds: 10
      timeout_seconds: 5
      success_threshold: 1
      failure_threshold: 3
```

### Resource Requirements

```yaml
resources:
  cpu:
    units:
      value: 100
      unit: "m"  # millicores
  memory:
    size:
      value: 128
      unit: "Mi"  # MiB
  storage:
    size:
      value: 1
      unit: "Gi"  # GiB
  gpu:
    units: 1
    attributes:
      vendor:
        nvidia:
          - model: "A100"
            ram: "40Gi"
            interface: "PCIe"
  network:
    bandwidth: "100Mbps"
```

## Profiles

Profiles define compute resources and configurations that can be applied to services.

### Profile Structure

```yaml
profiles:
  compute:
    profile-name:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
        storage:
          - name: "data"
            size: "1Gi"
            attributes:
              persistent: false
              class: "standard"
        gpu:
          units: 1
          attributes:
            vendor:
              nvidia:
                - model: "A100"
                  ram: "40Gi"
                  interface: "PCIe"
        network:
          bandwidth: "100Mbps"
  placement:
    profile-name:
      attributes:
        region: "us-west"
        zone: "us-west-1a"
      pricing:
        cpu:
          denom: "u2u"
          amount: 1000
        memory:
          denom: "u2u"
          amount: 500
        storage:
          denom: "u2u"
          amount: 200
```

### Compute Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cpu.request` | string | Yes | CPU request (e.g., "100m") |
| `cpu.limit` | string | Yes | CPU limit (e.g., "200m") |
| `memory.request` | string | Yes | Memory request (e.g., "128Mi") |
| `memory.limit` | string | Yes | Memory limit (e.g., "256Mi") |
| `storage` | []StorageVolume | No | Storage volumes |
| `gpu` | GPUResource | No | GPU requirements |
| `network` | NetworkResource | No | Network requirements |

### Placement Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `attributes` | map | No | Placement attributes (region, zone, etc.) |
| `pricing` | map | No | Resource pricing information |

## Deployment

Deployment configuration defines how services should be deployed.

### Deployment Structure

```yaml
deployment:
  deployment-name:
    profile: "profile-name"
    count: 1
```

### Deployment Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `profile` | string | Yes | Compute profile to use |
| `count` | int32 | Yes | Number of instances to deploy |

## Endpoints

Endpoints define external access points for your services.

### Endpoint Structure

```yaml
endpoints:
  endpoint-name:
    kind: "http"
```

### Endpoint Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kind` | string | Yes | Endpoint type (e.g., "http") |

## Complete Example

Here's a complete SDL example for a web application:

```yaml
version: "2.0"

services:
  web:
    image: "nginx:latest"
    count: 2
    expose:
      - port: 80
        as: 80
        proto: "TCP"
        to:
          - global: true
    params:
      health:
        readiness:
          http:
            path: "/"
            port: 80
          initial_delay_seconds: 5
          period_seconds: 10
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
        liveness:
          http:
            path: "/"
            port: 80
          initial_delay_seconds: 15
          period_seconds: 10
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
    resources:
      cpu:
        units:
          value: 100
          unit: "m"
      memory:
        size:
          value: 128
          unit: "Mi"

  api:
    image: "myapp/api:latest"
    count: 3
    depends-on:
      - "web"
    expose:
      - port: 8080
        as: 80
        proto: "TCP"
        to:
          - global: true
    env:
      - "DATABASE_URL=postgresql://user:pass@db:5432/mydb"
      - "REDIS_URL=redis://redis:6379"
    params:
      health:
        readiness:
          http:
            path: "/health"
            port: 8080
          initial_delay_seconds: 10
          period_seconds: 30
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
    resources:
      cpu:
        units:
          value: 200
          unit: "m"
      memory:
        size:
          value: 256
          unit: "Mi"

  db:
    image: "postgres:13"
    count: 1
    expose:
      - port: 5432
        as: 5432
        proto: "TCP"
        to:
          - service: "api"
    env:
      - "POSTGRES_DB=mydb"
      - "POSTGRES_USER=user"
      - "POSTGRES_PASSWORD=pass"
    resources:
      cpu:
        units:
          value: 500
          unit: "m"
      memory:
        size:
          value: 512
          unit: "Mi"
      storage:
        size:
          value: 10
          unit: "Gi"

profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
        storage:
          - name: "data"
            size: "1Gi"
            attributes:
              persistent: false
    database:
      resources:
        cpu:
          request: "500m"
          limit: "1000m"
        memory:
          request: "512Mi"
          limit: "1Gi"
        storage:
          - name: "data"
            size: "10Gi"
            attributes:
              persistent: true
              class: "ssd"
  placement:
    default:
      attributes:
        region: "us-west"
      pricing:
        cpu:
          denom: "u2u"
          amount: 1000
        memory:
          denom: "u2u"
          amount: 500
        storage:
          denom: "u2u"
          amount: 200

deployment:
  default:
    profile: "default"
    count: 1
  database:
    profile: "database"
    count: 1

endpoints:
  web:
    kind: "http"
  api:
    kind: "http"
```

## Validation Rules

### Service Names
- Must be valid Kubernetes resource names
- Regex: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
- Examples: `web`, `api-server`, `my-app`

### Endpoint Names
- Must be valid Kubernetes resource names
- Regex: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
- Examples: `web`, `api`, `admin`

### Resource Values
- CPU: Must be positive, units in millicores (m)
- Memory: Must be positive, units in MiB or GiB
- Storage: Must be positive, units in MiB, GiB, or TiB
- GPU: Must be positive integer

### Health Checks
- Path: Must be non-empty string
- Port: Must be positive integer
- Timeouts: Must be positive integers
- Thresholds: Must be positive integers

## Usage

### Parsing SDL Files

```go
package main

import (
    "github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
)

func main() {
    parser := manifest.NewParser()
    
    // Parse YAML file
    sdl, err := parser.ParseFile("deployment.yaml")
    if err != nil {
        panic(err)
    }
    
    // Convert to manifest
    manifest, err := sdl.ToManifest()
    if err != nil {
        panic(err)
    }
    
    // Use the manifest for deployment
    // ...
}
```

### Validation

```go
// Validate SDL
err := sdl.Validate()
if err != nil {
    panic(err)
}

// Get version hash for deterministic deployments
versionHash, err := sdl.Version()
if err != nil {
    panic(err)
}
```

## File Formats

The SDL supports both YAML and JSON formats:

### YAML Format
```yaml
version: "2.0"
services:
  web:
    image: "nginx:latest"
    count: 1
```

### JSON Format
```json
{
  "version": "2.0",
  "services": {
    "web": {
      "image": "nginx:latest",
      "count": 1
    }
  }
}
```

## Examples

See the `examples/` directory for complete SDL examples:

- `minimal.yaml` - Minimal deployment configuration
- `example.yaml` - Basic web application
- `complex.yaml` - Multi-service application with dependencies
- `network.yaml` - Network configuration examples
- `echo.json` - Simple HTTP echo service

## Best Practices

1. **Use descriptive service names** that reflect the application purpose
2. **Define health checks** for all services to ensure proper monitoring
3. **Set appropriate resource limits** to prevent resource exhaustion
4. **Use profiles** to standardize resource configurations across services
5. **Define dependencies** to ensure proper startup order
6. **Use environment variables** for configuration instead of hardcoding values
7. **Set up proper endpoints** for external access
8. **Use persistent storage** for databases and stateful applications

## Troubleshooting

### Common Issues

1. **Invalid service names**: Ensure service names follow Kubernetes naming conventions
2. **Missing required fields**: Check that all required fields are specified
3. **Invalid resource values**: Ensure resource values are positive and use correct units
4. **Circular dependencies**: Avoid circular dependencies between services
5. **Invalid health check paths**: Ensure health check paths are valid for your application

### Validation Errors

The parser provides detailed error messages for validation failures. Common error messages include:

- `"invalid service name"` - Service name doesn't match required pattern
- `"service image is required"` - Missing image specification
- `"invalid port number"` - Port must be positive integer
- `"invalid protocol"` - Protocol must be TCP or UDP
- `"compute profile not found"` - Referenced profile doesn't exist 