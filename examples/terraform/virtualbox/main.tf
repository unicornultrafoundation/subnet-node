terraform {
  required_providers {
    virtualbox = {
      source = "terra-farm/virtualbox"
      version = "~> 0.2.2-alpha.1"
    }
  }
}

# VirtualBox VM resource
resource "virtualbox_vm" "vm" {
  name   = var.vm_name
  image  = var.base_image_path
  cpus   = var.cpus
  memory = var.memory_mb

  # Network configuration
  network_adapter {
    type           = var.network_type
    host_interface = var.bridge_name
  }

  # Storage configuration
  disk {
    file = "${var.vm_name}.vdi"
    size = var.disk_size_gb * 1024  # Convert GB to MB
  }

  # Advanced settings
  audio = var.enable_audio ? "pulse" : "none"
  usb   = var.enable_usb ? "on" : "off"
  vrde  = var.enable_vrde ? "on" : "off"
  
  # Custom settings
  dynamic "custom_settings" {
    for_each = var.custom_settings
    content {
      name  = custom_settings.key
      value = custom_settings.value
    }
  }
}

# Output values
output "vm_uuid" {
  description = "VM UUID"
  value       = virtualbox_vm.vm.id
}

output "vm_name" {
  description = "VM name"
  value       = virtualbox_vm.vm.name
}

output "state" {
  description = "VM state"
  value       = virtualbox_vm.vm.state
}

output "ip_address" {
  description = "VM IP address"
  value       = virtualbox_vm.vm.network_adapter[0].ipv4_address
}

output "ssh_port" {
  description = "SSH port"
  value       = 22
}

output "vrde_port" {
  description = "VRDE port"
  value       = var.enable_vrde ? 3389 : null
}

output "disk_path" {
  description = "Disk file path"
  value       = virtualbox_vm.vm.disk[0].file
}

output "network_config" {
  description = "Network configuration"
  value       = jsonencode(virtualbox_vm.vm.network_adapter)
} 