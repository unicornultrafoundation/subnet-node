# Terraform Deployment Module

This module provides Terraform deployment management capabilities for the subnet-node project.

## Features

- Infrastructure as Code (IaC) deployment
- Terraform state management
- Multi-cloud deployment support
- Resource provisioning and management
- Configuration drift detection
- Plan and apply workflows
- Remote state storage

## Planned Implementation

- Terraform CLI integration
- HCL configuration processing
- State file management
- Workspace isolation per tenant
- Variable and secret management
- Module registry integration
- Cost estimation and optimization

## Usage

```go
// Example usage (to be implemented)
import (
    "github.com/unicornultrafoundation/subnet-node/core/deployments"
    "github.com/unicornultrafoundation/subnet-node/core/deployments/terraform"
)

// Create Terraform deployment
deployment := &deployments.Deployment{
    Type: deployments.DeploymentTypeTerraform,
    Name: "my-infrastructure",
    Manifest: &terraform.TerraformConfig{
        // Terraform-specific configuration
    },
}
```

## Configuration

```yaml
deployments:
  terraform:
    enabled: true
    version: "1.5.0"
    working_directory: "/opt/terraform"
    state_backend:
      type: "s3"
      bucket: "terraform-state"
      key: "subnet-node"
    providers:
      aws:
        region: "us-west-2"
      azure:
        subscription_id: "xxx"
      gcp:
        project: "my-project"
```

## Status

🚧 **Under Development** - This module is planned for future implementation. 