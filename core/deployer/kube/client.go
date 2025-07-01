package kube

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubeClient struct {
	Client             kubernetes.Interface
	Config             *rest.Config
	Logger             *logrus.Logger
	DefaultServiceType string
	LocalhostEnabled   bool
	MetricsClient      metricsclient.Interface
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

	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &KubeClient{
		Client:             clientset,
		Config:             config,
		Logger:             logger,
		DefaultServiceType: defaultServiceType,
		LocalhostEnabled:   localhostEnabled,
		MetricsClient:      metricsClient,
	}, nil
}
