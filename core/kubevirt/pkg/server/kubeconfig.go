package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rancher/wrangler/v3/pkg/kubeconfig"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfig(kubeConfig string) (clientcmd.ClientConfig, error) {

	if kubeConfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		kubeConfig = filepath.Join(homeDir, ".kube", "config")
	}

	if isManual(kubeConfig) {
		return kubeconfig.GetNonInteractiveClientConfig(kubeConfig), nil
	}

	return getEmbedded()
}

func isManual(kubeConfig string) bool {
	if kubeConfig != "" {
		return true
	}
	_, inClusterErr := rest.InClusterConfig()
	return inClusterErr == nil
}

func getEmbedded() (clientcmd.ClientConfig, error) {
	return nil, fmt.Errorf("embedded only supported on linux")
}
