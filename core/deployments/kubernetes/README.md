# Kubernetes Deployment Module

This module provides Kubernetes deployment management capabilities for the subnet-node project.

## Features

- Kubernetes cluster management
- Pod deployment and scaling
- Service and ingress management
- ConfigMap and Secret management
- Namespace isolation per tenant
- Resource quotas and limits
- Health checks and monitoring

## Planned Implementation

- Kubernetes client integration
- YAML manifest processing
- Helm chart support
- Multi-cluster management
- RBAC and security policies
- Resource monitoring and metrics

## Usage

```go
// Example usage (to be implemented)
import (
    "github.com/unicornultrafoundation/subnet-node/core/deployments"
    "github.com/unicornultrafoundation/subnet-node/core/deployments/kubernetes"
)

// Create Kubernetes deployment
deployment := &deployments.Deployment{
    Type: deployments.DeploymentTypeKubernetes,
    Name: "my-app",
    Manifest: &kubernetes.K8sManifest{
        // Kubernetes-specific manifest
    },
}
```

## Configuration

```yaml
deployments:
  kubernetes:
    enabled: true
    kubeconfig: "/path/to/kubeconfig"
    context: "default"
    namespace_prefix: "tenant-"
    resource_quotas:
      enabled: true
      default_limits:
        cpu: "4"
        memory: "8Gi"
        pods: "10"
```

## Status

🚧 **Under Development** - This module is planned for future implementation. 