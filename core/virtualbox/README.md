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

## Template VM Creation Flow

The service uses pre-built OVA templates to create new VMs quickly. The flow for template VM creation is:

1. **Template Preparation**: Pre-configured OVA templates are stored in `~/VirtualBox VMs/Templates/`
2. **Template Download**: If a template doesn't exist locally, it's automatically downloaded from the configured URL
3. **OVA Import**: The template is imported to create a new VM
4. **Configuration Customization**: The VM is customized with requested CPU, memory, and disk settings
5. **Cloud-Init Setup**: A cloud-init ISO is generated to configure the VM on first boot
6. **Network Configuration**: SSH port forwarding is set up for remote access

### OVA Template URLs

The OVA template URLs are defined in `ova_manage.go` and currently include:

- **Ubuntu_64**: For x86_64/AMD64 architectures
- **Ubuntu_ARM64**: For ARM64/AArch64 architectures (e.g., Apple Silicon)

These URLs are defined in the `defaultOVAURLs` map and can be accessed using the `getOVAURLForOSType` function.

To add or update template URLs, modify the `defaultOVAURLs` map in `ova_manage.go`:

```go
var defaultOVAURLs = map[string]string{
    "Ubuntu_64":    "https://example.com/path/to/ubuntu_64.ova",
    "Ubuntu_ARM64": "https://example.com/path/to/ubuntu_arm64.ova",
    // Add more templates here
}
```

### Creating Your Own Templates

To create your own template OVA:

1. Use the `createTemplateVM` API to create a base VM:

   ```go
   req := vbtypes.VMCreateRequest{
       Name:       "template-base",
       OSType:     "Ubuntu_64",
       CPUCores:   2,
       MemoryMB:   2048,
       DiskSizeGB: 20,
       Username:   "admin",
       Password:   "template-password",
   }

   vm, err := service.CreateVM(ctx, req)
   ```

2. Start the VM and install any necessary software and configurations

   ```go
   vm, err = service.StartVM(ctx, vm.ID)
   ```

3. Connect to the VM via SSH (using the assigned SSH port)

   ```
   ssh admin@localhost -p <vm.SSHPort>
   ```

4. Clean the cloud-init data to ensure the template is pristine:

   ```
   sudo cloud-init clean
   sudo rm -rf /var/lib/cloud/*
   ```

5. Shut down the VM:

   ```go
   vm, err = service.StopVM(ctx, vm.ID)
   ```

6. Export the VM as an OVA file:

   ```
   VBoxManage export <vm-name> -o template_sample_<OSType>.ova
   ```

7. Place the OVA in `~/VirtualBox VMs/Templates/`

8. Update the `defaultOVAURLs` map in `ova_manage.go` if you want to distribute it

This process ensures that when new VMs are created from your template, they will get fresh cloud-init initialization without any remnants from the template creation process.

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
