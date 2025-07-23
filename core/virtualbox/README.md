# VirtualBox Package

This package provides a comprehensive interface for managing VirtualBox virtual machines programmatically. It allows for creating, managing, and controlling VMs with support for various operating systems and hardware configurations.

## Core Components

### VirtualBox Service

The `ServiceImpl` is the main entry point for interacting with VirtualBox. It implements the `Service` interface and provides methods for:

- Creating and managing VMs
- Starting, stopping, pausing, resuming, and resetting VMs
- Retrieving VM information and system resources
- Managing VM storage and configuration

### VBoxManage Executor

The `VBoxManageExecutor` provides a wrapper around the VirtualBox command-line interface (`VBoxManage`). It handles:

- Executing VBoxManage commands
- Parsing command output
- Managing VM configuration
- Setting up storage, networking, and hardware

### Hardware Detector

The `HardwareDetector` automatically detects system hardware capabilities and configures VMs accordingly:

- Detects CPU architecture (x86_64, ARM64, etc.)
- Identifies available memory and storage
- Detects GPU capabilities
- Determines optimal VM settings based on hardware

## Features

### VM Management

- **Create VM**: Create new VMs with customizable hardware configurations
- **Start/Stop VM**: Control VM power state
- **Pause/Resume VM**: Suspend and resume VM execution
- **Reset VM**: Restart a VM
- **Delete VM**: Remove a VM and its associated files
- **Update VM**: Modify VM hardware configuration

### Cloud-Init Integration

- **Template-based Configuration**: Generate cloud-init configurations from templates
- **User Data Customization**: Set hostname, username, and password
- **ISO Generation**: Automatically create cloud-init ISO images for VM initialization
- **Different Templates**: Support for both new VM and cloned VM configurations

### SSH Port Forwarding

- **Automatic Port Assignment**: Dynamically find available host ports
- **NAT Configuration**: Set up NAT rules for SSH access
- **Port Management**: Track and manage forwarded ports

### Resource Management

- **Resource Validation**: Check system resources before VM creation
- **Overcommit Prevention**: Prevent resource exhaustion
- **Hardware Compatibility**: Ensure VM configurations match host capabilities

## Prerequisites

### Required Software

1. **VirtualBox**: Must be installed and accessible via `VBoxManage` command

   ```
   # macOS
   brew install virtualbox

   # Ubuntu/Debian
   sudo apt-get install virtualbox
   ```

2. **ISO Generation Tools**: Required for cloud-init configuration

   ```
   # macOS
   brew install cdrtools

   # Ubuntu/Debian
   sudo apt-get install genisoimage
   ```

### Architecture Considerations

- **x86_64/AMD64**: Full support for all VirtualBox features
- **ARM64/AArch64**: Specialized configurations for Apple Silicon and other ARM platforms
- **Other Architectures**: Limited support with fallback configurations

### OVA Templates

Run the download script to fetch pre-configured OVA templates:

```
cd core/virtualbox/script
go run download_ovas.go
```

This will download OS templates to `~/VirtualBox VMs/Templates/`.

## VM Creation Strategy

The service follows a streamlined approach to VM creation:

1. **Resource Validation**: Check if sufficient system resources are available
2. **OVA Template Selection**: Choose appropriate OVA based on requested OS type
3. **OVA Import**: Import the OVA template to create a base VM
4. **Hardware Configuration**: Adjust CPU, memory, and other settings to match request
5. **Cloud-Init Configuration**: Generate custom cloud-init ISO with user credentials
6. **Network Configuration**: Set up SSH port forwarding for remote access
7. **VM Startup**: Boot the VM with the new configuration

This approach provides faster VM creation compared to installing from scratch, with consistent configuration across deployments.

## Usage Examples

### Creating a VM from OVA Template

```go
req := vbtypes.VMCreateRequest{
    Name:       "ubuntu-vm",
    OSType:     "Ubuntu_64",
    CPUCores:   2,
    MemoryMB:   2048,
    DiskSizeGB: 20,
    Username:   "admin",
    Password:   "secure-password",
}

vm, err := service.CreateAndStartVM(ctx, req)
if err != nil {
    log.Fatalf("Failed to create VM: %v", err)
}

log.Printf("VM created successfully: %s, SSH Port: %d", vm.Name, vm.SSHPort)
```

### Managing VMs

```go
// Get VM information
vm, err := service.GetVM(ctx, vmID)

// Start VM
vm, err = service.StartVM(ctx, vmID)

// Stop VM
vm, err = service.StopVM(ctx, vmID)

// Delete VM
err = service.DeleteVM(ctx, vmID)
```

## Architecture-Specific Considerations

The package automatically detects and adapts to different CPU architectures:

- **Apple Silicon (ARM64)**: Uses ARM-specific settings for chipset, firmware, and graphics
- **Intel/AMD (x86_64)**: Uses standard VirtualBox settings optimized for x86 architecture
- **Other ARM platforms**: Configures VMs with appropriate ARM virtualization settings

Each architecture receives optimized settings for:

- Chipset selection
- Firmware type (EFI vs BIOS)
- Graphics controller
- USB controller
- Audio configuration
