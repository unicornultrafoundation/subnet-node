# Libvirt Architecture Detection

This module provides comprehensive system architecture detection for KVM virtualization, specifically optimized for Linux environments (like colima).

## Features

- **Automatic Architecture Detection**: Detects x86_64, ARM64, PPC64LE, and S390x architectures
- **KVM Detection**: Automatically detects KVM availability and uses appropriate domain type
- **Linux CPU Information**: Reads detailed CPU information from `/proc/cpuinfo`
- **Libvirt Integration**: Uses libvirt capabilities to get host CPU model information
- **Fallback Support**: Graceful fallback to basic runtime information when detailed detection fails

## Architecture Detection Methods

### 1. Runtime Detection (Primary)

Uses Go's `runtime.GOARCH` to determine the basic architecture:

- `amd64` → `x86_64` with `pc-q35-2.12` machine type
- `arm64` → `aarch64` with `virt-8.2` machine type
- `ppc64le` → `ppc64le` with `pseries` machine type
- `s390x` → `s390x` with `s390-ccw-virtio` machine type

### 2. KVM Detection (Performance)

Automatically detects KVM availability via libvirt capabilities:

- **KVM Available**: Uses `<domain type='kvm'>` for hardware acceleration
- **KVM Not Available**: Falls back to `<domain type='qemu'>` for software emulation

### 3. Linux CPU Information (Enhanced)

On Linux systems, reads `/proc/cpuinfo` to get:

- CPU model name (e.g., "Intel(R) Core(TM) i7-9750H")
- CPU vendor (e.g., "Intel", "AMD")
- CPU features (e.g., "sse4_2", "avx2", "vmx")

### 4. Libvirt Capabilities (Advanced)

When libvirt is available, uses `virsh capabilities` to get:

- Host CPU model from libvirt XML capabilities
- Additional virtualization-specific information

## Usage

### Basic Usage

```go
// Create a domain manager
dm := libvirt.NewDomainManager(client, logger)

// Get system architecture information
arch := dm.GetSystemArchitecture()
fmt.Printf("Architecture: %s\n", arch.Architecture)
fmt.Printf("Machine: %s\n", arch.Machine)
fmt.Printf("CPU Model: %s\n", arch.CPUModel)
fmt.Printf("Vendor: %s\n", arch.Vendor)

// Check KVM availability
if client.IsKVMAvailable() {
    fmt.Println("KVM is available - using hardware acceleration")
} else {
    fmt.Println("KVM not available - using software emulation")
}

// Get appropriate domain type
domainType := client.GetDomainType() // "kvm" or "qemu"
```

### API Integration

```go
// Through the KVM service
archInfo, err := kvmService.GetSystemArchitecture(ctx)
if err != nil {
    log.Error("Failed to get architecture info:", err)
    return
}

// archInfo contains:
// - architecture: "x86_64" or "aarch64"
// - machine: "pc-q35-2.12" or "virt-8.2"
// - cpu_model: "Intel(R) Core(TM) i7-9750H"
// - vendor: "Intel" or "AMD"
// - features: ["sse4_2", "avx2", ...]
// - go_arch: "amd64" or "arm64"
// - go_os: "linux"
// - kvm_available: true/false
// - domain_type: "kvm" or "qemu"
// - libvirt_available: true/false
```

## VM Configuration

The detected architecture and KVM information is automatically used in VM domain XML generation:

### Domain Type Selection

- **KVM Available**: `<domain type='kvm'>` for hardware acceleration
- **KVM Not Available**: `<domain type='qemu'>` for software emulation

### Architecture-Specific Configuration

- **x86_64**: Uses `pc-q35-2.12` machine type with USB controllers, APIC, and ACPI power management
- **ARM64**: Uses `virt-8.2` machine type with virtio input devices (no APIC/ACPI)
- **Other architectures**: Uses appropriate machine types and simplified device configurations

### Architecture Compatibility Features

#### x86_64 Features

- **APIC**: Advanced Programmable Interrupt Controller
- **ACPI Power Management**: Suspend-to-memory and suspend-to-disk support
- **USB Controllers**: Full USB support with piix3-uhci and ehci controllers

#### ARM64 Features

- **No APIC**: Uses ARM GIC (Generic Interrupt Controller) instead
- **No ACPI Power Management**: Simplified power management without S3/S4
- **Virtio Input Devices**: Uses virtio-based keyboard and mouse

#### Cross-Architecture Features

- **Virtio Network**: High-performance network interface
- **Virtio Storage**: High-performance storage interface
- **VNC Graphics**: Remote display support
- **Virtio RNG**: Random number generation
- **Virtio Memory Balloon**: Dynamic memory management

## Performance Considerations

### KVM vs QEMU

- **KVM**: Hardware acceleration, much faster performance
- **QEMU**: Software emulation, slower but more compatible

### Detection Logic

1. Check libvirt capabilities for `domain type='kvm'`
2. If KVM is available, use `<domain type='kvm'>`
3. If KVM is not available, fallback to `<domain type='qemu'>`

## Testing

Run the tests to verify architecture and KVM detection:

```bash
go test ./core/kvm/libvirt -v
```

The tests will:

- Verify basic architecture detection
- Test KVM availability detection
- Validate domain type selection
- Test Linux CPU information enrichment (on Linux systems)
- Validate API integration

## Dependencies

- Linux: `/proc/cpuinfo` for CPU information
- Libvirt: `virsh` command for capabilities and KVM detection
- Go: `runtime` package for basic architecture detection

## Limitations

- Optimized for Linux environments (colima)
- Requires libvirt for KVM detection and advanced CPU information
- Fallback to basic detection when detailed information is unavailable
- QEMU fallback when KVM is not available (slower performance)
