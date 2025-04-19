package vpn

import (
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/network"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// This file provides type aliases for backward compatibility

// VPNConfig is an alias for vpnconfig.VPNConfig
type VPNConfig = vpnconfig.VPNConfig

// VPNStream is an alias for types.VPNStream
type VPNStream = types.VPNStream

// PacketInfo is an alias for packet.PacketInfo
type PacketInfo = packet.PacketInfo

// QueuedPacket is an alias for packet.QueuedPacket
type QueuedPacket = packet.QueuedPacket

// PacketWorker is an alias for packet.Worker
type PacketWorker = packet.Worker

// PacketDispatcher is an alias for packet.Dispatcher
type PacketDispatcher = packet.Dispatcher

// CircuitState is an alias for resilience.CircuitState
type CircuitState = resilience.CircuitState

// CircuitBreaker is an alias for resilience.CircuitBreaker
type CircuitBreaker = resilience.CircuitBreaker

// CircuitBreakerManager is an alias for resilience.CircuitBreakerManager
type CircuitBreakerManager = resilience.CircuitBreakerManager

// MultiplexerMetrics is an alias for types.MultiplexerMetrics
type MultiplexerMetrics = types.MultiplexerMetrics

// VPNMetrics is an alias for metrics.VPNMetrics
type VPNMetrics = metrics.VPNMetrics

// StreamPoolMetrics is an alias for metrics.StreamPoolMetrics
type StreamPoolMetrics = metrics.StreamPoolMetrics

// HealthMetrics is an alias for metrics.HealthMetrics
type HealthMetrics = metrics.HealthMetrics

// TUNService is an alias for network.TUNService
type TUNService = network.TUNService

// ClientService is an alias for network.ClientService
type ClientService = network.ClientService

// ServerService is an alias for network.ServerService
type ServerService = network.ServerService

// VirtualIPManager is an alias for network.VirtualIPManager
type VirtualIPManager = network.VirtualIPManager

// PeerDiscovery is an alias for discovery.PeerDiscovery
type PeerDiscovery = discovery.PeerDiscovery

// Interface aliases for backward compatibility

// VPNService is an alias for api.VPNService
type VPNService = api.VPNService

// PeerDiscoveryService is an alias for api.PeerDiscoveryService
type PeerDiscoveryService = api.PeerDiscoveryService

// StreamService is an alias for api.StreamService
type StreamService = api.StreamService

// StreamPoolService is an alias for api.StreamPoolService
type StreamPoolService = api.StreamPoolService

// CircuitBreakerService is an alias for api.CircuitBreakerService
type CircuitBreakerService = api.CircuitBreakerService

// StreamHealthService is an alias for api.StreamHealthService
type StreamHealthService = api.StreamHealthService

// StreamMultiplexService is an alias for api.StreamMultiplexService
type StreamMultiplexService = api.StreamMultiplexService

// RetryService is an alias for api.RetryService
type RetryService = api.RetryService

// MetricsService is an alias for api.MetricsService
type MetricsService = api.MetricsService

// ConfigService is an alias for api.ConfigService
type ConfigService = api.ConfigService
