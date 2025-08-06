package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/clientcommon"
	kcmd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/cmd/k8s-services/cmd"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rootCmd := kcmd.NewRootCmd()

	// set up k8s using the proper OpenKubeConfig function that handles in-cluster auth
	kubeconfig := rootCmd.PersistentFlags().String(common.FlagKubeConfig, common.KubeConfigDefaultPath, "kubernetes configuration file path")
	if err := viper.BindPFlag(common.FlagKubeConfig, rootCmd.PersistentFlags().Lookup(common.FlagKubeConfig)); err != nil {
		panic(err)
	}

	logger := logrus.New().WithField("service", "k8s-services").Logger
	config, err := clientcommon.OpenKubeConfig(*kubeconfig, logger)
	if err != nil {
		fmt.Printf("failed to build kubeconfig: %v\n", err)
		os.Exit(1)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeyKubeConfig, config)

	return rootCmd.ExecuteContext(ctx)
}

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}
