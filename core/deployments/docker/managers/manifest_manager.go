package managers

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	deployment "github.com/unicornultrafoundation/subnet-node/core/deployments/docker"
	"gopkg.in/yaml.v3"
)

// ManifestManagerImpl implements ManifestManager
type ManifestManagerImpl struct {
	logger *logrus.Logger
}

// NewManifestManager creates a new manifest manager
func NewManifestManager(logger *logrus.Logger) deployment.ManifestManager {
	return &ManifestManagerImpl{
		logger: logger,
	}
}

// ValidateManifest validates a compose manifest
func (mm *ManifestManagerImpl) ValidateManifest(ctx context.Context, manifest *deployment.ComposeManifest) error {
	if manifest.Version == "" {
		manifest.Version = "3.8"
	}

	if len(manifest.Services) == 0 {
		return fmt.Errorf("manifest must contain at least one service")
	}

	// Validate each service
	for serviceName, service := range manifest.Services {
		if err := mm.validateService(serviceName, service); err != nil {
			return fmt.Errorf("invalid service %s: %w", serviceName, err)
		}
	}

	// Validate networks
	for networkName, network := range manifest.Networks {
		if err := mm.validateNetwork(networkName, network); err != nil {
			return fmt.Errorf("invalid network %s: %w", networkName, err)
		}
	}

	return nil
}

// ProcessManifest processes and transforms a manifest for deployment
func (mm *ManifestManagerImpl) ProcessManifest(ctx context.Context, manifest *deployment.ComposeManifest, tenantID string) (*deployment.ComposeManifest, error) {
	// Create a copy of the manifest
	processed := &deployment.ComposeManifest{
		Version:    manifest.Version,
		Services:   make(map[string]deployment.ServiceConfig),
		Networks:   make(map[string]deployment.NetworkConfig),
		Volumes:    manifest.Volumes,
		Secrets:    manifest.Secrets,
		Configs:    manifest.Configs,
		Extensions: manifest.Extensions,
	}

	// Process services
	for serviceName, service := range manifest.Services {
		processedService := service

		// Add tenant labels
		if processedService.Labels == nil {
			processedService.Labels = make(map[string]string)
		}
		processedService.Labels["tenant.id"] = tenantID
		processedService.Labels["tenant.service"] = serviceName

		// Ensure service has a name
		if processedService.Name == "" {
			processedService.Name = serviceName
		}

		// Process networks
		if len(processedService.Networks) == 0 {
			processedService.Networks = []string{fmt.Sprintf("tenant_%s_network", tenantID)}
		}

		processed.Services[serviceName] = processedService
	}

	// Process networks
	for networkName, network := range manifest.Networks {
		processedNetwork := network

		// Add tenant labels
		if processedNetwork.Labels == nil {
			processedNetwork.Labels = make(map[string]string)
		}
		processedNetwork.Labels["tenant.id"] = tenantID

		// Ensure network has a name
		if processedNetwork.Name == "" {
			processedNetwork.Name = networkName
		}

		processed.Networks[networkName] = processedNetwork
	}

	// Ensure tenant network exists
	tenantNetworkName := fmt.Sprintf("tenant_%s_network", tenantID)
	if _, exists := processed.Networks[tenantNetworkName]; !exists {
		processed.Networks[tenantNetworkName] = deployment.NetworkConfig{
			Name:       tenantNetworkName,
			Driver:     "bridge",
			Labels:     map[string]string{"tenant.id": tenantID},
			Internal:   false,
			EnableIPv6: false,
		}
	}

	return processed, nil
}

// GenerateManifest generates a manifest from a template
func (mm *ManifestManagerImpl) GenerateManifest(ctx context.Context, templateStr string, params map[string]interface{}) (*deployment.ComposeManifest, error) {
	// Parse template
	tmpl, err := template.New("manifest").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Parse the generated YAML
	var manifest deployment.ComposeManifest
	if err := yaml.Unmarshal([]byte(result.String()), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse generated manifest: %w", err)
	}

	return &manifest, nil
}

// MergeManifests merges multiple manifests
func (mm *ManifestManagerImpl) MergeManifests(ctx context.Context, manifests ...*deployment.ComposeManifest) (*deployment.ComposeManifest, error) {
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no manifests to merge")
	}

	merged := &deployment.ComposeManifest{
		Version:    "3.8",
		Services:   make(map[string]deployment.ServiceConfig),
		Networks:   make(map[string]deployment.NetworkConfig),
		Volumes:    make(map[string]interface{}),
		Secrets:    make(map[string]interface{}),
		Configs:    make(map[string]interface{}),
		Extensions: make(map[string]interface{}),
	}

	for i, manifest := range manifests {
		// Merge services
		for serviceName, service := range manifest.Services {
			// Add prefix to avoid conflicts
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, serviceName)
			merged.Services[prefixedName] = service
		}

		// Merge networks
		for networkName, network := range manifest.Networks {
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, networkName)
			merged.Networks[prefixedName] = network
		}

		// Merge volumes
		for volumeName, volume := range manifest.Volumes {
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, volumeName)
			merged.Volumes[prefixedName] = volume
		}

		// Merge secrets
		for secretName, secret := range manifest.Secrets {
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, secretName)
			merged.Secrets[prefixedName] = secret
		}

		// Merge configs
		for configName, config := range manifest.Configs {
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, configName)
			merged.Configs[prefixedName] = config
		}

		// Merge extensions
		for extName, ext := range manifest.Extensions {
			prefixedName := fmt.Sprintf("manifest_%d_%s", i, extName)
			merged.Extensions[prefixedName] = ext
		}
	}

	return merged, nil
}

// SerializeManifest serializes a manifest to YAML
func (mm *ManifestManagerImpl) SerializeManifest(ctx context.Context, manifest *deployment.ComposeManifest) ([]byte, error) {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %w", err)
	}
	return data, nil
}

// DeserializeManifest deserializes a manifest from YAML
func (mm *ManifestManagerImpl) DeserializeManifest(ctx context.Context, data []byte) (*deployment.ComposeManifest, error) {
	var manifest deployment.ComposeManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to deserialize manifest: %w", err)
	}
	return &manifest, nil
}

// Helper methods

func (mm *ManifestManagerImpl) validateService(serviceName string, service deployment.ServiceConfig) error {
	if service.Image == "" {
		return fmt.Errorf("service must have an image")
	}

	// Validate ports
	for _, port := range service.Ports {
		if port.ContainerPort <= 0 || port.ContainerPort > 65535 {
			return fmt.Errorf("invalid container port: %d", port.ContainerPort)
		}
		if port.HostPort < 0 || port.HostPort > 65535 {
			return fmt.Errorf("invalid host port: %d", port.HostPort)
		}
		if port.Protocol != "" && port.Protocol != "tcp" && port.Protocol != "udp" {
			return fmt.Errorf("invalid protocol: %s", port.Protocol)
		}
	}

	// Validate restart policy
	if service.RestartPolicy != "" {
		validPolicies := []string{"no", "always", "on-failure", "unless-stopped"}
		valid := false
		for _, policy := range validPolicies {
			if service.RestartPolicy == policy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid restart policy: %s", service.RestartPolicy)
		}
	}

	return nil
}

func (mm *ManifestManagerImpl) validateNetwork(networkName string, network deployment.NetworkConfig) error {
	if network.Driver == "" {
		network.Driver = "bridge"
	}

	validDrivers := []string{"bridge", "host", "overlay", "macvlan", "none"}
	valid := false
	for _, driver := range validDrivers {
		if network.Driver == driver {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid network driver: %s", network.Driver)
	}

	return nil
}
