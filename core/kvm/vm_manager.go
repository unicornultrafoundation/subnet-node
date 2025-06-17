package kvm

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VMManagerImpl implements VMManager interface
type VMManagerImpl struct {
	libvirtClient  LibvirtClient
	storageManager StorageManager
	networkManager NetworkManager
	registry       VMRegistry
	logger         *logrus.Logger
}

// NewVMManagerImpl creates a new VM manager instance
func NewVMManagerImpl(libvirtClient LibvirtClient, storageManager StorageManager, networkManager NetworkManager, registry VMRegistry, logger *logrus.Logger) (VMManager, error) {
	return &VMManagerImpl{
		libvirtClient:  libvirtClient,
		storageManager: storageManager,
		networkManager: networkManager,
		registry:       registry,
		logger:         logger,
	}, nil
}

// Provision creates a new VM from specification
func (vm *VMManagerImpl) Provision(ctx context.Context, spec *VMSpec) (*VM, error) {
	vm.logger.WithFields(logrus.Fields{
		"vm_name":  spec.Name,
		"template": spec.Template,
		"cpu":      spec.CPU,
		"memory":   spec.Memory,
		"disk":     spec.Disk,
	}).Info("Provisioning new VM")

	// Generate VM ID
	vmID := uuid.New().String()

	// Create disk from template
	diskPath, err := vm.storageManager.CreateDiskFromTemplate(spec.Template, vmID, uint64(spec.Disk)*1024*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("failed to create disk: %w", err)
	}

	// Setup networking
	networkConfig, err := vm.networkManager.SetupVMNetworking(vmID, spec.NetworkType)
	if err != nil {
		return nil, fmt.Errorf("failed to setup networking: %w", err)
	}

	// Create cloud-init ISO if specified
	var cloudInitPath string
	if spec.CloudInit != nil {
		cloudInitPath, err = vm.storageManager.CreateCloudInitISO(spec.CloudInit, vmID)
		if err != nil {
			vm.logger.WithError(err).Warn("Failed to create cloud-init ISO, proceeding without it")
		}
	}

	// Generate libvirt XML
	domainXML := vm.generateDomainXML(vmID, spec, diskPath, networkConfig, cloudInitPath)

	// Create domain in libvirt
	domain, err := vm.libvirtClient.CreateDomain(domainXML)
	if err != nil {
		// Cleanup on failure
		vm.cleanup(vmID, diskPath, cloudInitPath)
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	// Get domain name
	domainName, err := domain.GetName()
	if err != nil {
		domainName = vmID // Fallback to VM ID
	}

	// Create VM object
	vmObj := &VM{
		ID:         vmID,
		Name:       spec.Name,
		Status:     VMStatusStopped,
		DomainName: domainName,
		Resources: &ResourceInfo{
			CPU:    spec.CPU,
			Memory: spec.Memory,
			Disk:   spec.Disk,
		},
		Network: &NetworkInfo{
			Name:        "default",
			Type:        spec.NetworkType,
			IPAddress:   networkConfig.IPAddress,
			MACAddress:  networkConfig.MACAddress,
			BridgeName:  networkConfig.BridgeName,
			NetworkCIDR: "192.168.100.0/24",
		},
		Storage: &StorageInfo{
			DiskPath: diskPath,
			Size:     uint64(spec.Disk) * 1024 * 1024 * 1024,
			Format:   "qcow2",
		},
		CreatedAt: time.Now(),
		Metadata:  spec.Metadata,
	}

	// Add template to metadata
	if vmObj.Metadata == nil {
		vmObj.Metadata = make(map[string]string)
	}
	vmObj.Metadata["template"] = spec.Template

	// Register VM in registry
	vmMetadata := vm.convertVMToMetadata(vmObj)
	if err := vm.registry.RegisterVM(vmMetadata); err != nil {
		vm.logger.WithError(err).Error("Failed to register VM in registry")
		// Continue without failing since VM is created
	}

	vm.logger.WithField("vm_id", vmID).Info("VM provisioned successfully")

	return vmObj, nil
}

// Start starts a VM
func (vm *VMManagerImpl) Start(ctx context.Context, vmID string) error {
	vm.logger.WithField("vm_id", vmID).Info("Starting VM")

	// Get VM metadata
	metadata, err := vm.registry.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	// Get domain
	domain, err := vm.libvirtClient.GetDomain(metadata.DomainName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Start domain
	if err := domain.Start(); err != nil {
		return fmt.Errorf("failed to start domain: %w", err)
	}

	// Update VM status
	if err := vm.registry.UpdateVM(vmID, map[string]interface{}{
		"status": VMStatusRunning,
	}); err != nil {
		vm.logger.WithError(err).Warn("Failed to update VM status in registry")
	}

	return nil
}

// Stop stops a VM
func (vm *VMManagerImpl) Stop(ctx context.Context, vmID string) error {
	vm.logger.WithField("vm_id", vmID).Info("Stopping VM")

	// Get VM metadata
	metadata, err := vm.registry.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	// Get domain
	domain, err := vm.libvirtClient.GetDomain(metadata.DomainName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Stop domain
	if err := domain.Stop(); err != nil {
		return fmt.Errorf("failed to stop domain: %w", err)
	}

	// Update VM status
	if err := vm.registry.UpdateVM(vmID, map[string]interface{}{
		"status": VMStatusStopped,
	}); err != nil {
		vm.logger.WithError(err).Warn("Failed to update VM status in registry")
	}

	return nil
}

// Restart restarts a VM
func (vm *VMManagerImpl) Restart(ctx context.Context, vmID string) error {
	vm.logger.WithField("vm_id", vmID).Info("Restarting VM")

	// Get VM metadata
	metadata, err := vm.registry.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	// Get domain
	domain, err := vm.libvirtClient.GetDomain(metadata.DomainName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Restart domain
	if err := domain.Restart(); err != nil {
		return fmt.Errorf("failed to restart domain: %w", err)
	}

	// Update VM status
	if err := vm.registry.UpdateVM(vmID, map[string]interface{}{
		"status": VMStatusRunning,
	}); err != nil {
		vm.logger.WithError(err).Warn("Failed to update VM status in registry")
	}

	return nil
}

// Destroy destroys a VM
func (vm *VMManagerImpl) Destroy(ctx context.Context, vmID string) error {
	vm.logger.WithField("vm_id", vmID).Info("Destroying VM")

	// Get VM metadata
	metadata, err := vm.registry.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	// Get domain
	domain, err := vm.libvirtClient.GetDomain(metadata.DomainName)
	if err != nil {
		vm.logger.WithError(err).Warn("Failed to get domain, proceeding with cleanup")
	} else {
		// Destroy domain
		if err := domain.Destroy(); err != nil {
			vm.logger.WithError(err).Warn("Failed to destroy domain")
		}
	}

	// Release IP address
	if err := vm.networkManager.ReleaseIP(vmID); err != nil {
		vm.logger.WithError(err).Warn("Failed to release IP address")
	}

	// Delete disk
	if err := vm.storageManager.DeleteDisk(vmID); err != nil {
		vm.logger.WithError(err).Warn("Failed to delete disk")
	}

	// Unregister from registry
	if err := vm.registry.UnregisterVM(vmID); err != nil {
		vm.logger.WithError(err).Warn("Failed to unregister VM from registry")
	}

	return nil
}

// GetStatus returns the status of a VM
func (vm *VMManagerImpl) GetStatus(ctx context.Context, vmID string) (VMStatus, error) {
	// Get VM metadata
	metadata, err := vm.registry.GetVM(vmID)
	if err != nil {
		return VMStatusUnknown, fmt.Errorf("VM not found: %w", err)
	}

	// Try to get actual status from libvirt
	domain, err := vm.libvirtClient.GetDomain(metadata.DomainName)
	if err != nil {
		vm.logger.WithError(err).Warn("Failed to get domain, returning cached status")
		return metadata.Status, nil
	}

	status, err := domain.GetState()
	if err != nil {
		vm.logger.WithError(err).Warn("Failed to get domain state, returning cached status")
		return metadata.Status, nil
	}

	// Update cached status if different
	if status != metadata.Status {
		if err := vm.registry.UpdateVM(vmID, map[string]interface{}{
			"status": status,
		}); err != nil {
			vm.logger.WithError(err).Warn("Failed to update VM status in registry")
		}
	}

	return status, nil
}

// ListDomains lists all VM domains
func (vm *VMManagerImpl) ListDomains(ctx context.Context) ([]*VM, error) {
	// Get all VMs from registry
	vmMetadataList, err := vm.registry.ListVMs(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs from registry: %w", err)
	}

	var vms []*VM
	for _, metadata := range vmMetadataList {
		vmObj := vm.convertMetadataToVM(metadata)
		vms = append(vms, vmObj)
	}

	return vms, nil
}

// Helper methods

func (vm *VMManagerImpl) generateDomainXML(vmID string, spec *VMSpec, diskPath string, networkConfig *NetworkConfig, cloudInitPath string) string {
	// TODO: Generate proper libvirt XML
	// This is a simplified template for demonstration
	xml := fmt.Sprintf(`
<domain type='kvm'>
  <name>%s</name>
  <memory unit='MiB'>%d</memory>
  <vcpu placement='static'>%d</vcpu>
  <os>
    <type arch='x86_64' machine='pc'>hvm</type>
    <boot dev='hd'/>
  </os>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <mac address='%s'/>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1' autoport='yes'/>
  </devices>
</domain>`, spec.Name, spec.Memory, spec.CPU, diskPath, networkConfig.MACAddress)

	// Add cloud-init disk if available
	if cloudInitPath != "" {
		// TODO: Add cloud-init disk to XML
	}

	return xml
}

func (vm *VMManagerImpl) cleanup(vmID, diskPath, cloudInitPath string) {
	// Best effort cleanup
	if err := vm.storageManager.DeleteDisk(vmID); err != nil {
		vm.logger.WithError(err).Warn("Failed to cleanup disk during error recovery")
	}

	if err := vm.networkManager.ReleaseIP(vmID); err != nil {
		vm.logger.WithError(err).Warn("Failed to release IP during error recovery")
	}
}

func (vm *VMManagerImpl) convertVMToMetadata(vmObj *VM) *VMMetadata {
	return &VMMetadata{
		ID:         vmObj.ID,
		Name:       vmObj.Name,
		Status:     vmObj.Status,
		IPAddress:  vmObj.Network.IPAddress.String(),
		MACAddress: vmObj.Network.MACAddress,
		Template:   vmObj.Metadata["template"],
		Resources: &ResourceAllocation{
			CPU:    big.NewInt(int64(vmObj.Resources.CPU)),
			Memory: big.NewInt(int64(vmObj.Resources.Memory) * 1024 * 1024),      // Convert MB to bytes
			Disk:   big.NewInt(int64(vmObj.Resources.Disk) * 1024 * 1024 * 1024), // Convert GB to bytes
		},
		CreatedAt:   vmObj.CreatedAt,
		UpdatedAt:   time.Now(),
		Metadata:    vmObj.Metadata,
		NetworkType: vmObj.Network.Type,
		DomainName:  vmObj.DomainName,
	}
}

func (vm *VMManagerImpl) convertMetadataToVM(metadata *VMMetadata) *VM {
	return &VM{
		ID:         metadata.ID,
		Name:       metadata.Name,
		Status:     metadata.Status,
		DomainName: metadata.DomainName,
		Resources: &ResourceInfo{
			CPU:    int(metadata.Resources.CPU.Int64()),
			Memory: int(metadata.Resources.Memory.Int64() / (1024 * 1024)),      // Convert bytes to MB
			Disk:   int(metadata.Resources.Disk.Int64() / (1024 * 1024 * 1024)), // Convert bytes to GB
		},
		Network: &NetworkInfo{
			Name:       "default",
			Type:       metadata.NetworkType,
			IPAddress:  net.ParseIP(metadata.IPAddress),
			MACAddress: metadata.MACAddress,
			BridgeName: "virbr0",
		},
		Storage: &StorageInfo{
			DiskPath: "/var/lib/libvirt/images/" + metadata.ID + ".qcow2",
			Size:     metadata.Resources.Disk.Uint64(),
			Format:   "qcow2",
		},
		CreatedAt: metadata.CreatedAt,
		Metadata:  metadata.Metadata,
	}
}
