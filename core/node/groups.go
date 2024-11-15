package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/libp2p"
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

func LibP2P(cfg *config.C, userResourceOverrides rcmgr.PartialLimitConfig) fx.Option {
	connmgr := fx.Provide(libp2p.ConnectionManager(cfg))
	autonat := fx.Provide(libp2p.AutoNATService(cfg))

	// Gather all the options
	opts := fx.Options(
		connmgr,
		autonat,
	)

	return opts
}

// Identity groups units providing cryptographic identity
func Identity(cfg *config.C) fx.Option {
	// PeerID

	cid := cfg.GetString("identity.peer_id", "")
	if cid == "" {
		return fx.Error(errors.New("identity was not set in config (was 'ipfs init' run?)"))
	}
	if len(cid) == 0 {
		return fx.Error(errors.New("no peer ID in config! (was 'ipfs init' run?)"))
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

// IPFS builds a group of fx Options based on the passed BuildCfg
func Subnet(ctx context.Context, bcfg *BuildCfg) fx.Option {
	bcfgOpts, cfg := bcfg.options(ctx)
	if cfg == nil {
		return bcfgOpts // error
	}

	return fx.Options()
}
