# Real KVM Service Deployment Guide

This guide explains how to deploy and use the **Real KVM Service** in Subnet Node, which provides actual virtualization capabilities using libvirt/QEMU/KVM.

## 🎯 What We've Built

### **✅ Real KVM Implementation**

1. **🔧 Libvirt Integration Components**:

   - `core/kvm/libvirt/client.go` - Libvirt connection management
   - `core/kvm/libvirt/domain.go` - VM lifecycle operations with XML generation
   - `core/kvm/libvirt/storage.go` - Storage pool and disk image management
   - `core/kvm/libvirt/network.go` - Virtual network creation and management

2. **⚙️ Auto-Detection Service**:

   - `core/kvm/service.go` - Enhanced service with libvirt auto-detection
   - Falls back to simulation mode if libvirt unavailable
   - Graceful degradation for development environments

3. **🌐 Complete API Integration**:
   - HTTP REST API for VM management
   - JSON-RPC API integration
   - Comprehensive error handling

## 🚀 Deployment Options

### **Option 1: Linux with Full KVM (Recommended)**

#### **Prerequisites**

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils

# CentOS/RHEL/Fedora
sudo dnf install qemu-kvm libvirt libvirt-python libguestfs-tools virt-install

# Add user to libvirt group
sudo usermod -aG libvirt $USER
sudo usermod -aG kvm $USER

# Start services
sudo systemctl enable libvirtd
sudo systemctl start libvirtd

# Verify
virsh version
virsh list --all
```

#### **Configuration**

```yaml
# config-production.yaml
kvm:
  enabled: true
  max_vms: 10
  max_cpu_cores: 8
  max_memory_mb: 16384
  max_disk_gb: 200

  # Real libvirt settings
  libvirt_uri: "qemu:///system"
  storage_pool: "subnet-vms"
  storage_path: "/var/lib/libvirt/images/subnet"
  network_name: "subnet-net"

provider:
  enable: true

api:
  enabled: true
  address: "0.0.0.0:8080"
```

#### **Deployment**

```bash
# Build subnet node
go build -o subnet ./cmd/subnet

# Initialize
./subnet init

# Start with real KVM
./subnet --config config-production.yaml
```

### **Option 2: macOS with Simulation Mode**

#### **Prerequisites**

```bash
# Install libvirt for compilation (no actual virtualization)
brew install pkg-config libvirt
```

#### **Configuration**

```yaml
# config-macos.yaml
kvm:
  enabled: true
  max_vms: 3
  max_cpu_cores: 2
  max_memory_mb: 4096
  max_disk_gb: 50

  # Libvirt will fail to connect, auto-fallback to simulation
  libvirt_uri: "qemu:///system"

provider:
  enable: true

api:
  enabled: true
  address: "127.0.0.1:8080"
```

#### **Deployment**

```bash
# Build subnet node (with libvirt bindings for compilation)
go build -o subnet ./cmd/subnet

# Initialize
./subnet init

# Start (will auto-detect libvirt unavailable and use simulation)
./subnet --config config-macos.yaml
```

### **Option 3: Docker Deployment**

#### **Dockerfile**

```dockerfile
FROM ubuntu:22.04

# Install KVM and libvirt
RUN apt-get update && apt-get install -y \
    qemu-kvm \
    libvirt-daemon-system \
    libvirt-clients \
    bridge-utils \
    && rm -rf /var/lib/apt/lists/*

# Add subnet user
RUN useradd -m -s /bin/bash subnet && \
    usermod -aG libvirt subnet && \
    usermod -aG kvm subnet

# Copy subnet binary
COPY subnet /usr/local/bin/subnet
COPY config-docker.yaml /etc/subnet/config.yaml

# Create directories
RUN mkdir -p /var/lib/libvirt/images/subnet && \
    chown -R subnet:subnet /var/lib/libvirt/images/subnet

USER subnet
WORKDIR /home/subnet

# Initialize and start
CMD ["/usr/local/bin/subnet", "--config", "/etc/subnet/config.yaml"]
```

## 🔧 Configuration Options

### **KVM Service Settings**

```yaml
kvm:
  enabled: true # Enable/disable KVM service
  max_vms: 10 # Maximum VMs allowed
  max_cpu_cores: 8 # Max CPU cores per VM
  max_memory_mb: 16384 # Max memory per VM (MB)
  max_disk_gb: 200 # Max disk per VM (GB)

  # Libvirt connection
  libvirt_uri: "qemu:///system" # Libvirt URI

  # Storage configuration
  storage_pool: "subnet-vms" # Storage pool name
  storage_path: "/var/lib/libvirt/images/subnet" # Storage directory

  # Network configuration
  network_name: "subnet-net" # Virtual network name
```

### **Auto-Detection Behavior**

- **Libvirt Available**: Automatically uses real virtualization
- **Libvirt Unavailable**: Falls back to simulation mode
- **No Code Changes**: Same API, different backend

## 🛠️ Usage Examples

### **1. Basic VM Operations**

#### **Create a VM**

```bash
curl -X POST http://localhost:8080/kvm/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server",
    "cpu_cores": 2,
    "memory_mb": 2048,
    "disk_gb": 20,
    "metadata": {
      "purpose": "web",
      "environment": "production"
    }
  }'
```

#### **Response (Real KVM)**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-server",
  "status": "stopped",
  "cpu_cores": 2,
  "memory_mb": 2048,
  "disk_gb": 20,
  "ip_address": "",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "metadata": {
    "purpose": "web",
    "environment": "production"
  }
}
```

#### **Start the VM**

```bash
curl -X POST http://localhost:8080/kvm/vms/550e8400-e29b-41d4-a716-446655440000/start
```

#### **Verify with virsh (Real KVM only)**

```bash
# List VMs
virsh list --all

# Get VM info
virsh dominfo web-server

# Connect to VM console
virsh console web-server
```

### **2. Monitor System Resources**

```bash
# Check system resources
curl http://localhost:8080/kvm/resources

# Example response
{
  "total_cpu_cores": 8,
  "total_memory_mb": 16384,
  "total_disk_gb": 500,
  "available_cpu_cores": 6,
  "available_memory_mb": 14336,
  "available_disk_gb": 480,
  "used_cpu_cores": 2,
  "used_memory_mb": 2048,
  "used_disk_gb": 20
}
```

### **3. VM Statistics**

```bash
# Get VM stats
curl http://localhost:8080/kvm/vms/550e8400-e29b-41d4-a716-446655440000/stats

# Example response
{
  "cpu_usage_percent": 25.5,
  "memory_usage_mb": 1024,
  "disk_usage_gb": 5.2,
  "network_rx_bytes": 1024000,
  "network_tx_bytes": 512000,
  "uptime_seconds": 3600
}
```

### **4. Service Status**

```bash
# Check service status
curl http://localhost:8080/kvm/status

# Example response (Real Mode)
{
  "enabled": true,
  "mode": "real",
  "libvirt_available": true,
  "libvirt_version": "11.4.0",
  "total_vms": 1,
  "running_vms": 1,
  "service": "kvm"
}

# Example response (Simulation Mode)
{
  "enabled": true,
  "mode": "simulation",
  "libvirt_available": false,
  "total_vms": 1,
  "running_vms": 0,
  "service": "kvm"
}
```

## 🏗️ Real vs Simulation Mode

### **Real Mode (Linux with KVM)**

- ✅ Actual QEMU/KVM virtual machines
- ✅ Real disk images (qcow2 format)
- ✅ Libvirt storage pools
- ✅ Virtual networks with DHCP
- ✅ Hardware-accelerated virtualization
- ✅ VM console access via virsh
- ✅ Integration with libvirt ecosystem

### **Simulation Mode (macOS/Windows/No KVM)**

- ✅ Complete API compatibility
- ✅ In-memory VM simulation
- ✅ Resource tracking and limits
- ✅ Persistent metadata storage
- ✅ Development and testing
- ❌ No actual virtualization
- ❌ No real disk images
- ❌ No network creation

## 🔍 Troubleshooting

### **Common Issues**

#### **1. Permission Denied**

```bash
# Error: Permission denied connecting to libvirt
sudo usermod -aG libvirt $USER
# Logout and login again
```

#### **2. Libvirt Not Running**

```bash
# Start libvirt daemon
sudo systemctl start libvirtd
sudo systemctl enable libvirtd
```

#### **3. Storage Directory Issues**

```bash
# Create and fix permissions
sudo mkdir -p /var/lib/libvirt/images/subnet
sudo chown libvirt-qemu:libvirt-qemu /var/lib/libvirt/images/subnet
sudo chmod 755 /var/lib/libvirt/images/subnet
```

#### **4. Network Bridge Issues**

```bash
# Check bridge configuration
ip link show
brctl show

# Restart network manager
sudo systemctl restart NetworkManager
```

### **Debug Commands**

```bash
# Check KVM support
lsmod | grep kvm
cat /proc/cpuinfo | grep vmx  # Intel
cat /proc/cpuinfo | grep svm  # AMD

# Test libvirt connection
virsh uri
virsh version
virsh capabilities

# Check logs
journalctl -u libvirtd -f
tail -f /var/log/libvirt/qemu/*.log
```

## 📊 Performance Considerations

### **Resource Allocation**

- **CPU**: Use host-model for best performance
- **Memory**: Direct allocation, avoid swap
- **Storage**: Use SSD for VM disks
- **Network**: virtio drivers for optimal throughput

### **Production Recommendations**

```yaml
kvm:
  max_vms: 20 # Adjust based on hardware
  max_cpu_cores: 16 # Leave cores for host
  max_memory_mb: 32768 # Leave memory for host
  max_disk_gb: 1000 # Monitor disk space

  storage_path: "/fast/ssd/path" # Use SSD storage
```

## 🔒 Security Considerations

### **Network Security**

- VMs use NAT by default (isolated from host network)
- Configure firewall rules for VM access
- Use bridge networks carefully in production

### **Storage Security**

- VM disk images stored in libvirt-owned directories
- Proper file permissions enforced
- Consider disk encryption for sensitive VMs

### **Access Control**

- Libvirt group membership required
- API access controls recommended
- Consider RBAC for multi-tenant environments

## 🎯 Success Criteria

After deployment, you should be able to:

1. ✅ **Create Real VMs**: VMs appear in `virsh list --all`
2. ✅ **Start/Stop VMs**: VMs actually boot and shutdown
3. ✅ **Access VMs**: Connect via console (`virsh console <vm>`)
4. ✅ **Monitor Resources**: Real resource consumption tracking
5. ✅ **Network Access**: VMs get DHCP IP addresses
6. ✅ **Persistent Storage**: VM disk images created and persisted

## 🔮 Next Steps

### **Enhancement Opportunities**

1. **Cloud-Init Integration**: Automated VM configuration
2. **Template Management**: Pre-built OS images
3. **Snapshot Support**: VM state snapshots
4. **Live Migration**: Move VMs between hosts
5. **Monitoring Integration**: Prometheus metrics
6. **Backup System**: Automated VM backups

This implementation successfully transforms the Subnet Node KVM service from a simulation into a **production-ready virtualization platform** with graceful fallback capabilities for development environments.
