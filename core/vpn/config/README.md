# VPN Configuration Package

This package provides configuration structures and validation for the VPN system.

## Components

### VPNConfig

The `VPNConfig` struct contains all configuration settings for the VPN system, including:

- Network settings (MTU, virtual IP, subnet, routes)
- Worker settings (idle timeout, buffer size, cleanup interval)
- Stream pool settings (max/min streams per peer, idle timeout)
- Circuit breaker settings (failure threshold, reset timeout, success threshold)
- Health check settings (interval, timeout, max consecutive failures)
- Multiplexer settings (max/min streams per multiplexer, auto-scaling interval)

### Validator

The validator provides functions to validate the VPN configuration, ensuring that:

- The virtual IP is set and valid
- The routes are set and valid
- The MTU is within a valid range (576-9000)

## Usage

```go
// Create a new VPN configuration
vpnConfig := config.NewVPNConfig(cfg)

// Validate the configuration
if err := vpnConfig.Validate(); err != nil {
    log.Errorf("Invalid VPN configuration: %v", err)
    return err
}
```
