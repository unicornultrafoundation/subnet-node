package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/viper"
	kcmd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/cmd/k8s-services/cmd"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	"k8s.io/client-go/tools/clientcmd"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rootCmd := kcmd.NewRootCmd()

	// set up k8s
	kubeconfig := rootCmd.PersistentFlags().String(common.FlagKubeConfig, common.KubeConfigDefaultPath, "kubernetes configuration file path")
	if err := viper.BindPFlag(common.FlagKubeConfig, rootCmd.PersistentFlags().Lookup(common.FlagKubeConfig)); err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
