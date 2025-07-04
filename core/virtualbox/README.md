# VirtualBox Service

The VirtualBox service provides comprehensive VM management capabilities for the subnet-node, acting as a hypervisor that can create, manage, and orchestrate VirtualBox VMs. It integrates with Terraform for infrastructure-as-code provisioning and automation.

## Features

- **VM Lifecycle Management**: Create, start, stop, shutdown, and delete VMs
- **Terraform Integration**: Automated VM provisioning using Terraform configurations
- **Resource Management**: Configure CPU, memory, disk, and network resources
- **Network Configuration**: Support for NAT, bridged, host-only, and internal networking
- **Storage Management**: Virtual disk creation and base image attachment
- **Monitoring**: Real-time VM status and resource usage monitoring
- **TTL-based Cleanup**: Automatic cleanup of expired VMs
- **REST API**: Full HTTP API for VM management operations
- **Persistent Storage**: VM metadata and state persistence

## Architecture

The VirtualBox service consists of several components:

```
VirtualBox Service
├── Service (core/virtualbox/service.go)
│   ├── VBox Client (core/virtualbox/vbox/client.go)
│   ├── Terraform Client (core/virtualbox/terraform/client.go)
│   ├── Storage Service (core/virtualbox/store/service.go)
│   └── Configuration (core/virtualbox/config/config.go)
├── Types (core/virtualbox/types/types.go)
├── API (internal/api/virtualbox.go)
└── Node Integration (core/node/virtualbox.go)
```

### Components

1. **Service**: Main orchestrator that coordinates VM operations
2. **VBox Client**: Direct VirtualBox operations using VBoxManage
3. **Terraform Client**: Terraform-based VM provisioning
4. **Storage Service**: Persistent storage for VM metadata
5. **Configuration**: Service configuration management
6. **API**: HTTP endpoints for external access
7. **Types**: Data structures and interfaces

## Configuration

### Enable VirtualBox Service

Add the following to your subnet-node configuration:

```yaml
virtualbox:
  enable: true
  vboxmanage_path: "VBoxManage"
  terraform_path: "terraform"

  # VM defaults
  default_memory_mb: 2048
  default_cpus: 2
  default_disk_size_gb: 20
  default_network_type: "bridged"

  # Paths
  default_vm_path: "~/VirtualBox VMs"
  terraform_work_dir: "~/.subnet-node/terraform"
  base_image_path: "~/.subnet-node/images"
  snapshot_path: "~/.subnet-node/snapshots"

  # Timeouts
  vm_start_timeout: "60s"
  vm_stop_timeout: "30s"
  vm_delete_timeout: "60s"
  terraform_timeout: "300s"

  # Monitoring
  monitor_interval: "30s"
  headless: true
```

### Configuration Options

| Option                 | Type     | Default                    | Description                   |
| ---------------------- | -------- | -------------------------- | ----------------------------- |
| `enable`               | bool     | false                      | Enable VirtualBox service     |
| `vboxmanage_path`      | string   | "VBoxManage"               | Path to VBoxManage executable |
| `terraform_path`       | string   | "terraform"                | Path to Terraform executable  |
| `default_memory_mb`    | int      | 2048                       | Default VM memory in MB       |
| `default_cpus`         | int      | 2                          | Default number of CPUs        |
| `default_disk_size_gb` | int      | 20                         | Default disk size in GB       |
| `default_network_type` | string   | "bridged"                  | Default network type          |
| `default_vm_path`      | string   | "~/VirtualBox VMs"         | Default VM storage path       |
| `terraform_work_dir`   | string   | "~/.subnet-node/terraform" | Terraform working directory   |
| `base_image_path`      | string   | "~/.subnet-node/images"    | Base image storage path       |
| `snapshot_path`        | string   | "~/.subnet-node/snapshots" | Snapshot storage path         |
| `vm_start_timeout`     | duration | "60s"                      | VM start timeout              |
| `vm_stop_timeout`      | duration | "30s"                      | VM stop timeout               |
| `vm_delete_timeout`    | duration | "60s"                      | VM delete timeout             |
| `terraform_timeout`    | duration | "300s"                     | Terraform operation timeout   |
| `monitor_interval`     | duration | "30s"                      | VM monitoring interval        |
| `headless`             | bool     | true                       | Run VMs in headless mode      |

## API Endpoints

### VM Management

#### Create VM

```http
POST /api/virtualbox/vms
Content-Type: application/json

{
  "name": "test-vm",
  "requester": "user123",
  "ttl": "1h",
  "config": {
    "memory_mb": 2048,
    "cpus": 2,
    "disk_size_gb": 20,
    "os_type": "Linux_64",
    "network_type": "bridged",
    "bridge_name": "en0",
    "enable_vrde": true,
    "vrde_port": 3389
  }
}
```

#### Create VM with Terraform

```http
POST /api/virtualbox/terraform/vms
Content-Type: application/json

{
  "name": "test-vm",
  "requester": "user123",
  "ttl": "1h",
  "config": {
    "memory_mb": 2048,
    "cpus": 2,
    "disk_size_gb": 20,
    "os_type": "Linux_64",
    "network_type": "bridged",
    "bridge_name": "en0",
    "base_image": "/path/to/image.ova"
  }
}
```

#### List VMs

```http
GET /api/virtualbox/vms?requester=user123
```

#### Get VM

```http
GET /api/virtualbox/vms/{vmID}
```

#### Start VM

```http
POST /api/virtualbox/vms/{vmID}/start
```

#### Stop VM

```http
POST /api/virtualbox/vms/{vmID}/stop
```

#### Shutdown VM

```http
POST /api/virtualbox/vms/{vmID}/shutdown
```

#### Delete VM

```http
DELETE /api/virtualbox/vms/{vmID}
```

## VM Configuration

### Basic Configuration

```go
type VMConfig struct {
    MemoryMB     int    `json:"memory_mb"`
    CPUs         int    `json:"cpus"`
    DiskSizeGB   int    `json:"disk_size_gb"`
    OSType       string `json:"os_type"`
    BaseImage    string `json:"base_image"`
    NetworkType  string `json:"network_type"`
    BridgeName   string `json:"bridge_name"`
    MACAddress   string `json:"mac_address"`
}
```

### Advanced Configuration

```go
type VMConfig struct {
    // ... basic config ...

    // Storage
    StorageController string `json:"storage_controller"`
    StorageType       string `json:"storage_type"`

    // Advanced settings
    EnableAudio        bool `json:"enable_audio"`
    EnableUSB          bool `json:"enable_usb"`
    EnableVRDE         bool `json:"enable_vrde"`
    VRDEPort           int  `json:"vrde_port"`
    EnablePAE          bool `json:"enable_pae"`
    EnableNestedPaging bool `json:"enable_nested_paging"`

    // Custom settings
    CustomSettings map[string]string `json:"custom_settings"`
}
```

## Network Configuration

### Network Types

1. **NAT**: Network Address Translation (default)

   - VM gets internet access through host
   - Host can't directly access VM
   - Good for development and testing

2. **Bridged**: Bridged networking

   - VM gets IP from same network as host
   - VM is directly accessible from network
   - Requires bridge adapter configuration

3. **Host-only**: Host-only networking

   - Private network between host and VMs
   - No internet access
   - Good for isolated testing

4. **Internal**: Internal networking
   - Private network between VMs only
   - No host or internet access
   - Good for VM-to-VM communication

### Bridge Configuration

For bridged networking, specify the correct bridge adapter:

```bash
# List available bridge adapters
VBoxManage list bridgedifs
```

Common bridge adapters:

- macOS: `en0`, `en1`, `en2`
- Linux: `eth0`, `wlan0`, `ens33`
- Windows: `Ethernet`, `Wi-Fi`

## Terraform Integration

The service can automatically generate and execute Terraform configurations for VM provisioning.

### Terraform Configuration

The service generates the following Terraform files:

1. **main.tf**: Main resource configuration
2. **variables.tf**: Variable definitions
3. **outputs.tf**: Output values
4. **terraform.tfvars**: Variable values

### Example Terraform Output

```hcl
# Generated main.tf
resource "virtualbox_vm" "vm" {
  name   = "subnet-vm-123"
  image  = "/path/to/image.ova"
  cpus   = 2
  memory = 2048

  network_adapter {
    type           = "bridged"
    host_interface = "en0"
  }

  disk {
    file = "subnet-vm-123.vdi"
    size = 20480
  }
}
```

## Usage Examples

### Go Code

```go
package main

import (
    "context"
    "time"
    "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func createVM(service *virtualbox.Service) error {
    request := &types.VMRequest{
        ID:        "vm-123",
        Name:      "test-vm",
        Requester: "user123",
        TTL:       1 * time.Hour,
        CreatedAt: time.Now(),
        Config: &types.VMConfig{
            MemoryMB:    2048,
            CPUs:        2,
            DiskSizeGB:  20,
            OSType:      "Linux_64",
            NetworkType: "bridged",
            BridgeName:  "en0",
        },
    }

    response, err := service.CreateVM(context.Background(), request)
    if err != nil {
        return err
    }

    fmt.Printf("Created VM: %s\n", response.Name)
    return nil
}
```

### HTTP API

```bash
# Create a VM
curl -X POST http://localhost:5001/api/virtualbox/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server",
    "requester": "dev-team",
    "ttl": "2h",
    "config": {
      "memory_mb": 4096,
      "cpus": 4,
      "disk_size_gb": 50,
      "os_type": "Linux_64",
      "network_type": "bridged",
      "bridge_name": "en0",
      "enable_vrde": true,
      "vrde_port": 3389
    }
  }'

# List VMs
curl http://localhost:5001/api/virtualbox/vms?requester=dev-team

# Start VM
curl -X POST http://localhost:5001/api/virtualbox/vms/vm-123/start

# Delete VM
curl -X DELETE http://localhost:5001/api/virtualbox/vms/vm-123
```

## Monitoring and Cleanup

### VM Monitoring

The service continuously monitors VMs and:

- Tracks VM state changes
- Monitors resource usage (CPU, memory, disk)
- Collects network statistics
- Detects expired VMs

### Automatic Cleanup

VMs are automatically cleaned up when:

- TTL expires
- VM is explicitly deleted
- Service detects VM failure

### Manual Cleanup

```bash
# List expired VMs
curl http://localhost:5001/api/virtualbox/vms | jq '.[] | select(.expires_at < now)'

# Force cleanup of specific VM
curl -X DELETE http://localhost:5001/api/virtualbox/vms/vm-123
```

## Troubleshooting

### Common Issues

1. **VirtualBox not found**

   ```bash
   # Check VBoxManage availability
   which VBoxManage
   VBoxManage --version
   ```

2. **Permission denied**

   ```bash
   # Check VirtualBox permissions
   sudo usermod -a -G vboxusers $USER
   ```

3. **Network interface not found**

   ```bash
   # List available bridge interfaces
   VBoxManage list bridgedifs
   ```

4. **Terraform provider not found**
   ```bash
   # Initialize Terraform
   terraform init
   ```

### Debugging

Enable debug logging:

```yaml
virtualbox:
  enable: true
  # ... other config ...

# Add to logging configuration
logging:
  level: debug
  virtualbox: debug
```

### Log Analysis

```bash
# Check VirtualBox service logs
grep "virtualbox" /var/log/subnet-node.log

# Check Terraform execution
grep "terraform" /var/log/subnet-node.log

# Check VM operations
grep "VM" /var/log/subnet-node.log
```

## Security Considerations

### Network Security

- Use appropriate network isolation for production VMs
- Consider using host-only networking for sensitive workloads
- Implement firewall rules for bridged networking
- Use VPN for remote VM access

### Access Control

- Implement authentication for API endpoints
- Use HTTPS for API communication
- Restrict VM creation to authorized users
- Monitor VM access and usage

### Resource Limits

- Set appropriate resource limits for VMs
- Monitor resource usage to prevent abuse
- Implement quotas for VM creation
- Use TTL to prevent long-running VMs

## Performance Optimization

### Resource Allocation

- Allocate appropriate CPU and memory based on workload
- Use SSD storage for better I/O performance
- Enable nested virtualization for better performance
- Use bridged networking for better network performance

### Monitoring

- Monitor VM resource usage
- Track VM creation and deletion rates
- Monitor Terraform execution times
- Alert on resource exhaustion

## Future Enhancements

- **VM Templates**: Pre-configured VM templates
- **Snapshot Management**: VM snapshot creation and restoration
- **Resource Pools**: Resource allocation pools
- **Load Balancing**: Automatic VM load balancing
- **Backup and Recovery**: Automated VM backup
- **Multi-hypervisor Support**: Support for other hypervisors
- **Container Integration**: Docker container support
- **Orchestration**: Kubernetes integration
