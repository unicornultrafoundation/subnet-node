package node

import (
	"github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// RepoConfig loads configuration from the repo
func RepoConfig(repo repo.Repo) (*config.C, error) {
	return repo.Config(), nil
}

// Datastore provides the datastore
func Datastore(repo repo.Repo) datastore.Datastore {
	return repo.Datastore()
}
