package libp2p

import (
	"context"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/unicornultrafoundation/subnet-node/config"
	"go.uber.org/fx"
)

// DockerACLFilter implements the relay.ACLFilter interface to only allow connections
// from Docker container IP ranges.
type DockerACLFilter struct {
	// Supported networks
	supportedNets []*net.IPNet
	// Whether to allow all connections (bypass ACL)
	allowAll bool
}

// NewDockerACLFilter creates a new ACL filter that only allows connections from Docker container IPs.
func NewDockerACLFilter(cfg *config.C) *DockerACLFilter {
	// Default Docker bridge network is 172.17.0.0/16
	_, defaultBridgeNet, _ := net.ParseCIDR("172.17.0.0/16")

	// Add common private network ranges
	_, privateNet1, _ := net.ParseCIDR("10.0.0.0/8")
	_, privateNet2, _ := net.ParseCIDR("172.16.0.0/12")
	_, privateNet3, _ := net.ParseCIDR("192.168.0.0/16")

	// Default networks to check
	defaultNets := []*net.IPNet{defaultBridgeNet, privateNet1, privateNet2, privateNet3}

	// Check if we should allow all connections
	allowAll := cfg.GetBool("swarm.relay_service.allow_all", false)

	return &DockerACLFilter{
		supportedNets: defaultNets,
		allowAll:      allowAll,
	}
}

// AllowReserve returns true if a reservation from a peer with the given peer ID and multiaddr
// is allowed. We check if the IP address in the multiaddr is within Docker container IP ranges.
func (f *DockerACLFilter) AllowReserve(p peer.ID, addr ma.Multiaddr) bool {
	if f.allowAll {
		return true
	}

	// Extract IP from multiaddr
	ip := extractIPFromMultiaddr(addr)
	if ip == nil {
		return false
	}

	// Check if IP is in Docker networks
	return f.isDockerIP(ip)
}

// AllowConnect returns true if a source peer with a given multiaddr is allowed to connect
// to a destination peer. We check if the source IP is within Docker container IP ranges.
func (f *DockerACLFilter) AllowConnect(src peer.ID, srcAddr ma.Multiaddr, dest peer.ID) bool {
	if f.allowAll {
		return true
	}

	// Extract IP from multiaddr
	ip := extractIPFromMultiaddr(srcAddr)

	if ip == nil {
		return false
	}

	// Check if IP is in Docker networks
	for _, net := range f.supportedNets {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

// isDockerIP checks if the given IP is within Docker container IP ranges.
func (f *DockerACLFilter) isDockerIP(ip net.IP) bool {
	// Check default Docker bridge network
	for _, net := range f.supportedNets {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

// extractIPFromMultiaddr extracts the IP address from a multiaddr.
func extractIPFromMultiaddr(addr ma.Multiaddr) net.IP {
	if addr == nil {
		return nil
	}

	// Convert multiaddr to string
	addrStr := addr.String()

	// Extract IP address from multiaddr string
	// Format examples: /ip4/192.168.1.1/tcp/1234, /ip6/::1/tcp/1234
	if strings.Contains(addrStr, "/ip4/") {
		parts := strings.Split(addrStr, "/")
		for i, part := range parts {
			if part == "ip4" && i+1 < len(parts) {
				return net.ParseIP(parts[i+1])
			}
		}
	} else if strings.Contains(addrStr, "/ip6/") {
		parts := strings.Split(addrStr, "/")
		for i, part := range parts {
			if part == "ip6" && i+1 < len(parts) {
				return net.ParseIP(parts[i+1])
			}
		}
	}

	return nil
}

func RelayTransport(enableRelay bool) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		if enableRelay {
			opts.Opts = append(opts.Opts, libp2p.EnableRelay())
		} else {
			opts.Opts = append(opts.Opts, libp2p.DisableRelay())
		}
		return
	}
}

func RelayService(enable bool, cfg *config.C) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		if enable {
			def := relay.DefaultResources()

			// Create a Docker ACL filter
			dockerACL := NewDockerACLFilter(cfg)

			// Real defaults live in go-libp2p.
			// Here we apply any overrides from user config.
			opts.Opts = append(opts.Opts, libp2p.EnableRelayService(
				relay.WithACL(dockerACL),
				relay.WithResources(relay.Resources{
					Limit: &relay.RelayLimit{
						Data:     int64(cfg.GetInt("swarm.relay_service.limit.data", int(def.Limit.Data))),
						Duration: cfg.GetDuration("swarm.relay_service.limit.duration", def.Limit.Duration),
					},
					MaxCircuits:           cfg.GetInt("swarm.relay_service.max_circuits", def.MaxCircuits),
					BufferSize:            cfg.GetInt("swarm.relay_service.buffer_size", def.BufferSize),
					ReservationTTL:        cfg.GetDuration("swarm.relay_service.reservation_ttl", def.ReservationTTL),
					MaxReservations:       cfg.GetInt("swarm.relay_service.max_reservations", def.MaxReservations),
					MaxReservationsPerIP:  cfg.GetInt("swarm.relay_service.max_reservations_per_ip", def.MaxReservationsPerIP),
					MaxReservationsPerASN: cfg.GetInt("swarm.relay_service.max_reservations_per_asn", def.MaxReservationsPerASN),
				}),
			))
		}
		return
	}
}

// libp2p.MaybeAutoRelay(cfg.Swarm.RelayClient.StaticRelays, cfg.Peering, enableRelayClient),
func MaybeAutoRelay(cfg *config.C, enabled bool) fx.Option {
	if !enabled {
		return fx.Options()
	}

	staticRelays := cfg.GetStringSlice("swarm.relay_client.static_relays", []string{})
	peers, _ := parsePeers(cfg, "swarm.peering.peers")
	if len(staticRelays) > 0 {
		return fx.Provide(func() (opts Libp2pOpts, err error) {
			if len(staticRelays) > 0 {
				static := make([]peer.AddrInfo, 0, len(staticRelays))
				for _, s := range staticRelays {
					var addr *peer.AddrInfo
					addr, err = peer.AddrInfoFromString(s)
					if err != nil {
						return
					}
					static = append(static, *addr)
				}
				opts.Opts = append(opts.Opts, libp2p.EnableAutoRelayWithStaticRelays(static))
			}
			return
		})
	}

	peerChan := make(chan peer.AddrInfo)
	return fx.Options(
		// Provide AutoRelay option
		fx.Provide(func() (opts Libp2pOpts, err error) {
			opts.Opts = append(opts.Opts,
				libp2p.EnableAutoRelayWithPeerSource(
					func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
						// TODO(9257): make this code smarter (have a state and actually try to grow the search outward) instead of a long running task just polling our K cluster.
						r := make(chan peer.AddrInfo)
						go func() {
							defer close(r)
							for ; numPeers != 0; numPeers-- {
								select {
								case v, ok := <-peerChan:
									if !ok {
										return
									}
									select {
									case r <- v:
									case <-ctx.Done():
										return
									}
								case <-ctx.Done():
									return
								}
							}
						}()
						return r
					},
					//autorelay.WithBootDelay(0),
					autorelay.WithMinInterval(0),
				))
			return
		}),
		autoRelayFeeder(peers, peerChan),
	)
}

func HolePunching(cfg *config.C, hasRelayClient bool) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		if cfg.GetBool("swarm.enable_hole_punching", true) {
			if !hasRelayClient {
				log.Fatal("Failed to enable `swarm.enable_hole_punching`, it requires `swarm.relay_client.enabled` to be true.")
				return
			}
			opts.Opts = append(opts.Opts, libp2p.EnableHolePunching())
		}
		return
	}
}
