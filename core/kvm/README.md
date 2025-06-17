# KVM Module for Subnet Node

A simple, focused KVM (Kernel-based Virtual Machine) module that provides basic virtual machine management capabilities. This module is designed to be easy to understand and extend incrementally.

## Overview

This KVM module provides:

- **Basic VM Lifecycle Management**: Create, start, stop, delete VMs
- **Resource Management**: CPU, memory, and disk allocation with limits
- **System Monitoring**: VM statistics and resource usage tracking
- **REST API**: HTTP endpoints for all operations
- **Persistent Storage**: VM metadata stored in datastore
- **Configuration**: YAML-based configuration with sensible defaults

## Architecture

```
core/kvm/
├── types.go           # VM data structures and types
├── service.go         # Main KVM service implementation
├── config/
│   ├── config.go      # Configuration management
│   └── validator.go   # Configuration validation
└── README.md          # This documentation
```

### Key Components

1. **Service**: Main orchestrator that manages VM lifecycle
2. **Types**: Define VM structures, status, and requests
3. **Configuration**: Handles YAML config loading with defaults
4. **API**: HTTP REST endpoints for VM operations

## Configuration

Add to your `config.yaml`:

```yaml
kvm:
  enabled: true # Enable/disable KVM service
  max_vms: 5 # Maximum number of VMs
  max_cpu_cores: 4 # Maximum CPU cores per VM
  max_memory_mb: 4096 # Maximum memory per VM (MB)
  max_disk_gb: 50 # Maximum disk size per VM (GB)
```

## Usage

### 1. Service Integration

```go
import (
    "github.com/unicornultrafoundation/subnet-node/core/kvm"
    "github.com/unicornultrafoundation/subnet-node/internal/api"
)

// Create KVM service
kvmService := kvm.NewService(config, logger, datastore, resourceService)

// Start the service
err := kvmService.Start(ctx)

// Register API routes
kvmAPI := api.NewKVMAPI(kvmService, logger)
kvmAPI.RegisterRoutes(router)
```

### 2. HTTP API Usage

#### Create a VM

```bash
curl -X POST http://localhost:8080/kvm/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "cpu_cores": 2,
    "memory_mb": 2048,
    "disk_gb": 20,
    "metadata": {
      "purpose": "development"
    }
  }'
```

#### List VMs

```bash
curl http://localhost:8080/kvm/vms
```

#### Get VM Details

```bash
curl http://localhost:8080/kvm/vms/{vm-id}
```

#### Start/Stop VM

```bash
curl -X POST http://localhost:8080/kvm/vms/{vm-id}/start
curl -X POST http://localhost:8080/kvm/vms/{vm-id}/stop
```

#### Get VM Statistics

```bash
curl http://localhost:8080/kvm/vms/{vm-id}/stats
```

#### Get System Resources

```bash
curl http://localhost:8080/kvm/resources
```

#### Check Service Status

```bash
curl http://localhost:8080/kvm/status
```

### 3. Programmatic Usage

```go
// Create a VM
req := &kvm.CreateVMRequest{
    Name:     "test-vm",
    CPUCores: 2,
    MemoryMB: 1024,
    DiskGB:   20,
    Metadata: map[string]string{
        "env": "testing",
    },
}

vm, err := kvmService.CreateVM(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Start the VM
err = kvmService.StartVM(ctx, vm.ID)

// Get VM stats
stats, err := kvmService.GetVMStats(ctx, vm.ID)
```

## VM Lifecycle

1. **Create**: VM is created in `stopped` state with allocated resources
2. **Start**: VM transitions to `starting` then `running` state
3. **Stop**: VM transitions to `stopping` then `stopped` state
4. **Delete**: VM must be stopped before deletion

## Features

### Current Implementation

- ✅ In-memory VM management (simulated)
- ✅ Resource validation and limits
- ✅ VM state management with status tracking
- ✅ Persistent metadata storage in datastore
- ✅ Background statistics collection
- ✅ Complete REST API
- ✅ Configuration management
- ✅ Proper error handling and logging

### Future Enhancements

- 🔄 Real libvirt integration for actual VM management
- 🔄 Network configuration and IP management
- 🔄 VM template system for OS images
- 🔄 Disk image management and storage pools
- 🔄 VM snapshots and backups
- 🔄 Resource monitoring and alerts
- 🔄 VM migration capabilities

## Error Handling

The module includes comprehensive error handling:

- Resource validation (CPU, memory, disk limits)
- VM state validation (can't delete running VM)
- Service availability checks
- Proper HTTP status codes in API responses

## Monitoring

The service includes built-in monitoring:

- VM statistics collection every 30 seconds
- Resource usage tracking
- System resource availability
- Service health checks

## Development

### Testing the Module

1. **Enable in config**:

```yaml
kvm:
  enabled: true
  max_vms: 3
```

2. **Start the service and test**:

```bash
# Check status
curl http://localhost:8080/kvm/status

# Create a test VM
curl -X POST http://localhost:8080/kvm/vms \
  -H "Content-Type: application/json" \
  -d '{"name":"test","cpu_cores":1,"memory_mb":512,"disk_gb":10}'

# List VMs
curl http://localhost:8080/kvm/vms

# Start the VM
curl -X POST http://localhost:8080/kvm/vms/{vm-id}/start
```

### Extending the Module

This simple foundation can be extended by:

1. Adding real libvirt integration
2. Implementing network management
3. Adding template/image management
4. Enhancing monitoring and metrics
5. Adding security features

The modular design makes it easy to enhance specific components without affecting others.
