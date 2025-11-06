package common

import (
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kutil "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/util"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

const (
	flagProviderAddress = "provider"
	FlagRESTPort        = "rest-port"
	FlagRESTAddress     = "rest-address"
	FlagGRPCPort        = "grpc-port"
	FlagPod             = "pod-name"
	FlagNamespace       = "pod-namespace"
)

const (
	FlagK8sManifestNS = "k8s-manifest-ns"

	FlagListenAddress = "listen"
	FlagPruneInterval = "prune-interval"

	FlagWebRefreshInterval = "web-refresh-interval"
	FlagRetryDelay         = "retry-delay"

	FlagKubeConfig = "kubeconfig"
)

func AddOperatorFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagK8sManifestNS, "subnet-services", "Cluster manifest namespace")
	if err := viper.BindPFlag(FlagK8sManifestNS, cmd.Flags().Lookup(FlagK8sManifestNS)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(FlagK8sManifestNS, "AP_K8S_MANIFEST_NS"); err != nil {
		panic(err)
	}

	cmd.Flags().Duration(FlagPruneInterval, 10*time.Minute, "data pruning interval")
	if err := viper.BindPFlag(FlagPruneInterval, cmd.Flags().Lookup(FlagPruneInterval)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(FlagPruneInterval, "AP_PRUNE_INTERVAL"); err != nil {
		panic(err)
	}

	cmd.Flags().Duration(FlagWebRefreshInterval, 5*time.Second, "web data refresh interval")
	if err := viper.BindPFlag(FlagWebRefreshInterval, cmd.Flags().Lookup(FlagWebRefreshInterval)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(FlagWebRefreshInterval, "AP_WEB_REFRESH_INTERVAL"); err != nil {
		panic(err)
	}

	cmd.Flags().Duration(FlagRetryDelay, 3*time.Second, "retry delay")
	if err := viper.BindPFlag(FlagRetryDelay, cmd.Flags().Lookup(FlagRetryDelay)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(FlagRetryDelay, "AP_RETRY_DELAY"); err != nil {
		panic(err)
	}
}

func AddProviderFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagProviderAddress, "", "address of associated provider in bech32")
	if err := viper.BindPFlag(flagProviderAddress, cmd.Flags().Lookup(flagProviderAddress)); err != nil {
		panic(err)
	}
}

func DetectPort(ctx context.Context, flags *flag.FlagSet, flag string, container, portName string) (int, error) {
	var port uint16

	log := fromctx.LogrFromCtx(ctx)

	if flags.Changed(flag) || !kutil.IsInsideKubernetes() {
		var err error
		port, err = flags.GetUint16(flag)
		if err != nil {
			return 0, err
		}

		return int(port), nil
	}

	if kutil.IsInsideKubernetes() {
		podName := viper.GetString(FlagPod)
		namespace := viper.GetString(FlagNamespace)

		kc, err := fromctx.KubeClientFromCtx(ctx)
		if err != nil {
			return 0, err
		}

		pod, err := kc.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return 0, err
		}

	loop:
		for _, scontainer := range pod.Spec.Containers {
			if scontainer.Name == container {
				for _, cport := range scontainer.Ports {
					if cport.Name == portName {
						port = uint16(cport.ContainerPort) // nolint: gosec
						break loop
					}
				}
			}
		}
	}

	if port == 0 {
		log.Info("unable to detect rest port from pod. falling back to default \"\". default might differ from what service expects!")
		var err error
		port, err = flags.GetUint16(flag)
		if err != nil {
			return 0, err
		}
	}

	return int(port), nil
}
