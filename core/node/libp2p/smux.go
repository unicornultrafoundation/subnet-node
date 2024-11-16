package libp2p

import (
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/config"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
)

// cfg.Swarm.Transports
func makeSmuxTransportOption(cfg *config.C) (libp2p.Option, error) {
	if cfg.GetInt("swarm.transports.multiplexers.yamux", 0) < 0 {
		return nil, fmt.Errorf("running libp2p with Swarm.Transports.Multiplexers.Yamux disabled is not supported")
	}

	return libp2p.Muxer(yamux.ID, yamux.DefaultTransport), nil
}

func SmuxTransport(cfg *config.C) func() (opts Libp2pOpts, err error) {
	return func() (opts Libp2pOpts, err error) {
		res, err := makeSmuxTransportOption(cfg)
		if err != nil {
			return opts, err
		}
		opts.Opts = append(opts.Opts, res)
		return opts, nil
	}
}
