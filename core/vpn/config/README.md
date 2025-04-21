# VPN Configuration Package

This package provides configuration structures and validation for the VPN system.

## Configuration Structure

The `VPNConfig` struct contains all configuration settings for the VPN system, organized into logical groups. The configuration is defined in YAML format:

### Basic Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `enable` | Enable or disable the VPN service | `false` | `true` or `false` |
| `mtu` | Maximum Transmission Unit for the VPN interface | `1400` | 576-9000 |
| `virtual_ip` | Virtual IP address for this node in the VPN network | `""` | Valid IPv4 address |
| `subnet` | Subnet mask for the virtual IP (CIDR notation) | `8` | 0-32 |
| `routes` | Routes to be added to the routing table | `["10.0.0.0/8"]` | Array of valid CIDR blocks |

### Security Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `unallowed_ports` | Ports that are not allowed to be used for VPN traffic | `[]` | Array of port numbers |

### Worker Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `worker_idle_timeout` | Timeout in seconds after which an idle worker is terminated | `300` | > 0 |
| `worker_buffer_size` | Buffer size for worker packet queues | `100` | > 0 |
| `max_workers` | Maximum number of worker goroutines | `1000` | > 0 |
| `worker_cleanup_interval` | Interval in seconds for cleaning up idle workers | `60` | > 0 |

### Stream Pool Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `max_streams_per_peer` | Maximum number of streams per peer | `10` | > 0 |
| `min_streams_per_peer` | Minimum number of streams per peer | `3` | > 0 |
| `stream_idle_timeout` | Timeout in seconds after which an idle stream is closed | `300` | > 0 |
| `cleanup_interval` | Interval in seconds for cleaning up idle streams | `60` | > 0 |

### Circuit Breaker Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `circuit_breaker_failure_threshold` | Number of failures before the circuit breaker opens | `5` | > 0 |
| `circuit_breaker_reset_timeout` | Timeout in seconds before the circuit breaker resets to half-open state | `60` | > 0 |
| `circuit_breaker_success_threshold` | Number of successes needed to close the circuit breaker | `2` | > 0 |

### Stream Health Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `health_check_interval` | Interval in seconds between health checks | `30` | > 0 |
| `health_check_timeout` | Timeout in seconds for health check operations | `5` | > 0 |
| `max_consecutive_failures` | Maximum number of consecutive health check failures before marking a resource as unhealthy | `3` | > 0 |
| `warm_interval` | Interval in seconds for warming up streams | `60` | > 0 |

### Retry Settings

| Setting | Description | Default | Valid Range |
|---------|-------------|---------|------------|
| `retry_max_attempts` | Maximum number of retry attempts | `5` | > 0 |
| `retry_initial_interval` | Initial retry interval in seconds | `1` | > 0 |
| `retry_max_interval` | Maximum retry interval in seconds | `30` | > 0 |

## Configuration

The VPN configuration is defined in a single structure with sensible defaults for all settings.

## Usage

### Example YAML Configuration

```yaml
# VPN Configuration
---
vpn:
  # Basic settings
  enable: true

  # Network settings
  mtu: 1400                     # Maximum Transmission Unit
  virtual_ip: 10.0.0.1          # Virtual IP for this node
  subnet: 24                    # CIDR subnet mask
  routes:
    - 10.0.0.0/24               # Routes to be added

  # Security settings
  unallowed_ports:
    - "25"   # SMTP
    - "587"  # Submission
    - "465"  # SMTPS

  # Worker settings
  worker_idle_timeout: 300      # 5 minutes
  worker_buffer_size: 100
  max_workers: 1000

  # ... other settings
```

### Basic Usage

```go
// Create a new VPN configuration
vpnConfig := config.New(cfg)

// Validate the configuration
if err := vpnConfig.Validate(); err != nil {
    log.Errorf("Invalid VPN configuration: %v", err)
    return err
}
```



### Dynamic Configuration Updates

```go
// Update the configuration
if err := configService.UpdateConfig(newCfg); err != nil {
    log.Errorf("Failed to update configuration: %v", err)
    return err
}

// Update the configuration with a callback
err := configService.UpdateConfigWithCallback(newCfg, func(oldCfg, newCfg *config.VPNConfig) {
    // Handle configuration changes
    if oldCfg.MTU != newCfg.MTU {
        log.Infof("MTU changed from %d to %d", oldCfg.MTU, newCfg.MTU)
    }
})
```

## Example Configurations

Example configurations for different scenarios are provided in the `examples` directory:

- `basic.yaml`: Basic configuration for general use
- `high_performance.yaml`: Configuration optimized for high performance
- `low_resource.yaml`: Configuration optimized for low resource usage

## Validation

The configuration is validated to ensure that all settings are within valid ranges and that there are no conflicts between settings. The validation is performed by the `Validate` method of the `VPNConfig` struct.

```go
if err := vpnConfig.Validate(); err != nil {
    log.Errorf("Invalid VPN configuration: %v", err)
    return err
}
```
