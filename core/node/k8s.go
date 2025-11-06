package node

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sirupsen/logrus"
	tpubsub "github.com/troian/pubsub"
	"github.com/unicornultrafoundation/subnet-node/bidengine"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/k8s"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube"
	ipmanager "github.com/unicornultrafoundation/subnet-node/core/vpn/ip_manager"

	kubehostname "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/hostname"
	kubeinventory "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/inventory"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"

	providerflags "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/waiter"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	ptypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
	"go.uber.org/fx"
)

const (
	serviceIPOperator       = "ip-operator"
	serviceHostnameOperator = "hostname-operator"
)

// DeployerService provides a lifecycle-managed Deployer service
func K8sService(lc fx.Lifecycle, cfg *config.C, account *account.AccountService, bidengine *bidengine.BidEngine, ipManager ipmanager.IPManager) (k8s.Service, error) {
	if !cfg.GetBool("deployer.enable", false) {
		return nil, nil
	}
	ctx := context.Background()
	logger := logrus.New().WithField("service", "k8s").Logger
	ctx = context.WithValue(ctx, fromctx.CtxKeyLogc, logger)

	// set up k8s
	kubeconfig := cfg.GetString("deployer.kubeconfig_path", "")
	if kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig is not set")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeyKubeConfig, config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeyKubeClientSet, clientset)

	subnetClientset, err := subnetclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet clientset: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeySubnetClientSet, subnetClientset)

	endpoint, err := providerflags.GetServiceEndpointFlagValue(logger, serviceHostnameOperator)
	if err != nil {
		return nil, fmt.Errorf("failed to get service endpoint for hostname operator: %w", err)
	}
	hostnameOperatorClient, err := kubehostname.NewClient(ctx, logger, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create hostname client: %w", err)
	}
	ctx = context.WithValue(ctx, cfromctx.CtxKeyClientHostname, hostnameOperatorClient)

	inventory, err := kubeinventory.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory client: %w", err)
	}
	ctx = context.WithValue(ctx, cfromctx.CtxKeyClientInventory, inventory)

	waitClients := make([]waiter.Waitable, 0)
	waitClients = append(waitClients, hostnameOperatorClient)

	operatorWaiter := waiter.NewOperatorWaiter(ctx, logger, waitClients...)

	group, ctx := errgroup.WithContext(ctx)
	ctx = context.WithValue(ctx, fromctx.CtxKeyErrGroup, group)

	startupch := make(chan struct{}, 1)
	ctx = context.WithValue(ctx, fromctx.CtxKeyStartupCh, (chan<- struct{})(startupch))

	pctx, pcancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, fromctx.CtxKeyPubSub, tpubsub.New(pctx, 1000))

	go func() {
		defer pcancel()

		select {
		case <-ctx.Done():
			return
		case <-startupch:
		}

		_ = group.Wait()
	}()

	client, err := kube.NewClient(ctx, logger, "subnet-services")
	if err != nil {
		return nil, err
	}

	providerID := account.GetAddress()
	provider := &ptypes.Provider{
		Owner: strings.ToLower(providerID.Hex()),
	}

	session := session.New(logger, provider)
	bus := pubsub.NewBus()

	service, err := k8s.NewServiceFromConfig(ctx, session, bus, client, cfg, bidengine, account.GetClient(), operatorWaiter)
	if err != nil {
		return nil, err
	}

	var observer ipmanager.IPManagerObserver
	if ipManager != nil {
		observer, err = ipManager.WatchIP()
		if err != nil {
			return nil, err
		}
	} else {
		observer = nil
		logger.Warn("No IP manager provided, consider enabling the VPN to expose the deployment services")
	}
	stopCh := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// start watching for IP updates
			if observer != nil {
				go func() {
					for {
						select {
						case ip := <-observer.GetChannel():
							// Update the virtual IP in the client
							client.SetVirtualIP(ip)
						case <-ctx.Done():
							return
						case <-stopCh:
							return
						}
					}
				}()
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			close(stopCh)
			if observer != nil {
				observer.Close()
			}
			return service.Close()
		},
	})

	return service, nil
}
