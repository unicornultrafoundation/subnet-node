package builder

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/crd"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"go.uber.org/zap"
)

var (
	manifestGVR = schema.GroupVersionResource{
		Group:    "provider.local",
		Version:  "v1",
		Resource: "manifests",
	}
)

// Manifest represents a Kubernetes manifest builder
type Manifest struct {
	settings   Settings
	client     dynamic.Interface
	deployment *types.ManagedDeployment
	ctx        context.Context
	ns         string
}

// NewManifest creates a new manifest builder
func NewManifest(settings Settings, client dynamic.Interface, deployment *types.ManagedDeployment, ctx context.Context, ns string) *Manifest {
	return &Manifest{
		settings:   settings,
		client:     client,
		deployment: deployment,
		ctx:        ctx,
		ns:         ns,
	}
}

// BuildManifest creates a new manifest builder
func BuildManifest(logger *zap.Logger, settings Settings, ns string, deployment *types.ManagedDeployment) *Manifest {
	return NewManifest(
		settings,
		settings.DynamicClient,
		deployment,
		context.Background(),
		ns,
	)
}

// Create creates a new manifest
func (b *Manifest) Create() error {
	if b.deployment == nil {
		return fmt.Errorf("deployment is required")
	}

	if b.deployment.Manifest == nil {
		return fmt.Errorf("deployment manifest is required")
	}

	if len(b.deployment.Manifest.Groups) == 0 {
		return fmt.Errorf("deployment manifest must have at least one group")
	}

	// Convert manifest groups to GroupSpec
	groups := make([]map[string]interface{}, len(b.deployment.Manifest.Groups))
	for groupIdx, group := range b.deployment.Manifest.Groups {
		services := make([]map[string]interface{}, len(group.Services))
		for serviceIdx, svc := range group.Services {
			service := map[string]interface{}{
				"name":    fmt.Sprintf("service-%d", serviceIdx),
				"image":   svc.Image,
				"command": svc.Command,
				"args":    svc.Args,
				"expose":  make([]map[string]interface{}, len(svc.Expose)),
			}

			// Add resources if present
			if svc.Resources != nil {
				resources := map[string]interface{}{}
				if svc.Resources.CPU != nil {
					resources["cpu"] = map[string]interface{}{
						"units": svc.Resources.CPU.Units.Value,
					}
				}
				if svc.Resources.Memory != nil {
					resources["memory"] = map[string]interface{}{
						"size": fmt.Sprintf("%d%s", svc.Resources.Memory.Size.Value, svc.Resources.Memory.Size.Unit),
					}
				}
				if len(resources) > 0 {
					service["resources"] = resources
				}
			}

			// Add expose ports
			for j, port := range svc.Expose {
				expose := map[string]interface{}{
					"port":  port.Port,
					"proto": port.Proto,
				}

				if port.HTTPOptions != nil {
					expose["http_options"] = map[string]interface{}{
						"max_body_size": port.HTTPOptions.MaxBodySize,
						"next_cases":    port.HTTPOptions.NextCases,
					}
				}

				if len(port.To) > 0 {
					expose["global"] = port.To[0].Global
				}

				service["expose"].([]map[string]interface{})[j] = expose
			}

			services[serviceIdx] = service
		}

		groups[groupIdx] = map[string]interface{}{
			"name":     group.Name,
			"services": services,
		}
	}

	manifest := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "provider.local/v1",
			"kind":       "Manifest",
			"metadata": map[string]interface{}{
				"name":      b.Name(),
				"namespace": b.NS(),
				"labels":    b.labels(),
			},
			"spec": map[string]interface{}{
				"groups": groups,
				"lease_id": map[string]interface{}{
					"dseq":     "1",
					"gseq":     1,
					"oseq":     1,
					"owner":    b.deployment.Requester.Hex(),
					"provider": "subnet-node1",
				},
			},
		},
	}

	// Set the GVK
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "provider.local",
		Version: "v1",
		Kind:    "Manifest",
	})

	// Check if manifest already exists
	existing, err := b.client.Resource(manifestGVR).Namespace(b.NS()).Get(b.ctx, b.Name(), metav1.GetOptions{})
	if err == nil {
		// Manifest exists, update it
		manifest.SetResourceVersion(existing.GetResourceVersion())
		_, err = b.client.Resource(manifestGVR).Namespace(b.NS()).Update(b.ctx, manifest, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update manifest: %w", err)
		}
		if b.settings.Logger != nil {
			b.settings.Logger.Info("Successfully updated existing manifest",
				zap.String("name", b.Name()),
				zap.String("namespace", b.NS()))
		}
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if manifest exists: %w", err)
	}

	// Create new manifest
	_, err = b.client.Resource(manifestGVR).Namespace(b.NS()).Create(b.ctx, manifest, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Manifest already exists, skip creation
			if b.settings.Logger != nil {
				b.settings.Logger.Info("Manifest already exists, skipping creation",
					zap.String("name", b.Name()),
					zap.String("namespace", b.NS()))
			}
			return nil
		}
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	if b.settings.Logger != nil {
		b.settings.Logger.Info("Successfully created manifest",
			zap.String("name", b.Name()),
			zap.String("namespace", b.NS()))
	}
	return nil
}

// Update updates an existing manifest
func (b *Manifest) Update() error {
	if b.deployment == nil {
		return fmt.Errorf("deployment is required")
	}

	// Convert manifest services to ServiceSpec
	services := make([]crd.ServiceSpec, len(b.deployment.Manifest.Groups[0].Services))
	for i, svc := range b.deployment.Manifest.Groups[0].Services {
		services[i] = crd.ServiceSpec{
			Name:  fmt.Sprintf("service-%d", i),
			Image: svc.Image,
			Ports: make([]crd.PortSpec, len(svc.Expose)),
			Resources: crd.ResourceSpec{
				CPU: crd.ResourceValue{
					Units: "1",
				},
				Memory: crd.ResourceValue{
					Size: "1Gi",
				},
			},
		}
		for j, port := range svc.Expose {
			services[i].Ports[j] = crd.PortSpec{
				Name:     fmt.Sprintf("port-%d", port.Port),
				Port:     port.Port,
				Protocol: port.Proto,
			}
		}
	}

	manifest := &crd.K8sManifest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Manifest",
			APIVersion: "provider.local/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: crd.ManifestSpec{
			Services: services,
		},
	}

	// Marshal to JSON and unmarshal into Unstructured
	{
		b, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest: %w", err)
		}
		var u unstructured.Unstructured
		if err := json.Unmarshal(b, &u.Object); err != nil {
			return fmt.Errorf("failed to unmarshal manifest to unstructured: %w", err)
		}
		manifest.Unstructured = u
	}

	_, err := b.client.Resource(manifestGVR).Namespace(b.NS()).Update(b.ctx, &manifest.Unstructured, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// Name returns the name of the manifest
func (b *Manifest) Name() string {
	return fmt.Sprintf("manifest-%s", b.deployment.ID)
}

// NS returns the namespace of the manifest
func (b *Manifest) NS() string {
	return b.ns
}

// labels returns the labels for the manifest
func (b *Manifest) labels() map[string]string {
	return map[string]string{
		"app":                  b.Name(),
		"subnet-node/lease-id": b.deployment.ID,
		"subnet-node/owner":    b.deployment.Requester.Hex(),
	}
}
