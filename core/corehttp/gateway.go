package corehttp

import (
	"fmt"
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
)

func GatewayOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		cfg := n.Repo.Config()
		ethclient := n.Account.GetClient()
		bmAddress := common.HexToAddress(cfg.GetString("contracts.bid_market", config.DefaultBidMarketAddr))
		if bmAddress == (common.Address{}) {
			return nil, fmt.Errorf("bid market address not found in config")
		}

		bidMarket, err := contracts.NewBidMarketContract(ethclient, bmAddress, nil)
		if err != nil {
			return nil, err
		}

		// Add deployment handler if deployer is enabled
		if n.Deployer != nil {
			deploymentHandler := api.NewDeploymentHandler(n.Deployer, cfg, bidMarket)
			mux.Handle("/", deploymentHandler.Router())
		}

		// Add k8s handler if k8s deployer is enabled
		if n.K8sDeployer != nil {
			k8sHandler := api.NewK8sHandler(n.K8sDeployer, n.Account)
			mux.Handle("/k8s/", http.StripPrefix("/k8s", k8sHandler.Router()))
		}

		return mux, nil
	}
}
