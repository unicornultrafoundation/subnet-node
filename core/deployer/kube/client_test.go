package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewKubeClient is an integration test that uses the real kubeconfig if present.
func TestNewKubeClient(t *testing.T) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Skip("Skipping integration test: kubeconfig not found at ~/.kube/config")
	}

	tests := []struct {
		name       string
		kubeconfig string
		wantErr    bool
	}{
		{
			name:       "valid kubeconfig",
			kubeconfig: kubeconfig,
			wantErr:    false,
		},
		{
			name:       "invalid kubeconfig",
			kubeconfig: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New().WithField("service", "kube").Logger
			client, err := NewKubeClient(context.Background(), tt.kubeconfig, logger, "NodePort", true)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestDeploy(t *testing.T) {
	// Create a fake Kubernetes client
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name         string
		deployment   types.Deployment
		wantErr      bool
		setupFunc    func()             // Additional setup if needed
		validateFunc func(t *testing.T) // Validation after deployment
	}{
		{
			name: "valid deployment",
			deployment: types.Deployment{
				ID: "test-deployment",
				Manifest: &manifest.Manifest{
					Groups: []manifest.Group{
						{
							Name: "test-group",
							Services: []manifest.Service{
								{
									Image: "test-image",
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
										Storage: &manifest.StorageResource{
											Size: manifest.ResourceValue{
												Value: 1,
												Unit:  "Gi",
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
			},
			wantErr: false,
			validateFunc: func(t *testing.T) {
				// Verify namespace was created
				ns, err := fakeClient.CoreV1().Namespaces().Get(context.Background(), "test-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.NotNil(t, ns)

				// Verify deployment was created
				deployments, err := fakeClient.AppsV1().Deployments("test-deployment").List(context.Background(), metav1.ListOptions{})
				assert.NoError(t, err)
				assert.Len(t, deployments.Items, 1)
				assert.Equal(t, "test-image", deployments.Items[0].Spec.Template.Spec.Containers[0].Image)
			},
		},
		{
			name: "deployment with service dependencies",
			deployment: types.Deployment{
				ID: "test-deployment-2",
				Manifest: &manifest.Manifest{
					Groups: []manifest.Group{
						{
							Name: "test-group-2",
							Services: []manifest.Service{
								{
									Image: "service-1",
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
										Storage: &manifest.StorageResource{
											Size: manifest.ResourceValue{
												Value: 1,
												Unit:  "Gi",
											},
										},
									},
								},
								{
									Image:     "service-2",
									Count:     1,
									DependsOn: []string{"service-1"},
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
										Storage: &manifest.StorageResource{
											Size: manifest.ResourceValue{
												Value: 1,
												Unit:  "Gi",
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
			},
			wantErr: false,
			validateFunc: func(t *testing.T) {
				// Verify both services were deployed
				deployments, err := fakeClient.AppsV1().Deployments("test-deployment-2").List(context.Background(), metav1.ListOptions{})
				assert.NoError(t, err)
				assert.Len(t, deployments.Items, 2)

				// Verify service-2 was deployed after service-1
				var service1Deployed, service2Deployed bool
				for _, dep := range deployments.Items {
					if dep.Spec.Template.Spec.Containers[0].Image == "service-1" {
						service1Deployed = true
					}
					if dep.Spec.Template.Spec.Containers[0].Image == "service-2" {
						service2Deployed = true
					}
				}
				assert.True(t, service1Deployed)
				assert.True(t, service2Deployed)
			},
		},
		{
			name: "deployment with exposed ports",
			deployment: types.Deployment{
				ID: "test-deployment-3",
				Manifest: &manifest.Manifest{
					Groups: []manifest.Group{
						{
							Name: "test-group-3",
							Services: []manifest.Service{
								{
									Image: "test-service",
									Count: 1,
									Expose: []manifest.Expose{
										{
											Port:  8080,
											As:    80,
											Proto: "TCP",
											To:    []manifest.To{{Global: true}},
										},
									},
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
										Storage: &manifest.StorageResource{
											Size: manifest.ResourceValue{
												Value: 1,
												Unit:  "Gi",
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
			},
			wantErr: false,
			validateFunc: func(t *testing.T) {
				// Verify service was created
				services, err := fakeClient.CoreV1().Services("test-deployment-3").List(context.Background(), metav1.ListOptions{})
				assert.NoError(t, err)
				assert.Len(t, services.Items, 1)
				assert.Equal(t, int32(80), services.Items[0].Spec.Ports[0].Port)
				assert.Equal(t, int32(8080), services.Items[0].Spec.Ports[0].TargetPort.IntVal)
			},
		},
		{
			name: "deployment with invalid manifest",
			deployment: types.Deployment{
				ID: "test-deployment-4",
				Manifest: &manifest.Manifest{
					Groups: []manifest.Group{
						{
							Name: "test-group-4",
							Services: []manifest.Service{
								{
									Image: "", // Invalid: empty image
									Count: 1,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			err := client.Deploy(context.Background(), tt.deployment)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}
		})
	}
}

func TestCreateRegistrySecret(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name      string
		service   manifest.Service
		namespace string
		wantErr   bool
	}{
		{
			name: "valid registry credentials",
			service: manifest.Service{
				Image: "test-image",
				Credentials: &manifest.RegistryAuth{
					Host:     "test-registry.com",
					Username: "test-user",
					Password: "test-password",
				},
			},
			namespace: "test-namespace",
			wantErr:   false,
		},
		{
			name: "no credentials",
			service: manifest.Service{
				Image: "test-image",
			},
			namespace: "test-namespace",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.createRegistrySecret(context.Background(), tt.service, tt.namespace)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.service.Credentials != nil {
					secret, err := fakeClient.CoreV1().Secrets(tt.namespace).Get(
						context.Background(),
						"registry-auth-"+tt.service.Image,
						metav1.GetOptions{},
					)
					assert.NoError(t, err)
					assert.NotNil(t, secret)
					assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
				}
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name     string
		manifest *manifest.Manifest
		wantErr  bool
	}{
		{
			name: "valid manifest",
			manifest: &manifest.Manifest{
				Groups: []manifest.Group{
					{
						Name: "test-group",
						Services: []manifest.Service{
							{
								Image: "test-image",
								Count: 1,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint name",
			manifest: &manifest.Manifest{
				Groups: []manifest.Group{
					{
						Name: "test-group",
						Services: []manifest.Service{
							{
								Image: "test-image",
								Count: 1,
							},
						},
					},
				},
				Endpoints: map[string]manifest.Endpoint{
					"invalid@name": {
						Kind: "nginx",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.validateManifest(tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name         string
		namespace    string
		version      string
		deploymentID string
		wantErr      bool
	}{
		{
			name:         "create new namespace",
			namespace:    "test-namespace",
			version:      "v1",
			deploymentID: "test-deployment",
			wantErr:      false,
		},
		{
			name:         "create existing namespace",
			namespace:    "test-namespace",
			version:      "v1",
			deploymentID: "test-deployment",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.createNamespace(context.Background(), tt.namespace, tt.version, tt.deploymentID, "test-requester")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				ns, err := fakeClient.CoreV1().Namespaces().Get(context.Background(), tt.namespace, metav1.GetOptions{})
				assert.NoError(t, err)
				assert.NotNil(t, ns)
				assert.Equal(t, tt.version, ns.Labels["version"])
				assert.Equal(t, tt.deploymentID, ns.Labels["deploymentID"])
				assert.Equal(t, "test-requester", ns.Labels["requester"])
			}
		})
	}
}

func TestCreateStorageVolume(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name           string
		service        manifest.Service
		computeProfile *manifest.ComputeProfile
		namespace      string
		versionStr     string
		wantVolume     bool
		wantErr        bool
	}{
		{
			name: "service with storage",
			service: manifest.Service{
				Image: "test-image",
				Resources: &manifest.ServiceResources{
					Storage: &manifest.StorageResource{
						Size: manifest.ResourceValue{
							Value: 1,
							Unit:  "Gi",
						},
					},
				},
			},
			namespace:  "test-namespace",
			versionStr: "v1",
			wantVolume: true,
			wantErr:    false,
		},
		{
			name: "service with compute profile storage",
			service: manifest.Service{
				Image: "test-image",
			},
			computeProfile: &manifest.ComputeProfile{
				Resources: manifest.ResourceRequirements{
					Storage: []manifest.StorageVolume{
						{
							Name: "custom-storage",
							Size: "2Gi",
							Attributes: &manifest.StorageAttributes{
								Persistent: true,
								Class:      "standard",
							},
						},
					},
				},
			},
			namespace:  "test-namespace",
			versionStr: "v1",
			wantVolume: false, // Compute profile storage is not supported in this version
			wantErr:    false,
		},
		{
			name: "service without storage",
			service: manifest.Service{
				Image: "test-image",
			},
			namespace:  "test-namespace",
			versionStr: "v1",
			wantVolume: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, volumeMounts, err := client.createStorageVolume(context.Background(), tt.service, tt.computeProfile, tt.namespace, tt.versionStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantVolume {
					assert.NotNil(t, volume)
					assert.NotNil(t, volumeMounts)
					assert.Len(t, volumeMounts, 1)
				} else {
					assert.Nil(t, volume)
					assert.Nil(t, volumeMounts)
				}
			}
		})
	}
}

func TestCreateHealthCheckProbes(t *testing.T) {
	client := &KubeClient{}

	tests := []struct {
		name          string
		service       manifest.Service
		wantReadiness bool
		wantLiveness  bool
	}{
		{
			name: "service with both probes",
			service: manifest.Service{
				Image: "test-image",
				Params: &manifest.ServiceParams{
					Health: &manifest.HealthParams{
						Readiness: &manifest.HealthCheck{
							HTTP: &manifest.HTTPHealthCheck{
								Path: "/ready",
								Port: 8080,
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       10,
							TimeoutSeconds:      2,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						Liveness: &manifest.HealthCheck{
							HTTP: &manifest.HTTPHealthCheck{
								Path: "/health",
								Port: 8080,
							},
							InitialDelaySeconds: 15,
							PeriodSeconds:       20,
							TimeoutSeconds:      3,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
					},
				},
			},
			wantReadiness: true,
			wantLiveness:  true,
		},
		{
			name: "service without health checks",
			service: manifest.Service{
				Image: "test-image",
			},
			wantReadiness: false,
			wantLiveness:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readiness, liveness := client.createHealthCheckProbes(tt.service)
			if tt.wantReadiness {
				assert.NotNil(t, readiness)
				assert.Equal(t, tt.service.Params.Health.Readiness.HTTP.Path, readiness.HTTPGet.Path)
				assert.Equal(t, int32(tt.service.Params.Health.Readiness.HTTP.Port), readiness.HTTPGet.Port.IntVal)
			} else {
				assert.Nil(t, readiness)
			}
			if tt.wantLiveness {
				assert.NotNil(t, liveness)
				assert.Equal(t, tt.service.Params.Health.Liveness.HTTP.Path, liveness.HTTPGet.Path)
				assert.Equal(t, int32(tt.service.Params.Health.Liveness.HTTP.Port), liveness.HTTPGet.Port.IntVal)
			} else {
				assert.Nil(t, liveness)
			}
		})
	}
}

func TestCreateResourceRequirements(t *testing.T) {
	client := &KubeClient{}

	tests := []struct {
		name           string
		service        manifest.Service
		computeProfile *manifest.ComputeProfile
		wantErr        bool
	}{
		{
			name: "service with resources",
			service: manifest.Service{
				Image: "test-image",
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
			wantErr: false,
		},
		{
			name: "service with compute profile",
			service: manifest.Service{
				Image: "test-image",
			},
			computeProfile: &manifest.ComputeProfile{
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
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := client.createResourceRequirements(tt.service, tt.computeProfile)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resources)
			}
		})
	}
}

func TestCreateContainerPorts(t *testing.T) {
	client := &KubeClient{}

	tests := []struct {
		name      string
		service   manifest.Service
		wantPorts int
	}{
		{
			name: "service with exposed ports",
			service: manifest.Service{
				Image: "test-image",
				Expose: []manifest.Expose{
					{
						Port:  8080,
						Proto: "TCP",
					},
					{
						Port:  9090,
						Proto: "UDP",
					},
				},
			},
			wantPorts: 2,
		},
		{
			name: "service without exposed ports",
			service: manifest.Service{
				Image: "test-image",
			},
			wantPorts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := client.createContainerPorts(tt.service)
			assert.Len(t, ports, tt.wantPorts)
			if tt.wantPorts > 0 {
				assert.Equal(t, int32(8080), ports[0].ContainerPort)
				assert.Equal(t, corev1.Protocol("TCP"), ports[0].Protocol)
			}
		})
	}
}

func TestCreateEnvironmentVariables(t *testing.T) {
	client := &KubeClient{}

	tests := []struct {
		name     string
		service  manifest.Service
		wantVars int
		firstVar *corev1.EnvVar
	}{
		{
			name: "service with environment variables",
			service: manifest.Service{
				Image: "test-image",
				Env: []string{
					"KEY1=value1",
					"KEY2=value2",
				},
			},
			wantVars: 2,
			firstVar: &corev1.EnvVar{
				Name:  "KEY1",
				Value: "value1",
			},
		},
		{
			name: "service without environment variables",
			service: manifest.Service{
				Image: "test-image",
			},
			wantVars: 0,
			firstVar: nil,
		},
		{
			name: "service with invalid environment variables",
			service: manifest.Service{
				Image: "test-image",
				Env: []string{
					"INVALID",
					"KEY=value",
				},
			},
			wantVars: 1,
			firstVar: &corev1.EnvVar{
				Name:  "KEY",
				Value: "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := client.createEnvironmentVariables(tt.service)
			assert.Len(t, vars, tt.wantVars)
			if tt.firstVar != nil {
				assert.Equal(t, tt.firstVar.Name, vars[0].Name)
				assert.Equal(t, tt.firstVar.Value, vars[0].Value)
			}
		})
	}
}

func TestCreateService(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client:             fakeClient,
		Logger:             logger,
		DefaultServiceType: "NodePort",
		LocalhostEnabled:   true,
	}

	tests := []struct {
		name         string
		service      manifest.Service
		group        manifest.Group
		expose       manifest.Expose
		namespace    string
		versionStr   string
		deploymentID string
		wantErr      bool
	}{
		{
			name: "service with HTTP options",
			service: manifest.Service{
				Image: "test-image",
			},
			group: manifest.Group{
				Name: "test-group",
			},
			expose: manifest.Expose{
				Port:  8080,
				As:    80,
				Proto: "TCP",
				To:    []manifest.To{{Global: true}},
				HTTPOptions: &manifest.HTTPOptions{
					MaxBodySize: 1024,
					NextCases:   []string{"case1", "case2"},
				},
			},
			namespace:    "test-namespace-http",
			versionStr:   "v1.0.0-abc",
			deploymentID: "test-deployment",
			wantErr:      false,
		},
		{
			name: "service with accept rules",
			service: manifest.Service{
				Image: "test-image",
			},
			group: manifest.Group{
				Name: "test-group",
			},
			expose: manifest.Expose{
				Port:   8080,
				As:     80,
				Proto:  "TCP",
				To:     []manifest.To{{Global: true}},
				Accept: []string{"10.0.0.0/8", "192.168.0.0/16"},
			},
			namespace:    "test-namespace-accept",
			versionStr:   "v1.0.0-abc",
			deploymentID: "test-deployment",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.createService(context.Background(), tt.service, tt.group, tt.expose, tt.namespace, tt.versionStr, tt.deploymentID, "test-requester")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				svc, err := fakeClient.CoreV1().Services(tt.namespace).Get(
					context.Background(),
					fmt.Sprintf("%s-%s-%d-%s", tt.group.Name, tt.service.Image, tt.expose.Port, tt.versionStr[:8]),
					metav1.GetOptions{},
				)
				assert.NoError(t, err)
				assert.NotNil(t, svc)
				assert.Equal(t, "test-requester", svc.Labels["requester"])
				if tt.expose.HTTPOptions != nil {
					assert.Equal(t, fmt.Sprintf("%d", tt.expose.HTTPOptions.MaxBodySize), svc.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"])
				}
				if len(tt.expose.Accept) > 0 {
					assert.Equal(t, strings.Join(tt.expose.Accept, ","), svc.Annotations["ingress.kubernetes.io/whitelist-source-range"])
				}
			}
		})
	}
}

func TestCreateIngress(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name         string
		service      manifest.Service
		group        manifest.Group
		endpoint     manifest.Endpoint
		endpointName string
		namespace    string
		deploymentID string
		wantErr      bool
	}{
		{
			name: "create ingress with HTTP options",
			service: manifest.Service{
				Image: "test-image",
				Expose: []manifest.Expose{
					{
						Port: 8080,
						HTTPOptions: &manifest.HTTPOptions{
							MaxBodySize: 1024,
							NextCases:   []string{"case1", "case2"},
						},
					},
				},
			},
			group: manifest.Group{
				Name: "test-group",
			},
			endpoint: manifest.Endpoint{
				Kind: "nginx",
			},
			endpointName: "test-endpoint",
			namespace:    "test-namespace",
			deploymentID: "test-deployment",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.createIngress(context.Background(), tt.service, tt.group, tt.endpoint, tt.endpointName, tt.namespace, tt.deploymentID, "test-requester")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				ingress, err := fakeClient.NetworkingV1().Ingresses(tt.namespace).Get(
					context.Background(),
					fmt.Sprintf("%s-%s", tt.group.Name, tt.endpointName),
					metav1.GetOptions{},
				)
				assert.NoError(t, err)
				assert.NotNil(t, ingress)
				assert.Equal(t, "test-requester", ingress.Labels["requester"])
				if tt.service.Expose[0].HTTPOptions != nil {
					assert.Equal(t, fmt.Sprintf("%d", tt.service.Expose[0].HTTPOptions.MaxBodySize), ingress.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"])
				}
			}
		})
	}
}

func TestValidateComputeProfiles(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	tests := []struct {
		name     string
		profiles manifest.Profiles
		wantErr  bool
	}{
		{
			name: "valid compute profile",
			profiles: manifest.Profiles{
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
			wantErr: false,
		},
		{
			name: "invalid compute profile - missing CPU request",
			profiles: manifest.Profiles{
				Compute: map[string]manifest.ComputeProfile{
					"default": {
						Resources: manifest.ResourceRequirements{
							CPU: manifest.ResourceLimit{
								Limit: "200m",
							},
							Memory: manifest.ResourceLimit{
								Request: "128Mi",
								Limit:   "256Mi",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.validateComputeProfiles(context.Background(), tt.profiles)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCleanupResources(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Create test resources
	resources := []struct {
		kind      string
		name      string
		namespace string
	}{
		{kind: "Deployment", name: "test-deployment", namespace: "test-namespace"},
		{kind: "Service", name: "test-service", namespace: "test-namespace"},
		{kind: "Ingress", name: "test-ingress", namespace: "test-namespace"},
		{kind: "Secret", name: "test-secret", namespace: "test-namespace"},
	}

	// Create the resources first
	for _, resource := range resources {
		switch resource.kind {
		case "Deployment":
			_, err := fakeClient.AppsV1().Deployments(resource.namespace).Create(context.Background(), &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: resource.name, Namespace: resource.namespace},
			}, metav1.CreateOptions{})
			assert.NoError(t, err)
		case "Service":
			_, err := fakeClient.CoreV1().Services(resource.namespace).Create(context.Background(), &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: resource.name, Namespace: resource.namespace},
			}, metav1.CreateOptions{})
			assert.NoError(t, err)
		case "Ingress":
			_, err := fakeClient.NetworkingV1().Ingresses(resource.namespace).Create(context.Background(), &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: resource.name, Namespace: resource.namespace},
			}, metav1.CreateOptions{})
			assert.NoError(t, err)
		case "Secret":
			_, err := fakeClient.CoreV1().Secrets(resource.namespace).Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: resource.name, Namespace: resource.namespace},
			}, metav1.CreateOptions{})
			assert.NoError(t, err)
		}
	}

	// Clean up resources
	client.cleanupResources(context.Background(), resources)

	// Verify resources are deleted
	for _, resource := range resources {
		switch resource.kind {
		case "Deployment":
			_, err := fakeClient.AppsV1().Deployments(resource.namespace).Get(context.Background(), resource.name, metav1.GetOptions{})
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		case "Service":
			_, err := fakeClient.CoreV1().Services(resource.namespace).Get(context.Background(), resource.name, metav1.GetOptions{})
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		case "Ingress":
			_, err := fakeClient.NetworkingV1().Ingresses(resource.namespace).Get(context.Background(), resource.name, metav1.GetOptions{})
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		case "Secret":
			_, err := fakeClient.CoreV1().Secrets(resource.namespace).Get(context.Background(), resource.name, metav1.GetOptions{})
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		}
	}
}

func TestGetDeployments(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Create test namespaces with different requesters
	testRequester := "0x1234567890abcdef"
	otherRequester := "0xfedcba0987654321"

	// Create namespaces with requester labels
	namespaces := []*corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "deployment-1",
				Labels: map[string]string{
					"requester":    testRequester,
					"deploymentID": "deployment-1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "deployment-2",
				Labels: map[string]string{
					"requester":    testRequester,
					"deploymentID": "deployment-2",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "deployment-3",
				Labels: map[string]string{
					"requester":    otherRequester,
					"deploymentID": "deployment-3",
				},
			},
		},
	}

	// Create the namespaces
	for _, ns := range namespaces {
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
	}

	// Test getting deployments for the test requester
	deployments, err := client.GetDeployments(context.Background(), testRequester)
	assert.NoError(t, err)
	assert.Len(t, deployments, 2)

	// Verify we got the correct deployments
	deploymentIDs := make(map[string]bool)
	for _, deployment := range deployments {
		deploymentIDs[deployment.ID] = true
	}
	assert.True(t, deploymentIDs["deployment-1"])
	assert.True(t, deploymentIDs["deployment-2"])
	assert.False(t, deploymentIDs["deployment-3"])

	// Test getting deployments for the other requester
	deployments, err = client.GetDeployments(context.Background(), otherRequester)
	assert.NoError(t, err)
	assert.Len(t, deployments, 1)
	assert.Equal(t, "deployment-3", deployments[0].ID)

	// Test getting deployments for a non-existent requester
	deployments, err = client.GetDeployments(context.Background(), "0xnonexistent")
	assert.NoError(t, err)
	assert.Len(t, deployments, 0)
}

func TestWaitForDeployment(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Create a test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
			Labels: map[string]string{
				"requester":    "0x1234567890abcdef",
				"deploymentID": "test-deployment",
			},
		},
	}
	_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Create a deployment that's not ready
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-deployment",
			Labels: map[string]string{
				"deploymentID": "test-deployment",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          2,
			ReadyReplicas:     0, // Not ready
			AvailableReplicas: 0,
			UpdatedReplicas:   0,
		},
	}
	_, err = fakeClient.AppsV1().Deployments("test-deployment").Create(context.Background(), deployment, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Test that waiting times out when deployment is not ready
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = client.WaitForDeployment(ctx, "test-deployment", 500*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to become ready within")

	// Update deployment to be ready
	deployment.Status.ReadyReplicas = 2
	deployment.Status.AvailableReplicas = 2
	deployment.Status.UpdatedReplicas = 2
	_, err = fakeClient.AppsV1().Deployments("test-deployment").Update(context.Background(), deployment, metav1.UpdateOptions{})
	assert.NoError(t, err)

	// Test that waiting succeeds when deployment is ready
	err = client.WaitForDeployment(context.Background(), "test-deployment", 5*time.Second)
	assert.NoError(t, err)
}

// Helper function for int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
