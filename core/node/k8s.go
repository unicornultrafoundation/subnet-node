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
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/k8s"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"

	// kubeip "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/ip"
	kubeinventory "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/inventory"
	kubeip "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/ip"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"

	// "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/waiter"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	ptypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
	"go.uber.org/fx"
)

// DeployerService provides a lifecycle-managed Deployer service
func K8sService(lc fx.Lifecycle, cfg *config.C, account *account.AccountService) (k8s.Service, error) {
	ctx := context.Background()
	logger := logrus.New().WithField("service", "k8s").Logger

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

	// temporary inventory
	inventory := kubeinventory.NewNull(ctx, "nodeA")
	ctx = context.WithValue(ctx, cfromctx.CtxKeyClientInventory, inventory)

	// metalLbEndpoint, err := common.GetServiceEndpointFlagValue(logger, "metal-lb")
	// if err != nil {
	// 	return nil, err
	// }
	// ip, err := kubeip.NewClient(ctx, logger, metalLbEndpoint)
	// if err != nil {
	// 	return nil, err
	// }
	ip := kubeip.NewNullClient()
	ctx = context.WithValue(ctx, cfromctx.CtxKeyClientIP, ip)

	group, ctx := errgroup.WithContext(ctx)

	startupch := make(chan struct{}, 1)
	ctx = context.WithValue(ctx, fromctx.CtxKeyStartupCh, (chan<- struct{})(startupch))

	pctx, pcancel := context.WithCancel(ctx)

	ctx = context.WithValue(ctx, fromctx.CtxKeyErrGroup, group)
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
	kubeSettings := builder.NewDefaultSettings()

	client, err := kube.NewClient(ctx, logger, "subnet-services")
	if err != nil {
		return nil, err
	}

	providerID := account.GetAddress()
	provider := &ptypes.Provider{
		Owner: strings.ToLower(providerID.Hex()),
	}

	waitClients := make([]waiter.Waitable, 0)
	waiter := waiter.NewOperatorWaiter(ctx, logger, waitClients...)
	session := session.New(logger, nil, provider, 0)
	bus := pubsub.NewBus()
	k8sCfg := k8s.NewDefaultConfig()
	k8sCfg.InventoryExternalPortQuantity = 10000
	k8sCfg.ClusterSettings = map[interface{}]interface{}{
		builder.SettingsKey: kubeSettings,
	}

	service, err := k8s.NewService(ctx, session, bus, client, waiter, k8sCfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return nil
		},
		OnStop: func(_ context.Context) error {
			return service.Close()
		},
	})

	return service, nil
}
