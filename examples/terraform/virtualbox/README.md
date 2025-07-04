# VirtualBox VM Provisioning with Terraform

This directory contains Terraform configurations for provisioning VirtualBox VMs as part of the subnet-node VirtualBox service.

## Prerequisites

1. **Terraform**: Install Terraform (version 1.0 or later)
2. **VirtualBox**: Install VirtualBox with VBoxManage command-line tools
3. **VirtualBox Terraform Provider**: The configuration uses the `terra-farm/virtualbox` provider

## Configuration

### 1. Copy the example configuration

```bash
cp terraform.tfvars.example terraform.tfvars
```

### 2. Edit terraform.tfvars

Modify the `terraform.tfvars` file with your desired VM configuration:

```hcl
vm_name = "my-vm"
memory_mb = 2048
cpus = 2
disk_size_gb = 20
network_type = "bridged"
bridge_name = "en0"  # Use appropriate network interface
base_image_path = "/path/to/your/image.ova"
```

### 3. Network Configuration

Choose the appropriate network type:

- **NAT**: Default, provides internet access through host
- **Bridged**: VM gets IP from same network as host
- **Host-only**: Private network between host and VMs
- **Internal**: Private network between VMs only

For bridged networking, specify the correct bridge adapter:

- macOS: `en0`, `en1`, etc.
- Linux: `eth0`, `wlan0`, etc.
- Windows: `Ethernet`, `Wi-Fi`, etc.

## Usage

### Initialize Terraform

```bash
terraform init
```

### Plan the deployment

```bash
terraform plan
```

### Apply the configuration

```bash
terraform apply
```

### Destroy the VM

```bash
terraform destroy
```

## Integration with subnet-node

This Terraform configuration is automatically used by the subnet-node VirtualBox service when creating VMs with Terraform. The service will:

1. Generate the Terraform files dynamically based on VM requests
2. Execute `terraform init` and `terraform apply`
3. Parse the output to get VM information
4. Clean up resources when VMs are deleted

## API Usage

### Create VM with Terraform

```bash
curl -X POST http://localhost:5001/api/virtualbox/terraform/vms \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

### Create VM with VBoxManage

```bash
curl -X POST http://localhost:5001/api/virtualbox/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "requester": "user123",
    "ttl": "1h",
    "config": {
      "memory_mb": 2048,
      "cpus": 2,
      "disk_size_gb": 20,
      "os_type": "Linux_64"
    }
  }'
```

## Troubleshooting

### Common Issues

1. **VirtualBox not found**: Ensure VBoxManage is in your PATH
2. **Network interface not found**: Check available interfaces with `VBoxManage list bridgedifs`
3. **Permission denied**: Run with appropriate permissions for VirtualBox
4. **Terraform provider not found**: Run `terraform init` to download the provider

### Debugging

Enable debug logging in the subnet-node configuration:

```yaml
virtualbox:
  enable: true
  vboxmanage_path: "VBoxManage"
  terraform_path: "terraform"
  monitor_interval: "30s"
```

Check the logs for detailed error messages and Terraform output.

## Security Considerations

- Use appropriate network isolation for production VMs
- Consider using host-only networking for sensitive workloads
- Implement proper access controls for the API endpoints
- Regularly update VirtualBox and Terraform to latest versions
