# Subnet Node Kubernetes Setup Scripts

This directory contains cross-platform scripts to set up the subnet node infrastructure in a Kubernetes cluster.

## Available Scripts

### 1. `setup-kube.sh` - Bash Script (Linux/macOS/Windows WSL)
- **Platforms**: Linux, macOS, Windows (via WSL/Git Bash)
- **Shell**: Bash
- **Usage**: `./setup-kube.sh [wait] [command]`

### 2. `setup-kube.ps1` - PowerShell Script (Windows)
- **Platforms**: Windows (PowerShell 5.1+ / PowerShell Core 6+)
- **Shell**: PowerShell
- **Usage**: `.\setup-kube.ps1 [wait] [command]`

## Cross-Platform Support

| Platform | Script | Status | Notes |
|----------|--------|--------|-------|
| **Linux** | `setup-kube.sh` | ✅ Full Support | Native bash, all tools available |
| **macOS** | `setup-kube.sh` | ✅ Full Support | Native bash, Homebrew tools |
| **Windows WSL** | `setup-kube.sh` | ✅ Full Support | Linux environment in Windows |
| **Windows Git Bash** | `setup-kube.sh` | ✅ Full Support | Git Bash provides bash environment |
| **Windows PowerShell** | `setup-kube.ps1` | ✅ Full Support | Native PowerShell implementation |
| **Windows CMD** | `setup-kube.ps1` | ✅ Full Support | PowerShell available on modern Windows |

## Prerequisites

### Required Tools
All platforms require these tools:
- `kubectl` - Kubernetes command-line tool
- `jq` - JSON processor

**Note**: `grpcurl` is only required when gRPC checks are enabled (see Configuration section below).

### Platform-Specific Requirements

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get install netcat-openbsd jq

# CentOS/RHEL
sudo yum install nc jq

# Arch
sudo pacman -S netcat jq

# grpcurl (optional, for gRPC checks)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
export PATH=$PATH:~/go/bin
```

#### macOS
```bash
# Homebrew
brew install netcat jq

# grpcurl (optional, for gRPC checks)
brew install grpcurl
```

#### Windows
```bash
# Chocolatey
choco install kubernetes-cli jq

# Scoop
scoop install kubectl jq

# WSL (Ubuntu)
sudo apt-get install netcat-openbsd jq
```

## Configuration

### Environment Variables

Both scripts support the following environment variables for customization:

| Variable | Default | Description |
|----------|---------|-------------|
| `INVENTORY_REST_PORT` | `8080` | REST endpoint port for the inventory operator |
| `INVENTORY_GRPC_PORT` | `8081` | gRPC endpoint port for the inventory operator |
| `WITH_GRPC_CHECK` | `false` | Enable gRPC endpoint health checks |

### Port Configuration Examples

```bash
# Use custom ports
export INVENTORY_REST_PORT=18080
export INVENTORY_GRPC_PORT=18081
./setup-kube.sh

# PowerShell
$env:INVENTORY_REST_PORT="18080"
$env:INVENTORY_GRPC_PORT="18081"
.\setup-kube.ps1
```

### gRPC Check Control

```bash
# Enable gRPC checks (requires grpcurl)
export WITH_GRPC_CHECK=true
./setup-kube.sh

# PowerShell
$env:WITH_GRPC_CHECK="true"
.\setup-kube.ps1
```

## Usage

### Basic Setup
```bash
# Linux/macOS/WSL/Git Bash
./setup-kube.sh

# Windows PowerShell
.\setup-kube.ps1
```

### Wait Commands
```bash
# Wait for inventory operator to be available
./setup-kube.sh wait inventory-available
.\setup-kube.ps1 wait inventory-available

# Debug inventory operator
./setup-kube.sh wait debug
.\setup-kube.ps1 wait debug
```

## Platform Detection

The bash script automatically detects your platform and adjusts commands accordingly:

- **Linux**: Uses `nc -z` for port testing
- **macOS**: Uses `nc -z` (BSD version) for port testing  
- **Windows**: Detects available tools and uses appropriate fallbacks

## Script Features

### Automatic Port Mismatch Detection
- Detects when service ports don't match actual container listening ports
- Automatically fixes service configuration to use correct target ports
- Validates service configuration before proceeding

### Enhanced Service Readiness
- Waits for service endpoints to be ready before testing
- Better progress reporting during endpoint wait
- Comprehensive service validation

### Optional gRPC Health Checks
- gRPC checks are disabled by default for reliable setup
- Can be enabled via `WITH_GRPC_CHECK=true` environment variable
- Falls back to REST endpoint testing if gRPC is unavailable
- Continues setup even if gRPC services aren't ready

### Robust Error Handling
- Pod status monitoring before operations
- Temporary port-forward for testing
- Comprehensive debugging information
- Configurable timeouts and retries
- Graceful degradation when services aren't fully ready

### Cross-Platform Compatibility
- Platform-specific command detection
- Fallback methods for port testing
- Consistent behavior across environments

### Debugging Capabilities
- Detailed pod and service information
- Recent events and logs
- Network policy status
- Cluster connectivity verification
- Port configuration validation

## Troubleshooting

### Common Issues

#### 1. "Missing required tools" Error
- Install missing tools using platform-specific package managers
- Ensure tools are in your PATH
- Note: `grpcurl` is only required when `WITH_GRPC_CHECK=true`

#### 2. "kubectl not connected" Error
- Verify your kubeconfig is valid
- Check cluster connectivity: `kubectl cluster-info`

#### 3. "Pod not running" Error
- Check cluster resources and node availability
- Review pod events: `kubectl get events`

#### 4. "gRPC endpoint not responding" Error
- Use debug command: `./setup-kube.sh wait debug`
- Check network policies and service configuration
- Consider disabling gRPC checks: `WITH_GRPC_CHECK=false`

#### 5. Port Configuration Issues
- Script automatically detects and fixes port mismatches
- Check service configuration: `kubectl get svc operator-inventory -o yaml`
- Verify container ports: `kubectl get deployment operator-inventory -o yaml`

### Platform-Specific Issues

#### Windows PowerShell
- Ensure PowerShell execution policy allows script execution
- Run as Administrator if needed
- Use PowerShell 5.1+ or PowerShell Core 6+

#### macOS
- Ensure Homebrew tools are in PATH
- Check for conflicting netcat versions

#### Linux
- Verify netcat package installation
- Check firewall settings

## Advanced Usage

### Custom Port Configuration
```bash
# Full custom configuration
export INVENTORY_REST_PORT=18080
export INVENTORY_GRPC_PORT=18081
export WITH_GRPC_CHECK=true
./setup-kube.sh
```

### Conditional gRPC Testing
```bash
# Only enable gRPC checks in CI/dev environments
if [ "$CI" = "true" ] || [ "$ENVIRONMENT" = "dev" ]; then
    export WITH_GRPC_CHECK=true
fi
./setup-kube.sh
```

### Troubleshooting with Debug Mode
```bash
# Get detailed operator information
./setup-kube.sh wait debug

# Check specific service status
kubectl -n subnet-services get svc operator-inventory -o yaml
kubectl -n subnet-services get endpoints operator-inventory
```

## Contributing

When adding new features:
1. Test on all supported platforms
2. Use platform-agnostic commands when possible
3. Add platform-specific fallbacks for critical operations
4. Update this README with new platform support
5. Ensure both bash and PowerShell scripts maintain feature parity

## License

Same as the parent project.

