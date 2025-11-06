package hostname

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	providerflags "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "hostname",
		Short:        "kubernetes operator interfacing with k8s nginx ingress",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			group := fromctx.MustErrGroupFromCtx(ctx)

			ns := viper.GetString(providerflags.FlagK8sManifestNS)

			config := common.GetOperatorConfigFromViper()

			logger := fromctx.LogrFromCtx(ctx).WithField("op", "hostname").Logger

			restPort, err := common.DetectPort(ctx, cmd.Flags(), common.FlagRESTPort, "operator-hostname", "rest")
			if err != nil {
				return err
			}

			listenAddress := viper.GetString(common.FlagRESTAddress)
			restAddr := fmt.Sprintf("%s:%d", listenAddress, restPort)

			op, err := newHostnameOperator(ctx, logger, ns, config, common.IgnoreListConfigFromViper())
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

			group.Go(op.run)

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

	return cmd
}
