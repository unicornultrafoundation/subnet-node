# KVM Service

The KVM service provides virtual machine management capabilities for the Subnet Node, supporting both real KVM virtualization and simulation mode.

## Features

- **Real KVM Support**: Full virtualization using libvirt and KVM/QEMU
- **Simulation Mode**: Fallback mode when libvirt is not available
- **Cloud-init Integration**: Automated VM provisioning with SSH key injection
- **Resource Management**: CPU, memory, and disk allocation
- **Network Management**: Automatic network configuration
- **Permission Management**: Comprehensive file ownership and permission handling

## Configuration

The KVM service can be configured through the main configuration file:

```yaml
kvm:
  enabled: true # Enable/disable KVM service
  mode: "auto" # auto, real, simulation
  max_vms: 5 # Maximum number of VMs
  max_cpu_cores: 4 # Maximum CPU cores per VM
  max_memory_mb: 4096 # Maximum memory per VM (MB)
  max_disk_gb: 50 # Maximum disk size per VM (GB)
  libvirt_uri: "qemu:///system" # Libvirt connection URI
  storage_pool: "subnet-vms" # Storage pool name
  storage_path: "/var/lib/libvirt/images/subnet" # Storage path
  network_name: "subnet-net" # Network name
  ubuntu_version: "22.04" # Ubuntu version for VMs
  ssh_key_path: "/var/lib/libvirt/ssh/subnet-key" # SSH key path
```

## Modes

### Auto Mode (Default)

The service automatically detects if libvirt is available and falls back to simulation mode if needed.

### Real Mode

Forces the use of real KVM virtualization. Will fail if libvirt is not available.

### Simulation Mode

Uses simulation mode regardless of libvirt availability. Useful for testing and development.

## Common Permission Issues

The most common issue with the KVM service is permission problems with VM disk files. This happens because:

1. **File Ownership**: VM disk files need to be accessible by the libvirt daemon
2. **Group Permissions**: The user needs to be in the `libvirt` and `kvm` groups
3. **Directory Permissions**: Storage directories need proper permissions
4. **KVM Device Access**: The `/dev/kvm` device needs proper permissions

### Symptoms

- `Permission denied` errors when starting VMs
- `Failed to start domain` errors
- VM disk files owned by `libvirt-qemu` user
- Cannot access VM disk files

### Solutions

#### 1. Run the Permission Fix Script

The easiest solution is to run the provided permission fix script:

```bash
chmod +x scripts/fix_libvirt_permissions.sh
./scripts/fix_libvirt_permissions.sh
```

This script will:

- Add your user to the `libvirt` and `kvm` groups
- Fix permissions on VM disk files
- Fix directory permissions
- Check and start the libvirt daemon
- Test libvirt connectivity

#### 2. Manual Permission Fixes

If the script doesn't work, you can manually fix permissions:

```bash
# Add user to required groups
sudo usermod -a -G libvirt,kvm $USER

# Fix VM disk file permissions
sudo chown $USER:libvirt /path/to/vm-disk.qcow2
sudo chmod 660 /path/to/vm-disk.qcow2

# Fix directory permissions
sudo chown -R $USER:libvirt /path/to/storage/directory
sudo chmod -R 770 /path/to/storage/directory

# Fix KVM device permissions
sudo chmod 666 /dev/kvm

# Restart libvirt daemon
sudo systemctl restart libvirtd
```

#### 3. Log Out and Back In

After adding yourself to groups, you need to log out and back in for the changes to take effect:

```bash
# Check current groups
groups

# If libvirt and kvm are not listed, log out and back in
exit
# Then log back in and check again
groups
```

## Troubleshooting

### Diagnostic Function

The KVM service includes a diagnostic function to help identify permission issues:

```go
diagnosis, err := kvmService.DiagnosePermissionIssues(ctx)
if err != nil {
    log.Fatal(err)
}

// Print diagnosis as JSON
diagnosisJSON, _ := json.MarshalIndent(diagnosis, "", "  ")
fmt.Println(string(diagnosisJSON))
```

This will provide detailed information about:

- File and directory permissions
- User group membership
- Libvirt daemon status
- KVM device accessibility
- Specific permission issues and recommendations

### Common Error Messages

#### "Permission denied"

- **Cause**: VM disk files not accessible by libvirt
- **Solution**: Fix file permissions and ensure user is in libvirt group

#### "No such file or directory"

- **Cause**: VM disk files don't exist or wrong path
- **Solution**: Check storage path configuration and file existence

#### "Domain already exists"

- **Cause**: VM with same name already exists
- **Solution**: Delete existing VM or use different name

#### "Libvirt not available"

- **Cause**: Libvirt daemon not running or not accessible
- **Solution**: Start libvirt daemon and check user permissions

### Log Analysis

Enable debug logging to get detailed information:

```yaml
logging:
  level: debug
  kvm: debug
```

Look for these log messages:

- `"Creating VM disk file with correct ownership"`
- `"Fixing permissions for disk file"`
- `"Permission denied detected, attempting comprehensive permission fix"`

## File Structure

```
~/.subnet/libvirt/
├── images/                    # VM disk files (.qcow2)
├── cloud-init/               # Cloud-init ISO files
└── ssh/                      # SSH keys
```

## Security Considerations

- VM disk files should have permissions 660 (user and group read/write)
- SSH keys should have permissions 600 (user read/write only)
- Storage directories should have permissions 770 (user and group read/write/execute)
- The user should be in the `libvirt` and `kvm` groups
- Avoid using world-readable permissions (666, 777) for security

## Development

### Testing

To test the KVM service without real virtualization:

```yaml
kvm:
  enabled: true
  mode: "simulation"
```

### Adding New Features

When adding new features that create files:

1. Use the `UserContext` system for proper file ownership
2. Call `EnsureFileOwnership()` after file creation
3. Add appropriate error handling for permission issues
4. Update the diagnostic function to check new file types

## Support

If you encounter issues:

1. Run the diagnostic function: `kvmService.DiagnosePermissionIssues(ctx)`
2. Check the logs for detailed error messages
3. Run the permission fix script: `./scripts/fix_libvirt_permissions.sh`
4. Verify libvirt installation and configuration
5. Check BIOS settings for KVM virtualization support
