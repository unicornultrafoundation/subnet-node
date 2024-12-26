package corehttp

import (
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/coreapi"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
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
		server.RegisterName("node", api.NewNodeAPI(capi.Resource()))
		server.RegisterName("pubsub", api.NewPubsubAPI(capi.PubSub()))
		server.RegisterName("app", api.NewAppAPI(n.Apps))
		server.RegisterName("uptime", api.NewUptimeAPI(n.Uptime))
		server.RegisterName("account", api.NewAccountAPI(n.Account))
		server.RegisterName("config", api.NewConfigAPI(n.Repo))

		if cfg.GetString("api.authorizations", "") != "" {
			authorizations := parseAuthorizationsFromConfig(cfg)
			smux.Handle(APIPath, WithAuth(authorizations, server))
		} else {
			smux.Handle(APIPath, server)
		}

		return smux, nil
	}
}
