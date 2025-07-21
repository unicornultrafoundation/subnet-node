# CreateAndStartVM API Documentation

## Overview

The `CreateAndStartVM` API is a new VirtualBox API method that allows you to quickly create and start a VM by cloning from a pre-existing template VM named `template_sample`. This approach eliminates the need to reinstall the operating system for each new VM, making VM creation much faster.

## Prerequisites

Before using the `CreateAndStartVM` API, you must have:

1. **VirtualBox installed** and properly configured
2. **A template VM named `template_sample`** that has been pre-configured with the desired operating system
3. **The VirtualBox service enabled** in your configuration

## Creating the Template VM

To create the template VM that will be used for cloning:

```bash
# Create a template VM with ubuntu/ubuntu credentials
# This should be done once to set up the template
```

The template VM should be:

- Named exactly `template_sample`
- Pre-configured with the desired operating system
- Have cloud-init properly configured
- Be in a stopped state when not in use

## API Usage

### Method Signature

```go
func (api *VirtualBoxAPI) CreateAndStartVM(ctx context.Context, name string, cpuCores int, memoryMB int, diskSizeGB int, osType string, username string, password string) (*vmResult, error)
```

### Parameters

- `name` (string): The name for the new VM
- `cpuCores` (int): Number of CPU cores to allocate
- `memoryMB` (int): Memory allocation in MB
- `diskSizeGB` (int): Disk size in GB
- `osType` (string): Operating system type (optional, inherited from template)
- `username` (string): Username for cloud-init configuration
- `password` (string): Password for cloud-init configuration

### Return Value

Returns a `vmResult` struct containing:

- `ID`: VM UUID
- `Name`: VM name
- `Status`: VM status (should be "running")
- `CPUCores`: Allocated CPU cores
- `MemoryMB`: Allocated memory
- `DiskSizeGB`: Disk size
- `IPAddress`: VM IP address (if available)
- `SSHPort`: SSH port (if configured)

## How It Works

1. **Template Validation**: Checks if the `template_sample` VM exists
2. **Resource Validation**: Validates system resources against the requested configuration
3. **VM Cloning**: Clones the template VM to create the new VM
4. **Hardware Configuration**: Updates CPU cores and memory to match the request
5. **Cloud-init Setup**: Generates new cloud-init ISO with the specified username/password
6. **VM Start**: Starts the VM in headless mode
7. **Status Update**: Updates the VM status to "running"

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "log"

    "github.com/unicornultrafoundation/subnet-node/internal/api"
)

func main() {
    // Create VirtualBox API instance
    vboxAPI := api.NewVirtualBoxAPI(vboxService)

    // Create and start a VM
    vm, err := vboxAPI.CreateAndStartVM(
        context.Background(),
        "my-new-vm",           // VM name
        2,                     // CPU cores
        4096,                  // Memory (4GB)
        50,                    // Disk size (50GB)
        "Ubuntu_ARM64",        // OS type
        "admin",               // Username
        "mypassword123",       // Password
    )
    if err != nil {
        log.Fatalf("Failed to create and start VM: %v", err)
    }

    log.Printf("VM created and started successfully!")
    log.Printf("VM ID: %s", vm.ID)
    log.Printf("VM Name: %s", vm.Name)
    log.Printf("Status: %s", vm.Status)
    log.Printf("CPU Cores: %d", vm.CPUCores)
    log.Printf("Memory: %d MB", vm.MemoryMB)
    log.Printf("Disk Size: %d GB", vm.DiskSizeGB)
}
```

### RPC/JSON-RPC Example

```json
{
  "jsonrpc": "2.0",
  "method": "virtualbox.CreateAndStartVM",
  "params": [
    "my-new-vm",
    2,
    4096,
    50,
    "Ubuntu_ARM64",
    "admin",
    "mypassword123"
  ],
  "id": 1
}
```

## Error Handling

The API includes comprehensive error handling:

- **Template Not Found**: Returns error if `template_sample` VM doesn't exist
- **Resource Validation**: Validates CPU, memory, and disk requirements
- **Cleanup on Failure**: Automatically cleans up failed VMs
- **Cloud-init Failures**: Continues operation even if cloud-init setup fails

## Common Error Scenarios

1. **Template VM Missing**:

   ```
   template VM 'template_sample' not found. Please create the template VM first
   ```

2. **Insufficient Resources**:

   ```
   insufficient CPU cores: requested 8, available 4
   ```

3. **VirtualBox Not Available**:
   ```
   VBoxManage not found or not accessible
   ```

## Performance Benefits

- **Fast VM Creation**: No OS installation required
- **Consistent Templates**: All VMs start from the same base configuration
- **Immediate Availability**: VMs are ready to use immediately after creation
- **Resource Efficiency**: Reduced storage and time requirements

## Security Considerations

- **Template Security**: Ensure the template VM is secure and up-to-date
- **Password Management**: Use strong passwords for cloud-init configuration
- **Network Isolation**: Consider network configuration for cloned VMs
- **Access Control**: Implement proper access controls for the API

## Troubleshooting

### VM Won't Start

- Check if VirtualBox is running
- Verify template VM exists and is properly configured
- Check system resources availability
- Review VirtualBox logs for detailed error messages

### Cloud-init Issues

- Verify cloud-init templates are valid
- Check username/password format
- Ensure cloud-init ISO generation succeeds

### Resource Allocation Issues

- Reduce CPU cores or memory requirements
- Check available system resources
- Consider stopping other VMs to free resources

## Best Practices

1. **Template Management**: Keep the template VM updated and secure
2. **Resource Planning**: Plan resource allocation based on system capacity
3. **Naming Convention**: Use descriptive names for VMs
4. **Monitoring**: Monitor VM performance and resource usage
5. **Backup Strategy**: Implement backup strategy for important VMs
