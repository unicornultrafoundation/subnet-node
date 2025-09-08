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

	kubeinventory "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/inventory"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	ptypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
	"go.uber.org/fx"
)

// DeployerService provides a lifecycle-managed Deployer service
func K8sService(lc fx.Lifecycle, cfg *config.C, account *account.AccountService, bidengine *bidengine.BidEngine) (k8s.Service, error) {
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

	inventory, err := kubeinventory.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory client: %w", err)
	}
	ctx = context.WithValue(ctx, cfromctx.CtxKeyClientInventory, inventory)

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

	client, err := kube.NewClient(ctx, logger, "subnet-services", cfg.GetString("vpn.virtual_ip", "localhost"))
	if err != nil {
		return nil, err
	}

	providerID := account.GetAddress()
	provider := &ptypes.Provider{
		Owner: strings.ToLower(providerID.Hex()),
	}

	session := session.New(logger, provider)
	bus := pubsub.NewBus()

	service, err := k8s.NewServiceFromConfig(ctx, session, bus, client, cfg, bidengine, account.GetClient())
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
