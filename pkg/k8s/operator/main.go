package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"

	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"
	cip "github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/ip"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd := cip.Cmd()

	kubeconfig := "/home/duchuong.linux/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeyKubeConfig, config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeyKubeClientSet, clientset)

	subnetClientset, err := subnetclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create subnet clientset: %w", err)
	}
	ctx = context.WithValue(ctx, fromctx.CtxKeySubnetClientSet, subnetClientset)

	group, ctx := errgroup.WithContext(ctx)

	startupch := make(chan struct{}, 1)
	ctx = context.WithValue(ctx, fromctx.CtxKeyStartupCh, (chan<- struct{})(startupch))

	ctx = context.WithValue(ctx, fromctx.CtxKeyErrGroup, group)

	return cmd.ExecuteContext(ctx)
}

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}
