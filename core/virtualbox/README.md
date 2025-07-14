# VirtualBox Service

## Overview

The VirtualBox service provides a comprehensive API for managing Virtual Machines through VirtualBox. It allows you to create, manage, and control VMs programmatically with support for ARM64 architecture, automatic ISO downloads, and resource management.

## Features

- **VM Management**: Create, update, delete, and list Virtual Machines
- **VM Control**: Start, stop, pause, resume, and reset VMs
- **ISO Management**: Automatic download and management of ISO files
- **OS Type Management**: Intelligent OS type and ISO URL determination
- **Resource Monitoring**: Track VM resource usage and system information
- **Resource Validation**: Check system resources before VM creation to prevent failures
- **ARM64 Support**: Optimized for ARM64 architecture (Apple Silicon, etc.)
- **Multi-OS Support**: Support for Ubuntu, Debian, Fedora, CentOS, RedHat, Oracle Linux, Windows, FreeBSD, and more

## Architecture

The VirtualBox service is organized into several components:

### Core Components

1. **Service (`service.go`)**: Main service that coordinates all VirtualBox operations
2. **Client (`client.go`)**: Handles VBoxManage command execution
3. **Storage (`storage.go`)**: Manages ISO downloads and file operations
4. **Types (`types/types.go`)**: Defines data structures and interfaces
5. **API (`internal/api/virtualbox.go`)**: HTTP API layer for external access

### Key Interfaces

- `Service`: Main service interface for VM operations
- `VBoxClient`: Interface for VirtualBox CLI operations
- `StorageManager`: Interface for file storage operations
- `ISOOSTypeManager`: Interface for OS type and ISO URL determination

## Installation

### Prerequisites

1. **VirtualBox**: Install VirtualBox on your system

   - macOS: `brew install virtualbox`
   - Linux: `sudo apt-get install virtualbox`
   - Windows: Download from Oracle website

2. **Go**: Ensure Go 1.24+ is installed

### Validation

The VirtualBox service validates the installation once during service creation. This includes:

- Checking if VBoxManage is available in the PATH
- Verifying VirtualBox version
- Testing basic VirtualBox functionality

If validation fails, the service creation will return an error immediately.

### Setup

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Build the service: `go build ./cmd/subnet`

## Usage

### Basic VM Creation

```go
package main

import (
    "context"
    "log"

    "github.com/unicornultrafoundation/subnet-node/core/node/resource"
    "github.com/unicornultrafoundation/subnet-node/core/virtualbox"
    vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func main() {
    // Create VirtualBox service
    service, err := virtualbox.NewService(config)
    if err != nil {
        log.Fatal(err)
    }

    // Start the service
    ctx := context.Background()
    err = service.Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer service.Stop(ctx)

    // Create a new VM
    req := vbtypes.VMCreateRequest{
        Name:       "my-ubuntu-vm",
        CPUCores:   2,
        MemoryMB:   4096,
        DiskSizeGB: 20,
        ISOURL:     "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso",
    }

    vm, err := service.CreateVM(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created VM: %s", vm.Name)

    // Start the VM
    vm, err = service.StartVM(ctx, vm.ID)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("VM started: %s", vm.Name)
}
```

### API Usage

The service provides a REST API for external access:

```bash
# Create a VM
curl -X POST http://localhost:8080/api/virtualbox/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "cpu_cores": 2,
    "memory_mb": 4096,
    "disk_size_gb": 20
  }'

# List VMs
curl http://localhost:8080/api/virtualbox/vms

# Start a VM
curl -X POST http://localhost:8080/api/virtualbox/vms/{vm-id}/start

# Stop a VM
curl -X POST http://localhost:8080/api/virtualbox/vms/{vm-id}/stop
```

## Configuration

### Default Settings

The service uses sensible defaults for ARM64 systems:

- **OS Type**: Ubuntu_ARM64
- **Chipset**: armv8virtual
- **Firmware**: EFI
- **Graphics**: VMSVGA
- **Storage**: VirtioSCSI
- **Network**: NAT

### Customization

You can customize VM settings by modifying the `configureVMHardware` function in `service.go`:

```go
params := map[string]string{
    "memory":              strconv.Itoa(req.MemoryMB),
    "vram":                "16",
    "cpus":                strconv.Itoa(req.CPUCores),
    "chipset":             "armv8virtual",
    "firmware":            "efi",
    "graphicscontroller":  "vmsvga",
    // Add custom parameters here
}
```

### Resource Integration

The VirtualBox service integrates with the resource system for comprehensive resource management:

- **Direct Resource Access**: Uses the resource package directly for system resource information
- **Real-time Resource Detection**: Gets current system resources on demand
- **Comprehensive Resource Info**: CPU, memory, disk, and network information
- **Cross-platform Support**: Works on multiple operating systems
- **Error Handling**: Graceful fallback if resource detection fails

### Resource Validation

The service includes comprehensive resource validation before VM creation to prevent failures due to insufficient system resources:

- **CPU Validation**: Checks available CPU cores against VM requirements
- **Memory Validation**: Validates available RAM with 20% buffer for system overhead
- **Disk Space Validation**: Ensures sufficient disk space with 10% buffer
- **Resource Aggregation**: Considers resources used by existing VMs
- **Graceful Degradation**: Continues with VM creation if resource detection fails

```go
// Resource validation is automatically performed during VM creation
req := vbtypes.VMCreateRequest{
    Name:       "my-vm",
    CPUCores:   4,
    MemoryMB:   8192,
    DiskSizeGB: 50,
}

// This will automatically validate resources before creating the VM
vm, err := service.CreateVM(ctx, req)
if err != nil {
    // Error will include resource validation details if validation fails
    log.Printf("VM creation failed: %v", err)
}
```

### Storage Manager with ISO OS Type Management

The service includes an enhanced Storage Manager that handles both file storage operations and ISO OS type management:

- **File Storage**: Download, manage, and validate ISO files
- **Automatic OS Type Detection**: Determines the appropriate OS type based on system architecture
- **ISO URL Resolution**: Maps OS types to their corresponding ISO download URLs
- **Architecture Compatibility**: Validates OS type compatibility with the current system architecture
- **Configuration Integration**: Uses service configuration defaults when available

```go
// The storage manager is automatically created with the service
// Determine OS type and ISO URL for a VM request
osType, isoURL, err := service.storageMgr.DetermineOSTypeAndISO(ctx, req)

// Get supported OS types for current architecture
supportedTypes := service.storageMgr.GetSupportedOSTypes()

// Download and manage ISO files
err = service.storageMgr.DownloadFile(ctx, isoURL, destPath)
isoInfo, err := service.storageMgr.GetFileInfo(isoPath)
```

**Supported OS Types:**

- Ubuntu (64-bit, ARM64)
- Debian (64-bit, ARM64)
- Fedora (64-bit, ARM64)
- CentOS (64-bit, ARM64)
- RedHat (64-bit, ARM64)
- Oracle Linux (64-bit, ARM64)
- Windows (64-bit, ARM64)
- FreeBSD (64-bit, ARM64)
- NetBSD (64-bit, ARM64)

## File Structure

```
core/virtualbox/
├── types/
│   └── types.go          # Data structures and types
├── interfaces.go         # Service interfaces
├── client.go            # VirtualBox CLI client
├── storage.go           # File storage manager with ISO OS type management
├── service.go           # Main service implementation
├── service_test.go      # Service unit tests
├── storage_test.go      # Storage manager unit tests
└── README.md           # This file
```

## API Endpoints

### VM Management

- `POST /api/virtualbox/vms` - Create a new VM
- `GET /api/virtualbox/vms` - List VMs
- `GET /api/virtualbox/vms/{id}` - Get VM details
- `PUT /api/virtualbox/vms/{id}` - Update VM
- `DELETE /api/virtualbox/vms/{id}` - Delete VM

### VM Control

- `POST /api/virtualbox/vms/{id}/start` - Start VM
- `POST /api/virtualbox/vms/{id}/stop` - Stop VM
- `POST /api/virtualbox/vms/{id}/pause` - Pause VM
- `POST /api/virtualbox/vms/{id}/resume` - Resume VM
- `POST /api/virtualbox/vms/{id}/reset` - Reset VM

### Resource Management

- `GET /api/virtualbox/vms/{id}/usage` - Get VM usage
- `GET /api/virtualbox/usage` - Get all VM usage
- `GET /api/virtualbox/system` - Get system info

### ISO Management

- `POST /api/virtualbox/isos` - Download ISO
- `GET /api/virtualbox/isos` - List ISOs
- `GET /api/virtualbox/isos/{path}` - Get ISO info
- `DELETE /api/virtualbox/isos/{path}` - Delete ISO

## Testing

Run the tests to verify the service works:

```bash
# Run all tests
go test ./core/virtualbox/...

# Run service tests
go test -v ./core/virtualbox/ -run TestService

# Run storage manager tests
go test -v ./core/virtualbox/ -run TestStorageManager

# Run specific test
go test -v ./core/virtualbox/ -run TestCreateVM

# Run tests with verbose output
go test -v ./core/virtualbox/...
```

Note: Service tests require VirtualBox to be installed and will be skipped if not available. Storage manager tests can run without VirtualBox.

## Troubleshooting

### Common Issues

1. **VBoxManage not found**

   - Ensure VirtualBox is installed
   - Check PATH environment variable
   - Verify installation on macOS ARM: `/opt/homebrew/bin/VBoxManage`

2. **Permission denied**

   - Run with appropriate permissions
   - Check VirtualBox installation
   - Verify user has access to VirtualBox

3. **ISO download fails**

   - Check internet connection
   - Verify ISO URL is accessible
   - Check disk space for downloads

4. **VM creation fails**
   - Verify VirtualBox version compatibility
   - Check available system resources
   - Ensure proper ARM64 support

### Debug Mode

Enable debug logging by setting the log level:

```go
import "github.com/sirupsen/logrus"

logrus.SetLevel(logrus.DebugLevel)
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the same license as the main subnet-node project.
