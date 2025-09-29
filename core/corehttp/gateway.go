package corehttp

import (
	"fmt"
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rancher/wrangler/v3/pkg/signals"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
	kubeconfig "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/config"
	kubeServer "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/server"
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

		// Add KubeVirt handler if KubeVirt is enabled
		if cfg.GetBool("kubevirt.enable", false) {
			ctx := signals.SetupSignalContext()

			kubeConfig, err := kubeServer.GetConfig(cfg.GetString("kubevirt.kubeconfig", ""))
			if err != nil {
				return nil, err
			}

			options := kubeconfig.Options{
				Namespace: cfg.GetString("kubevirt.namespace", "default"),
			}

			kubeServer, err := kubeServer.New(ctx, kubeConfig, options)
			if err != nil {
				return nil, err
			}

			// Start the kubevirt controllers and steve server
			if err := kubeServer.StartControllers(); err != nil {
				return nil, fmt.Errorf("failed to start kubevirt controllers: %v", err)
			}

			// Mount kubevirt API under /kubevirt path
			mux.Handle("/kubevirt/", http.StripPrefix("/kubevirt", kubeServer.Handler))
		}

		return mux, nil
	}
}
