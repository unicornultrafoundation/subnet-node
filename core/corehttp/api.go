package corehttp

import (
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
	"github.com/unicornultrafoundation/subnet-node/rpc"
)

// APIPath is the path at which the API is mounted.
const APIPath = "/"

func APIOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, smux *http.ServeMux) (*http.ServeMux, error) {
		cfg := n.Repo.Config()

		server := rpc.NewServer()
		server.RegisterName("version", api.NewVersionAPI()) // Register the VersionAPI
		smux.Handle(APIPath, WithCORSHeaders(cfg, server))
		return smux, nil
	}
}
