package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/unicornultrafoundation/subnet-node/config"
	"go.uber.org/fx"
)

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

			// Real defaults live in go-libp2p.
			// Here we apply any overrides from user config.
			opts.Opts = append(opts.Opts, libp2p.EnableRelayService(relay.WithResources(relay.Resources{
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
			})))
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
