package kube

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setupIntegrationTest creates a real KubeClient and returns a cleanup function
func setupIntegrationTest(t *testing.T) (*KubeClient, func()) {
	// Check if kubeconfig exists
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Skip("Skipping integration test: kubeconfig not found")
	}

	// Create logger
	logger := logrus.New().WithField("service", "kube").Logger

	// Create client
	client, err := NewKubeClient(context.Background(), kubeconfig, logger, "NodePort", true)
	require.NoError(t, err)

	// Return cleanup function
	return client, func() {
		// Clean up test namespaces
		namespaces := []string{"test-ns-1", "test-ns-2", "test-ns-3", "test-integration-1", "test-dep-2", "test-integration-3"}
		for _, ns := range namespaces {
			// Delete all resources in the namespace first
			// Delete services
			services, err := client.Client.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
			if err == nil {
				for _, svc := range services.Items {
					if err := client.Client.CoreV1().Services(ns).Delete(context.Background(), svc.Name, metav1.DeleteOptions{}); err != nil {
						t.Logf("Warning: Failed to delete service %s in namespace %s: %v", svc.Name, ns, err)
					}
				}
			}

			// Delete deployments
			deployments, err := client.Client.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
			if err == nil {
				for _, dep := range deployments.Items {
					if err := client.Client.AppsV1().Deployments(ns).Delete(context.Background(), dep.Name, metav1.DeleteOptions{}); err != nil {
						t.Logf("Warning: Failed to delete deployment %s in namespace %s: %v", dep.Name, ns, err)
					}
				}
			}

			// Delete ingresses
			ingresses, err := client.Client.NetworkingV1().Ingresses(ns).List(context.Background(), metav1.ListOptions{})
			if err == nil {
				for _, ing := range ingresses.Items {
					if err := client.Client.NetworkingV1().Ingresses(ns).Delete(context.Background(), ing.Name, metav1.DeleteOptions{}); err != nil {
						t.Logf("Warning: Failed to delete ingress %s in namespace %s: %v", ing.Name, ns, err)
					}
				}
			}

			// Delete secrets
			secrets, err := client.Client.CoreV1().Secrets(ns).List(context.Background(), metav1.ListOptions{})
			if err == nil {
				for _, secret := range secrets.Items {
					if err := client.Client.CoreV1().Secrets(ns).Delete(context.Background(), secret.Name, metav1.DeleteOptions{}); err != nil {
						t.Logf("Warning: Failed to delete secret %s in namespace %s: %v", secret.Name, ns, err)
					}
				}
			}

			// Wait for resources to be deleted
			time.Sleep(2 * time.Second)

			// Delete the namespace
			if err := client.Client.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil {
				if !errors.IsNotFound(err) {
					t.Logf("Warning: Failed to delete namespace %s: %v", ns, err)
				}
			}
		}

		// Wait for namespaces to be deleted
		for _, ns := range namespaces {
			for i := 0; i < 10; i++ {
				_, err := client.Client.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					break
				}
				time.Sleep(time.Second)
			}
		}
	}
}

func TestIntegrationDeploy(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Test deployment with a simple service
	deployment := types.Deployment{
		ID: "test-integration-1",
		Manifest: &manifest.Manifest{
			Groups: []manifest.Group{
				{
					Name: "test-group",
					Services: []manifest.Service{
						{
							Image: "nginx",
							Count: 1,
							Resources: &manifest.ServiceResources{
								CPU: &manifest.CPUResource{
									Units: manifest.ResourceValue{
										Value: 100,
										Unit:  "m",
									},
								},
								Memory: &manifest.MemoryResource{
									Size: manifest.ResourceValue{
										Value: 128,
										Unit:  "Mi",
									},
								},
							},
							Expose: []manifest.Expose{
								{
									Port:  80,
									As:    80,
									Proto: "TCP",
									To:    []manifest.To{{Global: true}},
								},
							},
						},
					},
				},
			},
			Profiles: manifest.Profiles{
				Compute: map[string]manifest.ComputeProfile{
					"default": {
						Resources: manifest.ResourceRequirements{
							CPU: manifest.ResourceLimit{
								Request: "100m",
								Limit:   "200m",
							},
							Memory: manifest.ResourceLimit{
								Request: "128Mi",
								Limit:   "256Mi",
							},
						},
					},
				},
			},
		},
	}

	// Deploy the service
	err := client.Deploy(context.Background(), deployment)
	assert.NoError(t, err)

	// Wait for deployment to be ready
	time.Sleep(5 * time.Second)

	// Verify deployment exists
	deployments, err := client.Client.AppsV1().Deployments(deployment.ID).List(context.Background(), metav1.ListOptions{
		LabelSelector: "deploymentID=test-integration-1",
	})
	assert.NoError(t, err)
	assert.Len(t, deployments.Items, 1)
	assert.Equal(t, "nginx", deployments.Items[0].Spec.Template.Spec.Containers[0].Image)

	// Verify service exists
	services, err := client.Client.CoreV1().Services(deployment.ID).List(context.Background(), metav1.ListOptions{
		LabelSelector: "deploymentID=test-integration-1",
	})
	assert.NoError(t, err)
	assert.Len(t, services.Items, 1)
	assert.Equal(t, int32(80), services.Items[0].Spec.Ports[0].Port)
}

func TestIntegrationServiceDependencies(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create a test deployment with service dependencies
	deployment := types.Deployment{
		ID: "test-dep-2",
		Manifest: &manifest.Manifest{
			Groups: []manifest.Group{
				{
					Name: "test-group",
					Services: []manifest.Service{
						{
							Image: "nginx",
							Count: 1,
							Resources: &manifest.ServiceResources{
								CPU: &manifest.CPUResource{
									Units: manifest.ResourceValue{
										Value: 100,
										Unit:  "m",
									},
								},
								Memory: &manifest.MemoryResource{
									Size: manifest.ResourceValue{
										Value: 128,
										Unit:  "Mi",
									},
								},
							},
						},
						{
							Image:     "redis",
							Count:     1,
							DependsOn: []string{"nginx"},
							Resources: &manifest.ServiceResources{
								CPU: &manifest.CPUResource{
									Units: manifest.ResourceValue{
										Value: 100,
										Unit:  "m",
									},
								},
								Memory: &manifest.MemoryResource{
									Size: manifest.ResourceValue{
										Value: 128,
										Unit:  "Mi",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Deploy the services
	err := client.Deploy(context.Background(), deployment)
	assert.NoError(t, err)

	// Wait for deployments to be ready
	time.Sleep(5 * time.Second)

	// Verify both deployments exist
	deployments, err := client.Client.AppsV1().Deployments(deployment.ID).List(context.Background(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, deployments.Items, 2)

	// Verify the dependency relationship in the manifest
	var nginxDeployment, redisDeployment *appsv1.Deployment
	for i := range deployments.Items {
		dep := &deployments.Items[i]
		if dep.Spec.Template.Spec.Containers[0].Image == "nginx" {
			nginxDeployment = dep
		} else if dep.Spec.Template.Spec.Containers[0].Image == "redis" {
			redisDeployment = dep
		}
	}

	// Verify both deployments were found
	assert.NotNil(t, nginxDeployment, "nginx deployment not found")
	assert.NotNil(t, redisDeployment, "redis deployment not found")

	// Verify the dependency relationship in the manifest
	assert.Contains(t, deployment.Manifest.Groups[0].Services[1].DependsOn, "nginx", "redis service should depend on nginx")

	// Verify both deployments are ready
	assert.Equal(t, int32(1), *nginxDeployment.Spec.Replicas, "nginx deployment should have 1 replica")
	assert.Equal(t, int32(1), *redisDeployment.Spec.Replicas, "redis deployment should have 1 replica")
}

func TestIntegrationHealthChecks(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Test deployment with health checks
	deployment := types.Deployment{
		ID: "test-integration-3",
		Manifest: &manifest.Manifest{
			Groups: []manifest.Group{
				{
					Name: "test-group",
					Services: []manifest.Service{
						{
							Image: "nginx",
							Count: 1,
							Resources: &manifest.ServiceResources{
								CPU: &manifest.CPUResource{
									Units: manifest.ResourceValue{
										Value: 100,
										Unit:  "m",
									},
								},
								Memory: &manifest.MemoryResource{
									Size: manifest.ResourceValue{
										Value: 128,
										Unit:  "Mi",
									},
								},
							},
							Params: &manifest.ServiceParams{
								Health: &manifest.HealthParams{
									Readiness: &manifest.HealthCheck{
										HTTP: &manifest.HTTPHealthCheck{
											Path: "/",
											Port: 80,
										},
										InitialDelaySeconds: 5,
										PeriodSeconds:       10,
										TimeoutSeconds:      5,
										SuccessThreshold:    1,
										FailureThreshold:    3,
									},
									Liveness: &manifest.HealthCheck{
										HTTP: &manifest.HTTPHealthCheck{
											Path: "/",
											Port: 80,
										},
										InitialDelaySeconds: 15,
										PeriodSeconds:       20,
										TimeoutSeconds:      5,
										SuccessThreshold:    1,
										FailureThreshold:    3,
									},
								},
							},
						},
					},
				},
			},
			Profiles: manifest.Profiles{
				Compute: map[string]manifest.ComputeProfile{
					"default": {
						Resources: manifest.ResourceRequirements{
							CPU: manifest.ResourceLimit{
								Request: "100m",
								Limit:   "200m",
							},
							Memory: manifest.ResourceLimit{
								Request: "128Mi",
								Limit:   "256Mi",
							},
						},
					},
				},
			},
		},
	}

	// Deploy the service
	err := client.Deploy(context.Background(), deployment)
	assert.NoError(t, err)

	// Wait for deployment to be ready
	time.Sleep(10 * time.Second)

	// Verify deployment exists with health checks
	deployments, err := client.Client.AppsV1().Deployments(deployment.ID).List(context.Background(), metav1.ListOptions{
		LabelSelector: "deploymentID=test-integration-3",
	})
	assert.NoError(t, err)
	assert.Len(t, deployments.Items, 1)

	// Verify health check configuration
	container := deployments.Items[0].Spec.Template.Spec.Containers[0]
	assert.NotNil(t, container.ReadinessProbe)
	assert.NotNil(t, container.LivenessProbe)
	assert.Equal(t, "/", container.ReadinessProbe.HTTPGet.Path)
	assert.Equal(t, int32(80), container.ReadinessProbe.HTTPGet.Port.IntVal)
}
