package crd

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	manifestGVR = schema.GroupVersionResource{
		Group:    "subnet-node",
		Version:  "v2beta2",
		Resource: "manifests",
	}
)

// Client is a client for the Manifest CRD
type Client struct {
	DynamicClient dynamic.Interface
}

// NewClient creates a new client for the Manifest CRD
func NewClient(client dynamic.Interface) *Client {
	return &Client{
		DynamicClient: client,
	}
}

// Get gets a manifest
func (c *Client) Get(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*K8sManifest, error) {
	obj, err := c.DynamicClient.Resource(manifestGVR).Namespace(namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	return &K8sManifest{Unstructured: *obj}, nil
}

// Create creates a manifest
func (c *Client) Create(ctx context.Context, manifest *K8sManifest) (*K8sManifest, error) {
	namespace := manifest.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("manifest namespace is required")
	}
	obj, err := c.DynamicClient.Resource(manifestGVR).Namespace(namespace).Create(ctx, &manifest.Unstructured, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}
	return &K8sManifest{Unstructured: *obj}, nil
}

// Update updates a manifest
func (c *Client) Update(ctx context.Context, manifest *K8sManifest) (*K8sManifest, error) {
	namespace := manifest.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("manifest namespace is required")
	}
	obj, err := c.DynamicClient.Resource(manifestGVR).Namespace(namespace).Update(ctx, &manifest.Unstructured, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update manifest: %w", err)
	}
	return &K8sManifest{Unstructured: *obj}, nil
}

// Delete deletes a manifest
func (c *Client) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	err := c.DynamicClient.Resource(manifestGVR).Namespace(namespace).Delete(ctx, name, opts)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	return nil
}

// List lists manifests
func (c *Client) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*K8sManifestList, error) {
	obj, err := c.DynamicClient.Resource(manifestGVR).Namespace(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list manifests: %w", err)
	}
	return &K8sManifestList{UnstructuredList: *obj}, nil
}
