package corehttp

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/coreapi"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
	wsDeployerAPI "github.com/unicornultrafoundation/subnet-node/internal/api/ws/deployer"
	"github.com/unicornultrafoundation/subnet-node/rpc"
)

// APIPath is the path at which the API is mounted.
const APIPath = "/"

func APIOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, smux *http.ServeMux) (*http.ServeMux, error) {
		capi, err := coreapi.NewCoreAPI(n)

		if err != nil {
			return nil, err
		}

		cfg := n.Repo.Config()

		server := rpc.NewServer()
		server.RegisterName("swarm", api.NewSwarmAPI(capi.Swarm()))
		server.RegisterName("routing", api.NewRoutingAPI(capi.Routing()))
		server.RegisterName("node", api.NewNodeAPI(capi.Resource(), n.Apps, n))
		server.RegisterName("pubsub", api.NewPubsubAPI(capi.PubSub()))

		if cfg.GetBool("provider.enable", false) {
			server.RegisterName("app", api.NewAppAPI(n.Apps))
			server.RegisterName("deployments", api.NewDeployerAPI(n.Deployer))

			wsDeployer := wsDeployerAPI.NewDeployerWSAPI(n.Deployer)
			// Create a mux by Chi and register the handler
			wsMux := chi.NewRouter()
			wsMux.HandleFunc("/deployments/{orderID}/logs", wsDeployer.GetLogsHandler())
			wsMux.HandleFunc("/deployments/{orderID}/exec", wsDeployer.GetExecHandler())
			smux.Handle("/ws/", http.StripPrefix("/ws", wsMux))
		}

		server.RegisterName("account", api.NewAccountAPI(n.Account))
		server.RegisterName("config", api.NewConfigAPI(n.Repo))
		server.RegisterName("version", api.NewVersionAPI()) // Register the VersionAPI
		if cfg.GetBool("subgraph.enable", false) {
			server.RegisterName("peers", api.NewPeersAPI(n.Peers))
		}

		// Handle public APIs without authentication
		publicServer := rpc.NewServer()
		publicServer.RegisterName("version", api.NewVersionAPI())
		smux.Handle("/public", WithCORSHeaders(cfg, publicServer))

		smux.Handle(APIPath, WithCORSHeaders(cfg, WithAuth(cfg, server)))
		return smux, nil
	}
}
