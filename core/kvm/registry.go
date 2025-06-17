package kvm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
)

const (
	vmRegistryPrefix = "/kvm/vms/"
)

// VMRegistryImpl implements VMRegistry interface
type VMRegistryImpl struct {
	store  datastore.Datastore
	logger *logrus.Logger
}

// NewVMRegistryImpl creates a new VM registry instance
func NewVMRegistryImpl(store datastore.Datastore, logger *logrus.Logger) (VMRegistry, error) {
	return &VMRegistryImpl{
		store:  store,
		logger: logger,
	}, nil
}

// RegisterVM stores VM metadata in the registry
func (vr *VMRegistryImpl) RegisterVM(vm *VMMetadata) error {
	key := datastore.NewKey(vmRegistryPrefix + vm.ID)

	// Set timestamps
	now := time.Now()
	if vm.CreatedAt.IsZero() {
		vm.CreatedAt = now
	}
	vm.UpdatedAt = now

	// Serialize VM metadata
	data, err := json.Marshal(vm)
	if err != nil {
		return fmt.Errorf("failed to marshal VM metadata: %w", err)
	}

	// Store in datastore
	ctx := context.Background()
	if err := vr.store.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store VM metadata: %w", err)
	}

	vr.logger.WithFields(logrus.Fields{
		"vm_id":   vm.ID,
		"vm_name": vm.Name,
	}).Debug("VM registered in registry")

	return nil
}

// UpdateVM updates VM metadata in the registry
func (vr *VMRegistryImpl) UpdateVM(vmID string, updates map[string]interface{}) error {
	key := datastore.NewKey(vmRegistryPrefix + vmID)
	ctx := context.Background()

	// Get existing VM metadata
	data, err := vr.store.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return fmt.Errorf("VM not found: %s", vmID)
		}
		return fmt.Errorf("failed to get VM metadata: %w", err)
	}

	// Deserialize existing metadata
	var vm VMMetadata
	if err := json.Unmarshal(data, &vm); err != nil {
		return fmt.Errorf("failed to unmarshal VM metadata: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(VMStatus); ok {
				vm.Status = status
			} else if statusStr, ok := value.(string); ok {
				vm.Status = VMStatus(statusStr)
			}
		case "ip_address":
			if ip, ok := value.(string); ok {
				vm.IPAddress = ip
			}
		case "mac_address":
			if mac, ok := value.(string); ok {
				vm.MACAddress = mac
			}
		case "metadata":
			if metadata, ok := value.(map[string]string); ok {
				if vm.Metadata == nil {
					vm.Metadata = make(map[string]string)
				}
				for k, v := range metadata {
					vm.Metadata[k] = v
				}
			}
		}
	}

	// Update timestamp
	vm.UpdatedAt = time.Now()

	// Serialize updated metadata
	updatedData, err := json.Marshal(&vm)
	if err != nil {
		return fmt.Errorf("failed to marshal updated VM metadata: %w", err)
	}

	// Store updated metadata
	if err := vr.store.Put(ctx, key, updatedData); err != nil {
		return fmt.Errorf("failed to update VM metadata: %w", err)
	}

	vr.logger.WithFields(logrus.Fields{
		"vm_id":   vmID,
		"updates": updates,
	}).Debug("VM metadata updated")

	return nil
}

// UnregisterVM removes VM metadata from the registry
func (vr *VMRegistryImpl) UnregisterVM(vmID string) error {
	key := datastore.NewKey(vmRegistryPrefix + vmID)
	ctx := context.Background()

	// Check if VM exists
	exists, err := vr.store.Has(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check VM existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	// Delete VM metadata
	if err := vr.store.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete VM metadata: %w", err)
	}

	vr.logger.WithField("vm_id", vmID).Debug("VM unregistered from registry")

	return nil
}

// GetVM retrieves VM metadata from the registry
func (vr *VMRegistryImpl) GetVM(vmID string) (*VMMetadata, error) {
	key := datastore.NewKey(vmRegistryPrefix + vmID)
	ctx := context.Background()

	// Get VM metadata
	data, err := vr.store.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("VM not found: %s", vmID)
		}
		return nil, fmt.Errorf("failed to get VM metadata: %w", err)
	}

	// Deserialize metadata
	var vm VMMetadata
	if err := json.Unmarshal(data, &vm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM metadata: %w", err)
	}

	return &vm, nil
}

// ListVMs lists all VMs with optional filters
func (vr *VMRegistryImpl) ListVMs(filters map[string]string) ([]*VMMetadata, error) {
	ctx := context.Background()

	// Query all VM keys
	q := query.Query{
		Prefix: vmRegistryPrefix,
	}

	results, err := vr.store.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query VMs: %w", err)
	}
	defer results.Close()

	var vms []*VMMetadata

	// Process results
	for result := range results.Next() {
		if result.Error != nil {
			vr.logger.WithError(result.Error).Warn("Error processing VM registry entry")
			continue
		}

		// Deserialize VM metadata
		var vm VMMetadata
		if err := json.Unmarshal(result.Value, &vm); err != nil {
			vr.logger.WithError(err).WithField("key", result.Key).Warn("Failed to unmarshal VM metadata")
			continue
		}

		// Apply filters
		if vr.matchesFilters(&vm, filters) {
			vms = append(vms, &vm)
		}
	}

	return vms, nil
}

// VMExists checks if a VM exists in the registry
func (vr *VMRegistryImpl) VMExists(vmID string) bool {
	key := datastore.NewKey(vmRegistryPrefix + vmID)
	ctx := context.Background()

	exists, err := vr.store.Has(ctx, key)
	if err != nil {
		vr.logger.WithError(err).WithField("vm_id", vmID).Warn("Failed to check VM existence")
		return false
	}

	return exists
}

// matchesFilters checks if VM metadata matches the provided filters
func (vr *VMRegistryImpl) matchesFilters(vm *VMMetadata, filters map[string]string) bool {
	if filters == nil || len(filters) == 0 {
		return true
	}

	for key, value := range filters {
		switch key {
		case "status":
			if string(vm.Status) != value {
				return false
			}
		case "name":
			if !strings.Contains(strings.ToLower(vm.Name), strings.ToLower(value)) {
				return false
			}
		case "template":
			if vm.Template != value {
				return false
			}
		case "network_type":
			if string(vm.NetworkType) != value {
				return false
			}
		case "ip_address":
			if vm.IPAddress != value {
				return false
			}
		default:
			// Check in metadata
			if vm.Metadata != nil {
				if metaValue, exists := vm.Metadata[key]; !exists || metaValue != value {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}
