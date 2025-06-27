# KVM Service

The KVM service provides virtual machine management capabilities with support for both real libvirt virtualization and simulation mode.

## Features

- **Configurable Modes**: Choose between auto-detection, real libvirt, or simulation mode
- **Resource Management**: Manage CPU, memory, and disk resources
- **VM Lifecycle**: Create, start, stop, and delete virtual machines
- **Cloud-init Support**: Automated VM provisioning with cloud-init
- **Ubuntu Cloud Images**: Automatic download and management of Ubuntu cloud images

## Configuration

The KVM service is configured through the main configuration system. Key configuration options:

### Basic Configuration

```yaml
kvm:
  enabled: true # Enable/disable KVM service
  mode: "auto" # auto, real, simulation
  max_vms: 5 # Maximum number of VMs
  max_cpu_cores: 4 # Maximum CPU cores per VM
  max_memory_mb: 4096 # Maximum memory per VM (MB)
  max_disk_gb: 50 # Maximum disk size per VM (GB)
```

### Libvirt Configuration (Real Mode)

```yaml
kvm:
  libvirt_uri: "qemu:///system" # Libvirt connection URI
  storage_pool: "subnet-vms" # Storage pool name
  storage_path: "/var/lib/libvirt/images/subnet" # Storage path
  network_name: "subnet-net" # Network name
  ubuntu_version: "22.04" # Ubuntu version for cloud images
  ssh_key_path: "/var/lib/libvirt/ssh/subnet-key" # SSH key path
```

## Modes

### Auto Mode (Default)

- Automatically detects if libvirt is available
- Falls back to simulation mode if libvirt is not available
- Best for development and flexible deployment

### Real Mode

- Forces real libvirt virtualization
- Requires libvirt to be installed and running
- Best for production environments

### Simulation Mode

- Uses simulation without real virtualization
- No external dependencies
- Best for testing and development

## Usage

### Starting the Service

The KVM service is automatically started when the node starts if enabled in configuration.

### API Usage

```go
// Create a VM
vm, err := kvmService.CreateVM(ctx, &kvm.CreateVMRequest{
    Name:     "test-vm",
    CPUCores: 2,
    MemoryMB: 2048,
    DiskGB:   20,
})

// Start a VM
err = kvmService.StartVM(ctx, vm.ID)

// Stop a VM
err = kvmService.StopVM(ctx, vm.ID)

// Delete a VM
err = kvmService.DeleteVM(ctx, vm.ID)

// List VMs
vms, err := kvmService.ListVMs(ctx)

// Get system resources
resources, err := kvmService.GetSystemResources(ctx)
```

## Requirements

### Real Mode Requirements

- libvirt installed and running
- qemu-kvm installed
- virsh command available
- Appropriate permissions for libvirt operations

### Simulation Mode Requirements

- No external dependencies
- Works on any system

## Migration from Build Tags

Previously, the KVM service used build tags (`-tags libvirt`) to determine whether to include real libvirt support. This has been replaced with configuration-based detection:

**Old approach:**

```bash
go build -tags libvirt ./cmd/subnet
```

**New approach:**

```yaml
kvm:
  mode: "auto" # or "real" or "simulation"
```

## Benefits of Configuration-Based Approach

1. **No Build Tags**: Single binary works in all environments
2. **Runtime Detection**: Automatically adapts to available resources
3. **Flexible Deployment**: Same binary can be used in development and production
4. **Easy Configuration**: Simple YAML configuration
5. **Graceful Degradation**: Falls back to simulation when libvirt unavailable

## Troubleshooting

### Libvirt Not Available

If you see "Libvirt not available" messages:

1. Check if libvirt is installed: `which virsh`
2. Check if libvirt daemon is running: `systemctl status libvirtd`
3. Check permissions: ensure user is in libvirt group
4. Use simulation mode for testing: `mode: "simulation"`

### Permission Issues

- Add user to libvirt group: `usermod -a -G libvirt $USER`
- Restart libvirt daemon: `systemctl restart libvirtd`
- Check SELinux/AppArmor policies if applicable

### Storage Issues

- Ensure storage path exists and is writable
- Check disk space availability
- Verify storage pool configuration

## Development

### Adding New Features

1. Extend the configuration in `config/default_config.go`
2. Update the service in `core/kvm/service.go`
3. Add stub implementations in `core/kvm/libvirt/libvirt.go`
4. Update tests and documentation

### Testing

- Use simulation mode for unit tests
- Use real mode for integration tests
- Test both modes in CI/CD pipeline
