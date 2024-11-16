package node

import (
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

const DefaultIpnsCacheSize = 128

// RecordValidator provides namesys compatible routing record validator
func RecordValidator(ps peerstore.Peerstore) record.Validator {
	return record.NamespacedValidator{
		"pk": record.PublicKeyValidator{},
	}
}