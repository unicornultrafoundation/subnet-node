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
		server.RegisterName("version", api.NewVersionAPI())         // Register the VersionAPI
		server.RegisterName("keystore", api.NewKeystoreAPI(n.Repo)) // Register the KeystoreAPI

		// Load auth configuration from environment
		authConfig := rpc.LoadAuthConfigFromEnv()
		authMiddleware := rpc.NewAuthMiddleware(authConfig)

		// Wrap with auth middleware first, then CORS
		handler := WithCORSHeaders(cfg, authMiddleware.Wrap(server))
		smux.Handle(APIPath, handler)
		return smux, nil
	}
}
