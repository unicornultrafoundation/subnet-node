package ip

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	caddr "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"

	clusterClient "github.com/unicornultrafoundation/subnet-node/core/k8s/kube"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/operators/clients/metallb"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
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
			ns := viper.GetString(common.FlagK8sManifestNS)
			poolName := viper.GetString(flagMetalLbPoolName)
			logger := logrus.New().WithField("operator", "ip").Logger

			ctx := cmd.Context()

			opcfg := common.GetOperatorConfigFromViper()
			if !caddr.IsHexAddress(opcfg.ProviderAddress) {
				return fmt.Errorf("provider address must valid hex address")
			}

			client, err := clusterClient.NewClient(cmd.Context(), logger, ns)
			if err != nil {
				return err
			}

			metalLbEndpoint, err := common.GetServiceEndpointFlagValue(logger, serviceMetalLb)
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

	// Register the --rest-port flag
	cmd.Flags().Uint16(common.FlagRESTPort, 8082, "REST API port")
	if err := viper.BindPFlag(common.FlagRESTPort, cmd.Flags().Lookup(common.FlagRESTPort)); err != nil {
		panic(err)
	}

	if err := common.AddServiceEndpointFlag(cmd, serviceMetalLb); err != nil {
		return nil
	}

	cmd.Flags().String(flagMetalLbPoolName, "", "metal LB ip address pool to use")
	err := viper.BindPFlag(flagMetalLbPoolName, cmd.Flags().Lookup(flagMetalLbPoolName))
	if err != nil {
		panic(err)
	}

	return cmd
}
