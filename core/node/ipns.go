package node

import (
	"github.com/ipfs/boxo/ipns"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/validator"
)

const DefaultIpnsCacheSize = 128

// RecordValidator provides namesys compatible routing record validator
func RecordValidator(ps peerstore.Peerstore) record.Validator {
	return record.NamespacedValidator{
		"resource": resource.ResourceValidator{},
		"vpn":      validator.VPNValidator{},
		"pk":       record.PublicKeyValidator{},
		"ipns":     ipns.Validator{KeyBook: ps},
	}
}
