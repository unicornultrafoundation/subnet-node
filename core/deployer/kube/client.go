package kube

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	Client             kubernetes.Interface
	Logger             *logrus.Logger
	DefaultServiceType string
	LocalhostEnabled   bool
}

func NewKubeClient(ctx context.Context, kubeconfig string, logger *logrus.Logger, defaultServiceType string, localhostEnabled bool) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &KubeClient{
		Client:             clientset,
		Logger:             logger,
		DefaultServiceType: defaultServiceType,
		LocalhostEnabled:   localhostEnabled,
	}, nil
}
