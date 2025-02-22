package node

import (
	"context"
	"errors"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/timecache"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/node/libp2p"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

var BaseLibP2P = fx.Options(
	fx.Provide(libp2p.PNet),
	fx.Provide(libp2p.ConnectionManager),
	fx.Provide(libp2p.Host),
	fx.Provide(libp2p.MultiaddrResolver),

	fx.Provide(libp2p.DiscoveryHandler),

	fx.Invoke(libp2p.PNetChecker),
)

func LibP2P(bcfg *BuildCfg, cfg *config.C) fx.Option {
	connmgr := fx.Provide(libp2p.ConnectionManager(cfg))
	autonat := fx.Provide(libp2p.AutoNATService(cfg))

	ps, disc := fx.Options(), fx.Options()
	if bcfg.getOpt("pubsub") {
		disc = fx.Provide(libp2p.TopicDiscovery())

		var pubsubOptions []pubsub.Option
		pubsubOptions = append(
			pubsubOptions,
			pubsub.WithMessageSigning(!cfg.GetBool("pubsub.disable_signing", false)),
			pubsub.WithSeenMessagesTTL(cfg.GetDuration("pubsub.seen_messages_ttl", pubsub.TimeCacheDuration)),
		)

		var seenMessagesStrategy timecache.Strategy
		configSeenMessagesStrategy := cfg.GetString("pubsub.seen_messages_strategy", "last-seen")
		//configSeenMessagesStrategy := cfg.Pubsub.SeenMessagesStrategy.WithDefault(config.DefaultSeenMessagesStrategy)
		switch configSeenMessagesStrategy {
		case "last-seen":
			seenMessagesStrategy = timecache.Strategy_LastSeen
		case "first-seen":
			seenMessagesStrategy = timecache.Strategy_FirstSeen
		default:
			return fx.Error(fmt.Errorf("unsupported Pubsub.SeenMessagesStrategy %q", configSeenMessagesStrategy))
		}
		pubsubOptions = append(pubsubOptions, pubsub.WithSeenMessagesStrategy(seenMessagesStrategy))

		switch cfg.GetString("pubsub.router", "gossipsub") {
		case "":
			fallthrough
		case "gossipsub":
			ps = fx.Provide(libp2p.GossipSub(pubsubOptions...))
		case "floodsub":
			ps = fx.Provide(libp2p.FloodSub(pubsubOptions...))
		default:
			return fx.Error(fmt.Errorf("unknown pubsub router %s", cfg.GetString("pubsub.router", "gossipsub")))
		}
	}

	enableAutoTLS := cfg.GetBool("autotls.enabled", false)
	enableRelayTransport := cfg.GetBool("swarm.transports.network.relay", true)
	enableRelayService := cfg.GetBool("swarm.relay_service.enabled", enableRelayTransport)
	enableRelayClient := cfg.GetBool("swarm.relay_client.enabled", enableRelayService)

	// Gather all the options
	opts := fx.Options(
		BaseLibP2P,
		fx.Provide(libp2p.UserAgent()),
		maybeProvide(libp2p.P2PForgeCertMgr(bcfg.Repo.Path(), cfg), enableAutoTLS),
		maybeInvoke(libp2p.StartP2PAutoTLS, enableAutoTLS),
		fx.Provide(libp2p.AddrFilters(cfg)),
		fx.Provide(libp2p.AddrsFactory(cfg)),
		fx.Provide(libp2p.SmuxTransport(cfg)),
		fx.Provide(libp2p.RelayTransport(enableRelayTransport)),
		fx.Provide(libp2p.RelayService(enableRelayService, cfg)),
		fx.Provide(libp2p.Transports(cfg)),
		fx.Provide(libp2p.ListenOn(cfg)),
		fx.Invoke(libp2p.SetupDiscovery(cfg.GetBool("discovery.mdns.enabled", true))),
		fx.Provide(libp2p.ForceReachability(cfg)),
		fx.Provide(libp2p.HolePunching(cfg, enableRelayClient)),

		fx.Provide(libp2p.Security(!bcfg.DisableEncryptedConnections, cfg)),

		fx.Provide(libp2p.Routing),
		fx.Provide(libp2p.ContentRouting),

		fx.Provide(libp2p.BaseRouting(cfg)),
		maybeProvide(libp2p.PubsubRouter, bcfg.getOpt("ipsnps")),

		maybeProvide(libp2p.BandwidthCounter, !cfg.GetBool("swarm.disable_bandwidth_metrics", false)),
		maybeProvide(libp2p.NatPortMap, !cfg.GetBool("swarm.disable_nat_portmap", false)),
		libp2p.MaybeAutoRelay(cfg, enableRelayClient),
		connmgr,
		autonat,
		ps,
		disc,
	)

	return opts
}

// Identity groups units providing cryptographic identity
func Identity(cfg *config.C) fx.Option {
	// PeerID
	cid := cfg.GetString("identity.peer_id", "")
	if cid == "" {
		return fx.Error(errors.New("identity was not set in config (was 'subnet init' run?)"))
	}
	if len(cid) == 0 {
		return fx.Error(errors.New("no peer ID in config! (was 'subnet init' run?)"))
	}

	id, err := peer.Decode(cid)
	if err != nil {
		return fx.Error(fmt.Errorf("peer ID invalid: %s", err))
	}

	// Private Key
	sk, err := cfg.GetPrivKey("identity.privkey", "")
	if err != nil {
		return fx.Error(err)
	}

	if sk == nil {
		return fx.Options( // No PK (usually in tests)
			fx.Provide(PeerID(id)),
			fx.Provide(libp2p.Peerstore),
		)
	}

	return fx.Options( // Full identity
		fx.Provide(PeerID(id)),
		fx.Provide(PrivateKey(sk)),
		fx.Provide(libp2p.Peerstore),

		fx.Invoke(libp2p.PstoreAddSelfKeys),
	)
}

// IPNS groups namesys related units
var IPNS = fx.Options(
	fx.Provide(RecordValidator),
)

// Online groups online-only units
func Online(bcfg *BuildCfg, cfg *config.C) fx.Option {
	return fx.Options(
		LibP2P(bcfg, cfg),
		fx.Provide(DNSResolver),
		fx.Provide(Peering),
		fx.Provide(p2p.New),
	)
}

// Storage groups units which setup datastore based persistence and blockstore layers
func Storage(bcfg *BuildCfg) fx.Option {
	return fx.Options(
		fx.Provide(RepoConfig),
		fx.Provide(Datastore),
	)
}

func Core(cfg *config.C) fx.Option {
	return fx.Options(
		fx.Provide(ResourceService),
		fx.Provide(AppService),
		fx.Provide(account.EthereumService),
	)
}

// IPFS builds a group of fx Options based on the passed BuildCfg
func Subnet(ctx context.Context, bcfg *BuildCfg) fx.Option {
	bcfgOpts, cfg := bcfg.options(ctx)
	if cfg == nil {
		return bcfgOpts // error
	}

	return fx.Options(
		bcfgOpts,
		fx.Provide(baseProcess),
		Identity(cfg),
		Storage(bcfg),
		IPNS,
		Online(bcfg, cfg),
		Core(cfg),
	)
}
