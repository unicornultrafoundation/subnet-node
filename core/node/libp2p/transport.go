package libp2p

import (
	"fmt"

	"github.com/ipshipyard/p2p-forge/client"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/metrics"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	webrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	webtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/unicornultrafoundation/subnet-node/config"

	"go.uber.org/fx"
)

func Transports(cfg *config.C) interface{} {
	return func(params struct {
		fx.In
		Fprint   PNetFingerprint         `optional:"true"`
		ForgeMgr *client.P2PForgeCertMgr `optional:"true"`
	},
	) (opts Libp2pOpts, err error) {
		privateNetworkEnabled := params.Fprint != nil

		if cfg.GetBool("swarm.network.tcp", true) {
			// TODO(9290): Make WithMetrics configurable
			opts.Opts = append(opts.Opts, libp2p.Transport(tcp.NewTCPTransport, tcp.WithMetrics()))
		}

		if cfg.GetBool("swarm.network.websocket", true) {
			if params.ForgeMgr == nil {
				opts.Opts = append(opts.Opts, libp2p.Transport(websocket.New))
			} else {
				opts.Opts = append(opts.Opts, libp2p.Transport(websocket.New, websocket.WithTLSConfig(params.ForgeMgr.TLSConfig())))
			}
		}

		if cfg.GetBool("swarm.network.quic", !privateNetworkEnabled) {
			if privateNetworkEnabled {
				return opts, fmt.Errorf(
					"QUIC transport does not support private networks, please disable Swarm.Transports.Network.QUIC",
				)
			}
			opts.Opts = append(opts.Opts, libp2p.Transport(quic.NewTransport))
		}

		if cfg.GetBool("swarm.network.web_transport", !privateNetworkEnabled) {
			if privateNetworkEnabled {
				return opts, fmt.Errorf(
					"WebTransport transport does not support private networks, please disable Swarm.Transports.Network.WebTransport",
				)
			}
			opts.Opts = append(opts.Opts, libp2p.Transport(webtransport.New))
		}

		if cfg.GetBool("swarm.network.webrtc_direct", !privateNetworkEnabled) {
			if privateNetworkEnabled {
				return opts, fmt.Errorf(
					"WebRTC Direct transport does not support private networks, please disable Swarm.Transports.Network.WebRTCDirect",
				)
			}
			opts.Opts = append(opts.Opts, libp2p.Transport(webrtc.New))
		}

		return opts, nil
	}
}

func BandwidthCounter() (opts Libp2pOpts, reporter *metrics.BandwidthCounter) {
	reporter = metrics.NewBandwidthCounter()
	opts.Opts = append(opts.Opts, libp2p.BandwidthReporter(reporter))
	return opts, reporter
}
