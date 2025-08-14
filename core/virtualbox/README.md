# VirtualBox Package

This package provides a comprehensive interface for managing VirtualBox virtual machines programmatically. It allows for creating, managing, and controlling VMs with support for various operating systems and hardware configurations using cloud images.

## Core Components

### VirtualBox Service

The `VirtualboxService` is the main entry point for interacting with VirtualBox. It implements the `Service` interface and provides methods for:

- Creating and managing VMs from cloud images
- Starting, stopping, pausing, resuming, and resetting VMs
- Retrieving VM information and system resources
- Managing VM storage and configuration

### VBoxManage Executor

The `VBoxManageExecutor` provides a wrapper around the VirtualBox command-line interface (`VBoxManage`). It handles:

- Executing VBoxManage commands
- Parsing command output
- Managing VM configuration
- Setting up storage, networking, and hardware

### Storage Manager

The `StorageManager` handles image downloads and storage operations:

- Downloads Ubuntu cloud images from official sources
- Converts cloud images to VDI format for VirtualBox
- Generates cloud-init ISO files for VM initialization
- Manages organized directory structure for images

### Hardware Detector

The `HardwareDetector` automatically detects system hardware capabilities and configures VMs accordingly:

- Detects CPU architecture (x86_64, ARM64, etc.)
- Identifies available memory and storage
- Detects GPU capabilities
- Determines optimal VM settings based on hardware

## Features

### VM Management

- **Create VM from Image**: Create new VMs from cloud images with customizable hardware configurations
- **Start/Stop VM**: Control VM power state
- **Pause/Resume VM**: Suspend and resume VM execution
- **Reset VM**: Restart a VM
- **Delete VM**: Remove a VM and its associated files
- **Update VM**: Modify VM hardware configuration

### Cloud Image Integration

- **Automatic Image Download**: Downloads Ubuntu cloud images from official sources
- **Image Conversion**: Converts cloud images to VDI format for VirtualBox
- **Version Management**: Supports multiple Ubuntu versions (20.04, 22.04, etc.)
- **Architecture Support**: Supports both x86_64 and ARM64 architectures

### Cloud-Init Integration

- **Dynamic Configuration**: Generate cloud-init configurations for each VM
- **User Data Customization**: Set hostname, username, and password
- **ISO Generation**: Automatically create cloud-init ISO images for VM initialization
- **Secure Credentials**: Temporary credential files with automatic cleanup

### SSH Port Forwarding

- **Automatic Port Assignment**: Dynamically find available host ports
- **NAT Configuration**: Set up NAT rules for SSH access
- **Port Management**: Track and manage forwarded ports

### SSH Connection Management

The package provides a comprehensive SSH connection system that enables secure terminal access to VMs through WebSocket connections:

- **SSHServer**: Manages VM SSH configurations and port mappings
- **SSHConnection**: Handles active SSH sessions with bidirectional data flow
- **WebSocket Integration**: Provides real-time terminal access through web browsers
- **Terminal Emulation**: Full PTY support with xterm terminal emulation
- **Concurrent Connections**: Thread-safe handling of multiple SSH sessions
- **Automatic Cleanup**: Proper resource management and connection cleanup

#### SSH Connection Components

**SSHServer**

- Manages VM SSH configurations (host, port mappings)
- Thread-safe configuration storage and retrieval
- Supports adding, removing, and querying VM configurations

**SSHConnection**

- Establishes and maintains SSH connections to VMs
- Handles bidirectional data flow between SSH and WebSocket
- Provides terminal emulation with proper PTY configuration
- Manages connection lifecycle and cleanup

**VMPort**

- Represents VM port configuration with host and port information
- Used for SSH connection targeting and management

### Resource Management

- **Resource Validation**: Check system resources before VM creation
- **Overcommit Prevention**: Prevent resource exhaustion
- **Hardware Compatibility**: Ensure VM configurations match host capabilities

## Prerequisites

### Required Software

1. **VirtualBox**: Must be installed and accessible via `VBoxManage` command

   ```
   # macOS
   brew install virtualbox

   # Ubuntu/Debian
   sudo apt-get install virtualbox
   ```

2. **QEMU Tools**: Required for image conversion

   ```
   # macOS
   brew install qemu

   # Ubuntu/Debian
   sudo apt-get install qemu-utils
   ```

3. **ISO Generation Tools**: Required for cloud-init configuration

   ```
   # macOS
   brew install cdrtools

   # Ubuntu/Debian
   sudo apt-get install genisoimage
   ```

### Architecture Considerations

- **x86_64/AMD64**: Full support for all VirtualBox features
- **ARM64/AArch64**: Specialized configurations for Apple Silicon and other ARM platforms
- **Other Architectures**: Limited support with fallback configurations

## VM Creation Flow

The service uses cloud images to create new VMs efficiently. The flow for VM creation is:

1. **Image Preparation**: Downloads Ubuntu cloud images from official sources
2. **Image Conversion**: Converts cloud images to VDI format for VirtualBox
3. **VM Creation**: Creates a new VM without OS type specification
4. **Hardware Configuration**: Configures CPU, memory, and other hardware settings
5. **Storage Setup**: Attaches the converted VDI file using VirtioSCSI controller
6. **Cloud-Init Setup**: Generates and attaches cloud-init ISO for VM initialization
7. **Network Configuration**: Sets up SSH port forwarding for remote access

### Image Management

The storage manager organizes images in a structured directory:

```
~/VirtualBox VMs/
├── Images/
│   ├── 20.04/
│   │   ├── ubuntu-20.04-server-cloudimg-amd64.img
│   │   └── ubuntu-20.04-server-cloudimg-arm64.img
│   └── 22.04/
│       ├── ubuntu-22.04-server-cloudimg-amd64.img
│       └── ubuntu-22.04-server-cloudimg-arm64.img
└── vm-{orderId}/
    ├── vm-{orderId}.vdi
    └── cloud-init.iso
```

### Supported Ubuntu Versions

The system supports multiple Ubuntu versions:

- **Ubuntu 20.04 LTS**: Long-term support version
- **Ubuntu 22.04 LTS**: Latest long-term support version
- **Future versions**: Automatically supported as they become available

## API Types and VM Creation

The VirtualBox service provides a unified API for VM creation from cloud images:

### VM Creation from Image (`CreateVM`)

Used for creating VMs from cloud images with automatic resource configuration:

- **Request Type**: `VMCreateFromImageRequest` with hardware specifications and credentials
- **Resource Configuration**: Manually specified CPU, memory, and disk settings
- **OS Type**: Automatically determined based on host architecture
- **VM Naming**: Uses format `vm-{orderId}` when created from orders

### Request Structure

```go
type VMCreateFromImageRequest struct {
    Name       string `json:"name"`
    CPUCores   int    `json:"cpu_cores"`
    OS         string `json:"os"`
    Version    string `json:"version"`
    MemoryMB   int    `json:"memory_mb"`
    DiskSizeGB int    `json:"disk_size_gb"`
    Username   string `json:"username,omitempty"`
    Password   string `json:"password,omitempty"`
}
```

## VM Creation Strategy

The service follows an asynchronous job-based approach to VM creation for better scalability and user experience:

### Asynchronous Job Processing

1. **Request Submission**: VM creation requests are submitted through the `CreateVM` API
2. **Job Creation**: The request is immediately converted to a job and stored in the job manager
3. **Queue Processing**: The request is added to a buffered channel (queue) for background processing
4. **Background Execution**: A dedicated worker processes requests from the queue asynchronously
5. **Progress Tracking**: Job status and progress can be monitored through the job management APIs

### VM Creation Flow

When a VM creation request is processed from the queue:

1. **Resource Validation**: Check if sufficient system resources are available
2. **Hardware Detection**: Automatically detect host architecture
3. **Image Download**: Download or retrieve existing cloud image for the specified version
4. **Image Conversion**: Convert cloud image to VDI format for VirtualBox
5. **VM Creation**: Create a new VM without OS type specification
6. **Hardware Configuration**: Configure CPU, memory, and other hardware settings
7. **Storage Setup**: Attach the VDI file using VirtioSCSI controller
8. **Cloud-Init Configuration**: Generate custom cloud-init ISO with provided credentials
9. **Network Configuration**: Set up SSH port forwarding for remote access

### Job Management

- **Job Status**: Jobs can be in `Pending`, `Running`, `Completed`, `Failed`, or `Cancelled` states
- **Progress Monitoring**: Use `GetJobProgress` to check job status and results
- **Job Cancellation**: Long-running jobs can be cancelled using `CancelJob`
- **Automatic Cleanup**: Completed and failed jobs are automatically cleaned up after 24 hours

This approach provides efficient VM creation from cloud images with consistent configuration across deployments, while allowing for better resource management and user experience through asynchronous processing.

## Usage Examples

### Creating a VM from Cloud Image

The VM creation process uses cloud images to create VMs efficiently. The system automatically downloads and converts images as needed.

```
curl --location 'http://localhost:8081/api/v1/virtualbox/' \
--header 'Content-Type: application/json' \
--data '{
    "order_id": "1",
    "os": "ubuntu",
    "version": "22.04",
    "username": "ubuntu123",
    "password": "ubuntu"
}'
```

The system will:

1. Extract CPU cores, memory, and disk size from the order
2. Download Ubuntu 22.04 cloud image if not already available
3. Convert the cloud image to VDI format
4. Create a VM with the name format `vm-{orderId}`
5. Configure the VM with the order's specifications
6. Generate and attach cloud-init ISO for initialization

### Managing VMs

```
http://localhost:8081/api/v1/virtualbox/{vmId}/start
http://localhost:8081/api/v1/virtualbox/{vmId}/stop
http://localhost:8081/api/v1/virtualbox/{vmId}/pause
http://localhost:8081/api/v1/virtualbox/{vmId}/resume
http://localhost:8081/api/v1/virtualbox/{vmId}/reset
```

### Job Management

Get the progress of task in job queue

```
http://localhost:8081/api/v1/virtualbox/jobs
```

### SSH Connection Examples

#### Setting up SSH Server

```go
// Create SSH server instance
sshServer := virtualbox.NewSSHServer()

// Add VM configurations
sshServer.AddVMConfig("vm-123", "localhost", "2222")
sshServer.AddVMConfig("vm-456", "192.168.1.100", "2223")

// Get VM configuration
config, exists := sshServer.GetVMConfig("vm-123")
if exists {
    fmt.Printf("VM SSH: %s:%s\n", config.Host, config.Port)
}

// Get all configurations
allConfigs := sshServer.GetAllVMConfigs()
for vmID, config := range allConfigs {
    fmt.Printf("VM %s: %s:%s\n", vmID, config.Host, config.Port)
}
```

#### Creating SSH Connection

```go
// Create WebSocket connection (from your web server)
wsConn := // ... your WebSocket connection

// Create SSH connection
sshConn := virtualbox.NewSSHConnection("vm-123", wsConn, sshServer)

// Establish SSH connection with credentials
sshConn := virtualbox.NewSSHConnection(vmID, wsConn, sshServer)
    if err := sshConn.Connect("admin", "password"); err != nil {
        wsConn.WriteMessage(websocket.TextMessage, []byte("Connection failed: "+err.Error()))
        return
    }
// Start handling the connection
sshConn.Start()

// The connection will now handle bidirectional data flow:
// - SSH output → WebSocket (for display in browser)
// - WebSocket input → SSH (for user commands)
```

#### Terminal Configuration

The SSH connection automatically configures a full-featured terminal with:

- **Terminal Type**: xterm
- **Window Size**: 80x40 characters
- **Terminal Modes**:
  - Echo enabled
  - Echo control disabled
  - Input/Output speed: 14400 baud
- **Shell**: Interactive shell session

This configuration provides a complete terminal experience suitable for most command-line applications and development work.

#### Security Considerations

When using the SSH connection feature, consider the following security best practices:

- **Credential Management**: Store SSH credentials securely and avoid hardcoding them
- **Host Key Verification**: The current implementation uses `ssh.InsecureIgnoreHostKey()` for development. In production, implement proper host key verification
- **Connection Timeout**: Implement appropriate timeout mechanisms for idle connections
- **Access Control**: Implement proper authentication and authorization before allowing SSH connections
- **Network Security**: Ensure WebSocket connections are properly secured (WSS) in production environments
- **Resource Limits**: Monitor and limit the number of concurrent SSH connections to prevent resource exhaustion

#### Error Handling

The SSH connection system provides comprehensive error handling:

```go
// Handle connection errors
if err := sshConn.Connect("admin", "password"); err != nil {
    switch {
    case strings.Contains(err.Error(), "VM config not found"):
        // Handle missing VM configuration
    case strings.Contains(err.Error(), "SSH dial error"):
        // Handle network connectivity issues
    case strings.Contains(err.Error(), "Request PTY error"):
        // Handle terminal configuration issues
    default:
        // Handle other SSH-related errors
    }
}
```

#### Performance Considerations

- **Buffer Management**: The system uses 1KB buffers for data transfer, suitable for most terminal applications
- **Concurrent Connections**: Multiple SSH connections can run simultaneously with thread-safe operations
- **Memory Usage**: Each connection maintains minimal state and properly cleans up resources
- **Network Efficiency**: Binary WebSocket messages are used for optimal data transfer

## SSH WebSocket Connection Flow

To connect to a Virtual Machine via SSH using WebSocket, follow these steps:

### Step 1: Generate SSH Token

First, generate a one-time SSH access token using the JSON-RPC API:

```
http://localhost:8081/api/v1/virtualbox/{vmId}/ssh/token
```

**Parameters:**

- `username`: SSH username
- `password`: SSH password

### Step 2: Connect via WebSocket

Use the returned token to establish a WebSocket connection:

```
ws://localhost:8081/api/v1/vms/{vmId}/ssh?token={generated_token}
```

**Example:**

```
ws://localhost:8081/api/v1/vms/683c12f6-8af8-4cef-b8a9-960fa50e803d/ssh?token=abc123def456
```

The WebSocket connection provides a secure SSH terminal session to the VM.

## Architecture-Specific Considerations

The package automatically detects and adapts to different CPU architectures:

- **Apple Silicon (ARM64)**: Uses ARM-specific settings for chipset, firmware, and graphics
- **Intel/AMD (x86_64)**: Uses standard VirtualBox settings optimized for x86 architecture
- **Other ARM platforms**: Configures VMs with appropriate ARM virtualization settings

Each architecture receives optimized settings for:

- Chipset selection
- Firmware type (EFI vs BIOS)
- Graphics controller
- USB controller
- Audio configuration
