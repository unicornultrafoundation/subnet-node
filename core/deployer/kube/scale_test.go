package kube

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScaleDeployment(t *testing.T) {
	tests := []struct {
		name         string
		deploymentID string
		replicas     int32
		setupFunc    func(*fake.Clientset) error
		wantErr      bool
		validateFunc func(*testing.T, *fake.Clientset)
	}{
		{
			name:         "successful scale up",
			deploymentID: "test-deployment",
			replicas:     3,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployment with 1 replica
				_, err = client.AppsV1().Deployments("test-deployment").Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment-1",
						Namespace: "test-deployment",
						Labels: map[string]string{
							"deploymentID": "test-deployment",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(1),
					},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: false,
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				deployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "test-deployment-1", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(3), *deployment.Spec.Replicas)
			},
		},
		{
			name:         "successful scale down",
			deploymentID: "test-deployment",
			replicas:     0,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployment with 2 replicas
				_, err = client.AppsV1().Deployments("test-deployment").Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment-1",
						Namespace: "test-deployment",
						Labels: map[string]string{
							"deploymentID": "test-deployment",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(2),
					},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: false,
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				deployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "test-deployment-1", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(0), *deployment.Spec.Replicas)
			},
		},
		{
			name:         "scale multiple deployments",
			deploymentID: "test-deployment",
			replicas:     5,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create multiple deployments
				deployments := []*appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "deployment-1",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(1),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "deployment-2",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(2),
						},
					},
				}

				for _, deployment := range deployments {
					_, err := client.AppsV1().Deployments("test-deployment").Create(context.Background(), deployment, metav1.CreateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				deployments, err := client.AppsV1().Deployments("test-deployment").List(context.Background(), metav1.ListOptions{})
				assert.NoError(t, err)
				assert.Len(t, deployments.Items, 2)
				for _, deployment := range deployments.Items {
					assert.Equal(t, int32(5), *deployment.Spec.Replicas)
				}
			},
		},
		{
			name:         "negative replicas should fail",
			deploymentID: "test-deployment",
			replicas:     -1,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: true,
		},
		{
			name:         "no deployments found should fail",
			deploymentID: "empty-deployment",
			replicas:     1,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace but no deployments
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-deployment"},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: true,
		},
		{
			name:         "deployment without deploymentID label should be ignored",
			deploymentID: "test-deployment",
			replicas:     3,
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployment without deploymentID label
				_, err = client.AppsV1().Deployments("test-deployment").Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment-1",
						Namespace: "test-deployment",
						Labels: map[string]string{
							"other-label": "value",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(1),
					},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: true, // Should fail because no deployments with deploymentID label found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fakeClient := fake.NewSimpleClientset()
			logger := logrus.New().WithField("service", "kube").Logger
			client := &KubeClient{
				Client: fakeClient,
				Logger: logger,
			}

			if tt.setupFunc != nil {
				err := tt.setupFunc(fakeClient)
				require.NoError(t, err)
			}

			// Execute
			err := client.ScaleDeployment(context.Background(), tt.deploymentID, tt.replicas)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, fakeClient)
				}
			}
		})
	}
}

func TestScaleDeploymentToOriginal(t *testing.T) {
	tests := []struct {
		name              string
		deploymentRequest *types.DeploymentRequest
		setupFunc         func(*fake.Clientset) error
		wantErr           bool
		validateFunc      func(*testing.T, *fake.Clientset)
	}{
		{
			name: "successful scale to original",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "test-deployment",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"nginx": {Name: "nginx", Image: "nginx:latest", Count: 2},
						"redis": {Name: "redis", Image: "redis:latest", Count: 1},
					},
				},
			},
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployments with different replica counts
				deployments := []*appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "nginx-deployment",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
								"service":      "nginx",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(5), // Scaled up from original 2
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "redis-deployment",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
								"service":      "redis",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(0), // Scaled down from original 1
						},
					},
				}

				for _, deployment := range deployments {
					_, err := client.AppsV1().Deployments("test-deployment").Create(context.Background(), deployment, metav1.CreateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				// Check nginx deployment
				nginxDeployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "nginx-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(2), *nginxDeployment.Spec.Replicas)

				// Check redis deployment
				redisDeployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "redis-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(1), *redisDeployment.Spec.Replicas)
			},
		},
		{
			name:              "nil deployment request should fail",
			deploymentRequest: nil,
			wantErr:           true,
		},
		{
			name: "empty order ID should fail",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"nginx": {Name: "nginx", Image: "nginx:latest", Count: 1},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deployment missing service label should be skipped",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "test-deployment",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"nginx": {Name: "nginx", Image: "nginx:latest", Count: 2},
					},
				},
			},
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployment without service label
				_, err = client.AppsV1().Deployments("test-deployment").Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-deployment",
						Namespace: "test-deployment",
						Labels: map[string]string{
							"deploymentID": "test-deployment",
							// Missing "service" label
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(5),
					},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: false, // Should not fail, just skip the deployment
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				// Deployment should remain unchanged
				deployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "nginx-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(5), *deployment.Spec.Replicas)
			},
		},
		{
			name: "service not found in manifest should be skipped",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "test-deployment",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"nginx": {Name: "nginx", Image: "nginx:latest", Count: 2},
					},
				},
			},
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployment for service not in manifest
				_, err = client.AppsV1().Deployments("test-deployment").Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "redis-deployment",
						Namespace: "test-deployment",
						Labels: map[string]string{
							"deploymentID": "test-deployment",
							"service":      "redis", // Not in manifest
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(3),
					},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: false, // Should not fail, just skip the deployment
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				// Deployment should remain unchanged
				deployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "redis-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(3), *deployment.Spec.Replicas)
			},
		},
		{
			name: "no deployments found should fail",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "empty-deployment",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"nginx": {Name: "nginx", Image: "nginx:latest", Count: 1},
					},
				},
			},
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace but no deployments
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-deployment"},
				}, metav1.CreateOptions{})
				return err
			},
			wantErr: true,
		},
		{
			name: "complex manifest with multiple services",
			deploymentRequest: &types.DeploymentRequest{
				OrderID: "test-deployment",
				Manifest: manifest.SDL{
					VersionStr: "v1",
					Profiles: manifest.Profiles{
						Compute: map[string]manifest.ComputeProfile{
							"default": {
								Resources: manifest.ResourceRequirements{
									CPU:    manifest.ResourceLimit{Request: "100m", Limit: "200m"},
									Memory: manifest.ResourceLimit{Request: "128Mi", Limit: "256Mi"},
								},
							},
						},
					},
					Services: map[string]manifest.Service{
						"web": {Name: "web", Image: "web:latest", Count: 3},
						"api": {Name: "api", Image: "api:latest", Count: 2},
						"db":  {Name: "db", Image: "db:latest", Count: 1},
					},
				},
			},
			setupFunc: func(client *fake.Clientset) error {
				// Create namespace
				_, err := client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
				}, metav1.CreateOptions{})
				if err != nil {
					return err
				}

				// Create deployments with scaled replica counts
				deployments := []*appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "web-deployment",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
								"service":      "web",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(1), // Scaled down from 3
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "api-deployment",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
								"service":      "api",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(5), // Scaled up from 2
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "db-deployment",
							Namespace: "test-deployment",
							Labels: map[string]string{
								"deploymentID": "test-deployment",
								"service":      "db",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: int32Ptr(0), // Scaled down from 1
						},
					},
				}

				for _, deployment := range deployments {
					_, err := client.AppsV1().Deployments("test-deployment").Create(context.Background(), deployment, metav1.CreateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, client *fake.Clientset) {
				// Check web deployment
				webDeployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "web-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(3), *webDeployment.Spec.Replicas)

				// Check api deployment
				apiDeployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "api-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(2), *apiDeployment.Spec.Replicas)

				// Check db deployment
				dbDeployment, err := client.AppsV1().Deployments("test-deployment").Get(context.Background(), "db-deployment", metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, int32(1), *dbDeployment.Spec.Replicas)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fakeClient := fake.NewSimpleClientset()
			logger := logrus.New().WithField("service", "kube").Logger
			client := &KubeClient{
				Client: fakeClient,
				Logger: logger,
			}

			if tt.setupFunc != nil {
				err := tt.setupFunc(fakeClient)
				require.NoError(t, err)
			}

			// Execute
			err := client.ScaleDeploymentToOriginal(context.Background(), tt.deploymentRequest)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, fakeClient)
				}
			}
		})
	}
}
