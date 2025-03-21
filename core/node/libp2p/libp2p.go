package libp2p

import (
	"fmt"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/sirupsen/logrus"
	version "github.com/unicornultrafoundation/subnet-node"
	"github.com/unicornultrafoundation/subnet-node/config"
	"go.uber.org/fx"
)

var log = logrus.WithField("service", "p2pnode")

type Libp2pOpts struct {
	fx.Out

	Opts []libp2p.Option `group:"libp2p"`
}

func ConnectionManager(cfg *config.C) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		cm, err := connmgr.NewConnManager(
			cfg.GetInt("swarm.connmrg.low", 600),
			cfg.GetInt("swarm.connmrg.high", 900),
			connmgr.WithGracePeriod(cfg.GetDuration("swarm.connmrg.grace", 20*time.Second)))
		if err != nil {
			return opts, err
		}
		opts.Opts = append(opts.Opts, libp2p.ConnectionManager(cm))
		return
	}
}

func PstoreAddSelfKeys(id peer.ID, sk crypto.PrivKey, ps peerstore.Peerstore) error {
	if err := ps.AddPubKey(id, sk.GetPublic()); err != nil {
		return err
	}

	return ps.AddPrivKey(id, sk)
}

func simpleOpt(opt libp2p.Option) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		opts.Opts = append(opts.Opts, opt)
		return
	}
}

type priorityOption struct {
	priority, defaultPriority int64
	opt                       libp2p.Option
}

func prioritizeOptions(opts []priorityOption) libp2p.Option {
	type popt struct {
		priority int64 // lower priority values mean higher priority
		opt      libp2p.Option
	}
	enabledOptions := make([]popt, 0, len(opts))
	for _, o := range opts {
		if o.priority > 0 {
			enabledOptions = append(enabledOptions, popt{
				priority: o.priority,
				opt:      o.opt,
			})
		}
	}
	sort.Slice(enabledOptions, func(i, j int) bool {
		return enabledOptions[i].priority < enabledOptions[j].priority
	})
	p2pOpts := make([]libp2p.Option, len(enabledOptions))
	for i, opt := range enabledOptions {
		p2pOpts[i] = opt.opt
	}
	return libp2p.ChainOptions(p2pOpts...)
}

func UserAgent() func() (opts Libp2pOpts, err error) {
	return simpleOpt(libp2p.UserAgent(version.GetUserAgentVersion()))
}

func ForceReachability(cfg *config.C) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		v := cfg.GetString("internal.libp2p_force_reachability", "public")
		switch v {
		case "public":
			opts.Opts = append(opts.Opts, libp2p.ForceReachabilityPublic())
		case "private":
			opts.Opts = append(opts.Opts, libp2p.ForceReachabilityPrivate())
		default:
			return opts, fmt.Errorf("unrecognized reachability option: %s", v)
		}
		return
	}
}
