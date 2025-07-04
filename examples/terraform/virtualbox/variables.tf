variable "vm_name" {
  description = "Name of the VM"
  type        = string
}

variable "os_type" {
  description = "Operating system type"
  type        = string
  default     = "Linux_64"
}

variable "memory_mb" {
  description = "Memory in MB"
  type        = number
  default     = 2048
}

variable "cpus" {
  description = "Number of CPUs"
  type        = number
  default     = 2
}

variable "disk_size_gb" {
  description = "Disk size in GB"
  type        = number
  default     = 20
}

variable "network_type" {
  description = "Network type (nat, bridged, hostonly, internal)"
  type        = string
  default     = "bridged"
}

variable "bridge_name" {
  description = "Bridge adapter name for bridged networking"
  type        = string
  default     = ""
}

variable "base_image_path" {
  description = "Path to base image (ISO or OVA file)"
  type        = string
}

variable "enable_audio" {
  description = "Enable audio support"
  type        = bool
  default     = false
}

variable "enable_usb" {
  description = "Enable USB support"
  type        = bool
  default     = false
}

variable "enable_vrde" {
  description = "Enable VRDE (VirtualBox Remote Desktop Extension)"
  type        = bool
  default     = false
}

variable "custom_settings" {
  description = "Custom VirtualBox settings"
  type        = map(string)
  default     = {}
} 