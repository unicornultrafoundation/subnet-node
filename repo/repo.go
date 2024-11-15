package repo

import (
	ds "github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/config"
)

type Repo interface {
	// Datastore returns a reference to the configured data storage backend.
	Datastore() Datastore
	Config() *config.C
	// SwarmKey returns the configured shared symmetric key for the private networks feature.
	SwarmKey() ([]byte, error)

	Close() error
}

// Datastore is the interface required from a datastore to be
// acceptable to FSRepo.
type Datastore interface {
	ds.Batching // must be thread-safe
}
