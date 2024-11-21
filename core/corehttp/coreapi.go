package corehttp

import (
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/coreapi"
	"github.com/unicornultrafoundation/subnet-node/rpc"
)

// APIPath is the path at which the API is mounted.
const APIPath = "/api/v0"

func CoreAPIOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, smux *http.ServeMux) (*http.ServeMux, error) {
		api, err := coreapi.NewCoreAPI(n)
		if err != nil {
			return nil, err
		}

		server := rpc.NewServer()
		server.RegisterName("pubsub", api.PubSub())
		server.RegisterName("swarm", api.Swarm())
		server.RegisterName("routing", api.Routing())
		smux.Handle(APIPath, server)
		return smux, nil
	}
}
