package corehttp

import (
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
)

func GatewayOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {

		// Add deployment handler if deployer is enabled
		cfg := n.Repo.Config()
		if cfg.GetBool("provider.enable", false) && n.Deployer != nil {
			deploymentHandler := api.NewDeploymentHandler(n.Deployer)
			mux.Handle("/api/v1/", deploymentHandler.Router())
		}

		return mux, nil
	}
}
