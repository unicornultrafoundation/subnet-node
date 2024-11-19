package core

import (
	"context"
	"io"
	"time"

	"github.com/cskr/pubsub"
	"github.com/ipfs/boxo/bootstrap"
	"github.com/ipfs/boxo/peering"
	"github.com/jbenet/goprocess"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"github.com/unicornultrafoundation/subnet-node/repo"
	irouting "github.com/unicornultrafoundation/subnet-node/routing"
)

var log = logrus.New().WithField("service", "core")

// SubnetNode is Subnet Core module. It represents an Subnet instance.
type SubnetNode struct {
	ctx     context.Context
	Process goprocess.Process

	// Self
	Identity peer.ID // the local node's identity

	PrivateKey ic.PrivKey `optional:"true"` // the local node's private Key

	Repo repo.Repo

	// Services
	Peerstore       pstore.Peerstore `optional:"true"` // storage for other Peer instances
	RecordValidator record.Validator

	// Online
	PeerHost     p2phost.Host               `optional:"true"` // the network host (server+client)
	Peering      *peering.PeeringService    `optional:"true"`
	Routing      irouting.ProvideManyRouter `optional:"true"`
	Filters      *ma.Filters                `optional:"true"`
	Bootstrapper io.Closer                  `optional:"true"` // the periodic bootstrapper

	PubSub *pubsub.PubSub `optional:"true"`

	DHT       *ddht.DHT       `optional:"true"`
	DHTClient routing.Routing `name:"dhtc" optional:"true"`

	P2P *p2p.P2P `optional:"true"`

	stop func() error

	// Flags
	IsOnline bool `optional:"true"` // Online is set when networking is enabled.
	IsDaemon bool `optional:"true"` // Daemon is set when running on a long-running daemon.
}

// Close calls Close() on the App object
func (n *SubnetNode) Close() error {
	return n.stop()
}

// Context returns the IpfsNode context
func (n *SubnetNode) Context() context.Context {
	if n.ctx == nil {
		n.ctx = context.TODO()
	}
	return n.ctx
}

func (n *SubnetNode) Bootstrap(cfg bootstrap.BootstrapConfig) error {
	// TODO what should return value be when in offlineMode?
	if n.Routing == nil {
		return nil
	}

	if n.Bootstrapper != nil {
		n.Bootstrapper.Close() // stop previous bootstrap process.
	}

	var err error

	repoConf := n.Repo.Config()

	if repoConf.GetBool("backup.enable", true) {
		cfg.BackupBootstrapInterval = repoConf.GetDuration("backup.duration", time.Hour)
	}

	n.Bootstrapper, err = bootstrap.Bootstrap(n.Identity, n.PeerHost, n.Routing, cfg)
	return err
}
