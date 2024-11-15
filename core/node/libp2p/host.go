package libp2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/helpers"
	"github.com/unicornultrafoundation/subnet-node/repo"
	"go.uber.org/fx"
)

type P2PHostIn struct {
	fx.In

	Repo          repo.Repo
	Validator     record.Validator
	HostOption    HostOption
	RoutingOption RoutingOption
	ID            peer.ID
	Peerstore     peerstore.Peerstore

	Opts [][]libp2p.Option `group:"libp2p"`
}

type P2PHostOut struct {
	fx.Out

	Host    host.Host
	Routing routing.Routing `name:"initialrouting"`
}

func Host(mctx helpers.MetricsCtx, lc fx.Lifecycle, params P2PHostIn) (out P2PHostOut, err error) {
	opts := []libp2p.Option{libp2p.NoListenAddrs}
	for _, o := range params.Opts {
		opts = append(opts, o...)
	}

	ctx := helpers.LifecycleCtx(mctx, lc)
	cfg := params.Repo.Config()

	bootstrappers, err := parseBootstrapPeers(cfg)
	if err != nil {
		return out, err
	}

	routingOptArgs := RoutingOptionArgs{
		Ctx:                           ctx,
		Datastore:                     params.Repo.Datastore(),
		Validator:                     params.Validator,
		BootstrapPeers:                bootstrappers,
		OptimisticProvide:             cfg.GetBool("experimental.optimistic_provide", false),
		OptimisticProvideJobsPoolSize: cfg.GetInt("experimental.optimistic_provide_jobs_pool_size", 0),
		LoopbackAddressesOnLanDHT:     cfg.GetBool("experimental.loopback_addresses_on_lan_dhl", false),
	}
	opts = append(opts, libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		args := routingOptArgs
		args.Host = h
		r, err := params.RoutingOption(args)
		out.Routing = r
		return r, err
	}))

	out.Host, err = params.HostOption(params.ID, params.Peerstore, opts...)
	if err != nil {
		return P2PHostOut{}, err
	}

	routingOptArgs.Host = out.Host

	// this code is necessary just for tests: mock network constructions
	// ignore the libp2p constructor options that actually construct the routing!
	if out.Routing == nil {
		r, err := params.RoutingOption(routingOptArgs)
		if err != nil {
			return P2PHostOut{}, err
		}
		out.Routing = r
		out.Host = routedhost.Wrap(out.Host, out.Routing)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return out.Host.Close()
		},
	})

	return out, err
}

// ParseBootstrapPeers parses a bootstrap list from the config into a list of AddrInfos.
func parseBootstrapPeers(cfg *config.C) ([]peer.AddrInfo, error) {
	// Lấy danh sách bootstrap từ config và chuyển thành []interface{}
	rawBootstrap := cfg.Get("bootstrap").([]interface{})

	// Chuyển đổi []interface{} thành []string
	addrs := make([]string, len(rawBootstrap))
	for i, addr := range rawBootstrap {
		strAddr, ok := addr.(string)
		if !ok {
			return nil, fmt.Errorf("invalid address format at index %d", i)
		}
		addrs[i] = strAddr
	}

	// Tạo danh sách Multiaddr từ các chuỗi địa chỉ
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
		}
	}

	// Chuyển đổi Multiaddr thành AddrInfo
	addrInfos, err := peer.AddrInfosFromP2pAddrs(maddrs...)
	if err != nil {
		return nil, fmt.Errorf("error converting to AddrInfo: %w", err)
	}

	return addrInfos, nil
}
