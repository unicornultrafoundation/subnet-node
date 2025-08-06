package operator

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/troian/pubsub"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/clientcommon"
	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	providerflags "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/inventory"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

func OperatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "operator",
		Short:        "kubernetes operators control",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			log := logrus.New()
			log.SetLevel(logrus.DebugLevel)
			fromctx.CmdSetContextValue(cmd, fromctx.CtxKeyLogc, log)

			group, ctx := errgroup.WithContext(cmd.Context())

			kubecfg, err := fromctx.KubeConfigFromCtx(cmd.Context())
			if err != nil {
				if err := clientcommon.SetKubeConfigToCmd(cmd); err != nil {
					return err
				}

				kubecfg = fromctx.MustKubeConfigFromCtx(cmd.Context())
			}

			if val := ctx.Value(fromctx.CtxKeyKubeClientSet); val == nil {
				kc, err := kubernetes.NewForConfig(kubecfg)
				if err != nil {
					return err
				}
				fromctx.CmdSetContextValue(cmd, fromctx.CtxKeyKubeClientSet, kubernetes.Interface(kc))
			}

			if val := ctx.Value(fromctx.CtxKeySubnetClientSet); val == nil {
				ac, err := subnetclientset.NewForConfig(kubecfg)
				if err != nil {
					return err
				}

				fromctx.CmdSetContextValue(cmd, fromctx.CtxKeySubnetClientSet, subnetclientset.Interface(ac))
			}

			startupch := make(chan struct{}, 1)
			fromctx.CmdSetContextValue(cmd, fromctx.CtxKeyStartupCh, (chan<- struct{})(startupch))

			pctx, pcancel := context.WithCancel(context.Background())

			fromctx.CmdSetContextValue(cmd, fromctx.CtxKeyErrGroup, group)
			fromctx.CmdSetContextValue(cmd, fromctx.CtxKeyPubSub, pubsub.New(pctx, 1000))

			go func() {
				defer pcancel()

				select {
				case <-ctx.Done():
					return
				case <-startupch:
				}

				_ = group.Wait()
			}()

			return nil
		},
	}

	cmd.PersistentFlags().Uint16(common.FlagRESTPort, 8080, "REST port")
	if err := viper.BindPFlag(common.FlagRESTPort, cmd.PersistentFlags().Lookup(common.FlagRESTPort)); err != nil {
		panic(err)
	}

	cmd.PersistentFlags().Uint16(common.FlagGRPCPort, 8081, "REST port")
	if err := viper.BindPFlag(common.FlagGRPCPort, cmd.PersistentFlags().Lookup(common.FlagGRPCPort)); err != nil {
		panic(err)
	}

	cmd.PersistentFlags().String(common.FlagRESTAddress, "0.0.0.0", "listen address for REST server")
	if err := viper.BindPFlag(common.FlagRESTAddress, cmd.PersistentFlags().Lookup(common.FlagRESTAddress)); err != nil {
		panic(err)
	}

	err := providerflags.AddKubeConfigPathFlag(cmd)
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(inventory.Cmd())

	return cmd
}

func ToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tools",
		Short:        "debug tools",
		SilenceUsage: true,
	}

	cmd.AddCommand(cmdPsutil())

	return cmd
}
