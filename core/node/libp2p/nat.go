package libp2p

import (
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/unicornultrafoundation/subnet-node/config"
)

var NatPortMap = simpleOpt(libp2p.NATPortMap())

func AutoNATService(c *config.C) func() Libp2pOpts {
	return func() (opts Libp2pOpts) {
		opts.Opts = append(opts.Opts, libp2p.EnableNATService(), libp2p.EnableAutoNATv2())
		if c.IsSet("autonat.throttle") {
			opts.Opts = append(opts.Opts,
				libp2p.AutoNATServiceRateLimit(
					c.GetInt("autonat.throttle.global_limit", 0),
					c.GetInt("autonat.throttle.peer_limit", 0),
					c.GetDuration("autonat.throttle.peer_limit", time.Minute),
				),
			)
		}
		return opts
	}
}
