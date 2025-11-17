package ip

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ecommon "github.com/ethereum/go-ethereum/common"
	clusterClient "github.com/unicornultrafoundation/subnet-node/core/k8s/kube"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/metallb"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	providerflags "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

const (
	flagMetalLbPoolName = "metal-lb-pool"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ip",
		Short:        "kubernetes operator interfacing with Metal LB",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns := viper.GetString(providerflags.FlagK8sManifestNS)
			poolName := viper.GetString(flagMetalLbPoolName)

			ctx := cmd.Context()

			logger := fromctx.LogrFromCtx(ctx).WithField("operator", "ip").Logger

			opcfg := common.GetOperatorConfigFromViper()
			if !ecommon.IsHexAddress(opcfg.ProviderAddress) {
				return fmt.Errorf("provider address must be a valid hex address")
			}

			client, err := clusterClient.NewClient(cmd.Context(), logger, ns)
			if err != nil {
				return err
			}

			metalLbEndpoint, err := providerflags.GetServiceEndpointFlagValue(logger, serviceMetalLb)
			if err != nil {
				return err
			}

			mllbc, err := metallb.NewClient(ctx, logger, poolName, metalLbEndpoint)
			if err != nil {
				return err
			}

			restPort, err := common.DetectPort(ctx, cmd.Flags(), common.FlagRESTPort, "operator-ip", "rest")
			if err != nil {
				return err
			}

			listenAddress := viper.GetString(common.FlagRESTAddress)
			restAddr := fmt.Sprintf("%s:%d", listenAddress, restPort)

			group := fromctx.MustErrGroupFromCtx(ctx)

			logger.Info("clients", "kube", client, "metallb", mllbc)

			op, err := newIPOperator(ctx, logger, ns, opcfg, common.IgnoreListConfigFromViper(), mllbc)
			if err != nil {
				return err
			}

			router := op.server.GetRouter()

			// fixme ovrclk/engineering#609
			// nolint: gosec
			srv := http.Server{Addr: restAddr, Handler: router}

			group.Go(func() error {
				logger.Info("HTTP listening", "address", restAddr)
				return srv.ListenAndServe()
			})

			group.Go(func() error {
				<-ctx.Done()
				_ = srv.Close()

				return ctx.Err()
			})

			group.Go(func() error {
				return op.run(ctx)
			})

			fromctx.MustStartupChFromCtx(ctx) <- struct{}{}
			err = group.Wait()

			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}

	common.AddOperatorFlags(cmd)
	common.AddIgnoreListFlags(cmd)
	common.AddProviderFlag(cmd)

	if err := providerflags.AddServiceEndpointFlag(cmd, serviceMetalLb); err != nil {
		return nil
	}

	cmd.Flags().String(flagMetalLbPoolName, "", "metal LB ip address pool to use")
	err := viper.BindPFlag(flagMetalLbPoolName, cmd.Flags().Lookup(flagMetalLbPoolName))
	if err != nil {
		panic(err)
	}

	return cmd
}
