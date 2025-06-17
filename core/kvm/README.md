# Real KVM Service for Subnet Node

This module provides **real KVM/QEMU virtualization** capabilities using libvirt, enabling actual virtual machine management on Linux systems with KVM support.

## ✅ What's Real Now

Unlike the previous simulation, this implementation provides:

### **🖥️ Actual Virtualization**

- **Real VMs**: Creates actual QEMU/KVM virtual machines
- **Libvirt Integration**: Uses libvirt.org Go bindings for VM management
- **KVM Hypervisor**: Leverages Linux KVM for hardware-accelerated virtualization
- **QEMU Emulation**: Uses QEMU as the machine emulator

### **💾 Real Storage Management**

- **qcow2 Disk Images**: Creates actual qcow2 disk files
- **Storage Pools**: Manages libvirt storage pools
- **Template Support**: Clone VMs from template images
- **Volume Management**: Real disk volume creation and deletion

### **🌐 Real Networking**

- **Virtual Networks**: Creates libvirt virtual networks
- **NAT Networks**: Automatic NAT network creation with DHCP
- **Bridge Networks**: Support for bridged networking
- **MAC Address Generation**: Proper MAC address allocation
- **IP Management**: DHCP-based IP assignment

### **📊 Real Resource Management**

- **CPU Allocation**: Actual CPU core assignment to VMs
- **Memory Management**: Real memory allocation and limits
- **Resource Monitoring**: Live resource usage tracking
- **System Integration**: Integrates with host system resources

## 🏗️ Architecture

```
Subnet Node KVM Service
├── Service (Auto-detection)
│   ├── Libvirt Available → Real Mode
│   └── Libvirt Unavailable → Simulation Mode
├── Libvirt Components
│   ├── Client (libvirt connection)
│   ├── DomainManager (VM lifecycle)
│   ├── StorageManager (disk management)
│   └── NetworkManager (network setup)
└── API Layer
    ├── HTTP REST API
    └── JSON-RPC API
```

## 📁 Module Structure

```
core/kvm/
├── service.go              # Main service with auto-detection
├── service_real.go         # Real libvirt implementation
├── types.go               # Data structures
├── config/
│   ├── config.go          # Configuration management
│   └── validator.go       # Config validation
├── libvirt/
│   ├── client.go          # Libvirt client wrapper
│   ├── domain.go          # VM (domain) management
│   ├── storage.go         # Storage pool & volume management
│   └── network.go         # Virtual network management
└── README.md              # This documentation
```

## 🚀 Prerequisites

### **System Requirements**

1. **Linux System** with KVM support
2. **libvirt daemon** installed and running
3. **qemu-kvm** installed
4. **User permissions** to access libvirt

### **Installation (Ubuntu/Debian)**

```bash
# Install KVM and libvirt
sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils

# Add user to libvirt group
sudo usermod -aG libvirt $USER
sudo usermod -aG kvm $USER

# Start libvirt service
sudo systemctl enable libvirtd
sudo systemctl start libvirtd

# Verify installation
virsh version
```

### **Installation (CentOS/RHEL)**

```bash
# Install KVM and libvirt
sudo yum install qemu-kvm libvirt libvirt-python libguestfs-tools virt-install

# Add user to libvirt group
sudo usermod -aG libvirt $USER

# Start libvirt service
sudo systemctl enable libvirtd
sudo systemctl start libvirtd

# Verify installation
virsh version
```

## ⚙️ Configuration

### **Real KVM Configuration**

```yaml
# config-examples/kvm-real.yaml
kvm:
  enabled: true
  max_vms: 5
  max_cpu_cores: 4
  max_memory_mb: 8192
  max_disk_gb: 100

  # Libvirt settings
  libvirt_uri: "qemu:///system" # System libvirt connection
  storage_pool: "subnet-vms" # Storage pool name
  storage_path: "/var/lib/libvirt/images/subnet" # Storage directory
  network_name: "subnet-net" # Virtual network name
```

### **Auto-Detection Behavior**

- **Libvirt Available**: Uses real virtualization automatically
- **Libvirt Unavailable**: Falls back to simulation mode
- **Graceful Degradation**: No configuration changes needed

## 🛠️ Usage Examples

### **1. Start Subnet Node with Real KVM**

```bash
# Using real KVM configuration
./subnet --config config-examples/kvm-real.yaml
```

### **2. Create a Real VM**

```bash
# Create VM via HTTP API
curl -X POST http://localhost:8080/kvm/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "cpu_cores": 2,
    "memory_mb": 1024,
    "disk_gb": 10
  }'
```

### **3. Start the VM**

```bash
# Start VM
curl -X POST http://localhost:8080/kvm/vms/{vm-id}/start
```

### **4. Check VM Status**

```bash
# Get VM details
curl http://localhost:8080/kvm/vms/{vm-id}

# List all VMs
curl http://localhost:8080/kvm/vms

# Check with virsh
virsh list --all
virsh dominfo test-vm
```

## 🔧 VM Management Operations

### **VM Lifecycle**

- **Create**: `CreateVM()` - Creates qcow2 disk, generates XML, defines domain
- **Start**: `StartVM()` - Powers on the VM using libvirt
- **Stop**: `StopVM()` - Graceful shutdown (with force fallback)
- **Delete**: `DeleteVM()` - Removes VM and deletes disk image

### **Storage Operations**

- **Disk Creation**: Uses `qemu-img` or libvirt volume creation
- **Template Support**: Clone from base images
- **Storage Pools**: Automatic pool creation and management
- **Cleanup**: Proper disk deletion on VM removal

### **Network Operations**

- **Network Creation**: Automatic NAT network setup
- **DHCP Configuration**: Built-in DHCP for VM IP assignment
- **Bridge Support**: Integration with host bridges
- **MAC Generation**: Proper locally-administered MAC addresses

## 📊 Monitoring & Status

### **Real VM Monitoring**

```bash
# System resources
curl http://localhost:8080/kvm/resources

# VM statistics
curl http://localhost:8080/kvm/vms/{vm-id}/stats

# Service status
curl http://localhost:8080/kvm/status
```

### **System Integration**

- **Resource Limits**: Enforced CPU, memory, and disk limits
- **Resource Monitoring**: Integration with host resource service
- **State Synchronization**: VM state sync between libvirt and datastore

## 🔍 Debugging & Troubleshooting

### **Check Libvirt Connection**

```bash
# Test libvirt connection
virsh uri
virsh version

# Check permissions
groups $USER  # Should include 'libvirt'
```

### **Logs and Debugging**

```bash
# Check libvirt logs
sudo journalctl -u libvirtd -f

# Check KVM support
lsmod | grep kvm
```

### **Common Issues**

1. **Permission Denied**

   - Add user to `libvirt` group
   - Restart session after group change

2. **Network Issues**

   - Check bridge configuration
   - Verify firewall settings

3. **Storage Issues**
   - Check storage path permissions
   - Verify disk space availability

## 🔒 Security Considerations

### **Access Control**

- Service runs with user privileges
- libvirt group membership required
- Storage paths should be properly secured

### **Network Security**

- NAT networks provide isolation
- Bridge networks need careful configuration
- Consider firewall rules for VM access

## 🚀 Performance Considerations

### **Resource Management**

- CPU: Uses host-model CPU for best performance
- Memory: Direct memory allocation (no swap)
- Storage: qcow2 format for space efficiency
- Network: virtio drivers for performance

### **Optimization Tips**

- Use SSD storage for VM disks
- Allocate appropriate CPU cores
- Enable KVM hardware acceleration
- Use virtio drivers in VMs

## 🔮 Future Enhancements

### **Planned Features**

- **Cloud-init Support**: Automated VM configuration
- **Snapshot Management**: VM state snapshots
- **Live Migration**: Move VMs between hosts
- **Template Management**: Pre-built VM templates
- **Monitoring Integration**: Prometheus metrics
- **Backup System**: Automated VM backups

### **Advanced Networking**

- **Multiple Networks**: Multiple network interfaces per VM
- **VLAN Support**: Network segmentation
- **SDN Integration**: Software-defined networking
- **Load Balancing**: VM traffic distribution

This implementation transforms the Subnet Node KVM service from a simulation into a **real virtualization platform** capable of managing actual virtual machines using industry-standard technologies.
