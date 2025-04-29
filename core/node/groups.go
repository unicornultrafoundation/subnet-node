package node

import (
	"context"
	"errors"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/timecache"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/node/libp2p"
	"github.com/unicornultrafoundation/subnet-node/overlay/ifce"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

var log = logrus.WithField("core", "node")

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

	enableAutoTLS := cfg.GetBool("autotls.enabled", true)
	enableWebsocketTransport := cfg.GetBool("swarm.network.websocket", true)
	enableTCPTransport := cfg.GetBool("swarm.network.tcp", true)
	enableRelayTransport := cfg.GetBool("swarm.transports.network.relay", true)
	enableRelayService := cfg.GetBool("swarm.relay_service.enabled", enableRelayTransport)
	enableRelayClient := cfg.GetBool("swarm.relay_client.enabled", enableRelayService)
	//enableAutoWSS := cfg.GetBool("autotls.auto_wss", true)
	swarm := cfg.GetStringSlice("addresses.swarm", []string{})

	switch {
	case enableAutoTLS && enableTCPTransport && enableWebsocketTransport:
		// cfg.GetString("autotls.domain_suffix", p2pforge.DefaultForgeDomain)
		// // AutoTLS for Secure WebSockets: ensure WSS listeners are in place (manual or automatic)
		// wssWildcard := fmt.Sprintf("/tls/sni/*.%s/ws", cfg.GetString("autotls.domain_suffix", p2pforge.DefaultForgeDomain))
		// wssWildcardPresent := false
		// customWsPresent := false
		// customWsRegex := regexp.MustCompile(`/wss?$`)
		// tcpRegex := regexp.MustCompile(`/tcp/\d+$`)

		// // inspect listeners defined in config at Addresses.Swarm
		// var tcpListeners []string
		// for _, listener := range swarm {
		// 	// detect if user manually added /tls/sni/.../ws listener matching AutoTLS.DomainSuffix
		// 	if strings.Contains(listener, wssWildcard) {
		// 		log.Infof("found compatible wildcard listener in Addresses.Swarm. AutoTLS will be used on %s", listener)
		// 		wssWildcardPresent = true
		// 		break
		// 	}
		// 	// detect if user manually added own /ws or /wss listener that is
		// 	// not related to AutoTLS feature
		// 	if customWsRegex.MatchString(listener) {
		// 		log.Infof("found custom /ws listener set by user in Addresses.Swarm. AutoTLS will not be used on %s.", listener)
		// 		customWsPresent = true
		// 		break
		// 	}
		// 	// else, remember /tcp listeners that can be reused for /tls/sni/../ws
		// 	if tcpRegex.MatchString(listener) {
		// 		tcpListeners = append(tcpListeners, listener)
		// 	}
		// }

		// // Append AutoTLS's wildcard listener
		// // if no manual /ws listener was set by the user
		// if enableAutoWSS && !wssWildcardPresent && !customWsPresent {
		// 	if len(tcpListeners) == 0 {
		// 		log.Error("Invalid configuration, AutoTLS will be disabled: AutoTLS.AutoWSS=true requires at least one /tcp listener present in Addresses.Swarm, see https://github.com/ipfs/kubo/blob/master/docs/config.md#autotls")
		// 		enableAutoTLS = false
		// 	}
		// 	for _, tcpListener := range tcpListeners {
		// 		wssListener := tcpListener + wssWildcard
		// 		swarm = append(swarm, wssListener)
		// 		log.Infof("appended AutoWSS listener: %s", wssListener)
		// 	}
		// }

		// if !wssWildcardPresent && !enableAutoWSS {
		// 	log.Error(fmt.Sprintf("Invalid configuration, AutoTLS will be disabled: AutoTLS.Enabled=true requires a /tcp listener ending with %q to be present in Addresses.Swarm or AutoTLS.AutoWSS=true, see https://github.com/ipfs/kubo/blob/master/docs/config.md#autotls", wssWildcard))
		// 	enableAutoTLS = false
		// }
	case enableAutoTLS && !enableTCPTransport:
		log.Errorf("Invalid configuration: AutoTLS.Enabled=true requires swarm.network.tcp to be true as well. AutoTLS will be disabled.")
		enableAutoTLS = false
	case enableAutoTLS && !enableWebsocketTransport:
		log.Errorf("Invalid configuration: AutoTLS.Enabled=true requires swarm.network.websocket to be true as well. AutoTLS will be disabled.")
		enableAutoTLS = false
	}

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
		fx.Provide(libp2p.ListenOn(swarm)),
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
		fx.Invoke(VPNService),
		fx.Invoke(ProxyService),
		fx.Provide(PeerService),
		fx.Provide(AppService),
		fx.Provide(VerifierService),
		fx.Provide(DockerService),
		fx.Provide(account.EthereumService),
		fx.Provide(ifce.OverlayService),
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
