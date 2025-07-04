package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// Service represents the storage service for VirtualBox VMs
type Service struct {
	ds     datastore.Datastore
	logger *logrus.Logger
}

// NewService creates a new storage service
func NewService(ds datastore.Datastore, logger *logrus.Logger) *Service {
	return &Service{
		ds:     ds,
		logger: logger,
	}
}

// Start starts the storage service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting VirtualBox storage service")
	return nil
}

// Stop stops the storage service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping VirtualBox storage service")
	return nil
}

// StoreVMRequest stores a VM request
func (s *Service) StoreVMRequest(ctx context.Context, request *types.VMRequest) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-requests/%s", request.ID))
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal VM request: %w", err)
	}

	if err := s.ds.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store VM request: %w", err)
	}

	s.logger.WithField("vmID", request.ID).Info("Stored VM request")
	return nil
}

// GetVMRequest retrieves a VM request by ID
func (s *Service) GetVMRequest(ctx context.Context, vmID string) (*types.VMRequest, error) {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-requests/%s", vmID))
	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("VM request not found: %s", vmID)
		}
		return nil, fmt.Errorf("failed to get VM request: %w", err)
	}

	var request types.VMRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM request: %w", err)
	}

	return &request, nil
}

// GetVMRequests retrieves all VM requests for a requester
func (s *Service) GetVMRequests(ctx context.Context, requester string) ([]*types.VMRequest, error) {
	prefix := datastore.NewKey("/virtualbox/vm-requests/")
	q := query.Query{
		Prefix: prefix.String(),
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query VM requests: %w", err)
	}
	defer results.Close()

	var requests []*types.VMRequest
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Error reading VM request")
			continue
		}

		var request types.VMRequest
		if err := json.Unmarshal(result.Value, &request); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal VM request")
			continue
		}

		if requester == "" || request.Requester == requester {
			requests = append(requests, &request)
		}
	}

	return requests, nil
}

// DeleteVMRequest deletes a VM request
func (s *Service) DeleteVMRequest(ctx context.Context, vmID string) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-requests/%s", vmID))
	if err := s.ds.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete VM request: %w", err)
	}

	s.logger.WithField("vmID", vmID).Info("Deleted VM request")
	return nil
}

// StoreVMInfo stores VM information
func (s *Service) StoreVMInfo(ctx context.Context, vmInfo *types.VMInfo) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-info/%s", vmInfo.ID))
	data, err := json.Marshal(vmInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal VM info: %w", err)
	}

	if err := s.ds.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store VM info: %w", err)
	}

	s.logger.WithField("vmID", vmInfo.ID).Debug("Stored VM info")
	return nil
}

// GetVMInfo retrieves VM information by ID
func (s *Service) GetVMInfo(ctx context.Context, vmID string) (*types.VMInfo, error) {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-info/%s", vmID))
	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("VM info not found: %s", vmID)
		}
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	var vmInfo types.VMInfo
	if err := json.Unmarshal(data, &vmInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM info: %w", err)
	}

	return &vmInfo, nil
}

// GetVMInfos retrieves all VM information for a requester
func (s *Service) GetVMInfos(ctx context.Context, requester string) ([]*types.VMInfo, error) {
	prefix := datastore.NewKey("/virtualbox/vm-info/")
	q := query.Query{
		Prefix: prefix.String(),
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query VM infos: %w", err)
	}
	defer results.Close()

	var vmInfos []*types.VMInfo
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Error reading VM info")
			continue
		}

		var vmInfo types.VMInfo
		if err := json.Unmarshal(result.Value, &vmInfo); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal VM info")
			continue
		}

		if requester == "" || vmInfo.Metadata["requester"] == requester {
			vmInfos = append(vmInfos, &vmInfo)
		}
	}

	return vmInfos, nil
}

// DeleteVMInfo deletes VM information
func (s *Service) DeleteVMInfo(ctx context.Context, vmID string) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-info/%s", vmID))
	if err := s.ds.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete VM info: %w", err)
	}

	s.logger.WithField("vmID", vmID).Info("Deleted VM info")
	return nil
}

// StoreVMOperation stores a VM operation
func (s *Service) StoreVMOperation(ctx context.Context, operation *types.VMOperation) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-operations/%s", operation.ID))
	data, err := json.Marshal(operation)
	if err != nil {
		return fmt.Errorf("failed to marshal VM operation: %w", err)
	}

	if err := s.ds.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store VM operation: %w", err)
	}

	s.logger.WithField("operationID", operation.ID).Debug("Stored VM operation")
	return nil
}

// GetVMOperation retrieves a VM operation by ID
func (s *Service) GetVMOperation(ctx context.Context, operationID string) (*types.VMOperation, error) {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-operations/%s", operationID))
	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("VM operation not found: %s", operationID)
		}
		return nil, fmt.Errorf("failed to get VM operation: %w", err)
	}

	var operation types.VMOperation
	if err := json.Unmarshal(data, &operation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM operation: %w", err)
	}

	return &operation, nil
}

// StoreVMOperationResult stores a VM operation result
func (s *Service) StoreVMOperationResult(ctx context.Context, result *types.VMOperationResult) error {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-operation-results/%s", result.ID))
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal VM operation result: %w", err)
	}

	if err := s.ds.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store VM operation result: %w", err)
	}

	s.logger.WithField("operationID", result.ID).Debug("Stored VM operation result")
	return nil
}

// GetVMOperationResult retrieves a VM operation result by ID
func (s *Service) GetVMOperationResult(ctx context.Context, operationID string) (*types.VMOperationResult, error) {
	key := datastore.NewKey(fmt.Sprintf("/virtualbox/vm-operation-results/%s", operationID))
	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("VM operation result not found: %s", operationID)
		}
		return nil, fmt.Errorf("failed to get VM operation result: %w", err)
	}

	var result types.VMOperationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM operation result: %w", err)
	}

	return &result, nil
}

// CleanupExpiredVMs removes expired VM data
func (s *Service) CleanupExpiredVMs(ctx context.Context) error {
	now := time.Now()

	// Clean up expired VM requests
	requests, err := s.GetVMRequests(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get VM requests for cleanup: %w", err)
	}

	for _, request := range requests {
		expiresAt := request.CreatedAt.Add(request.TTL)
		if now.After(expiresAt) {
			s.logger.WithField("vmID", request.ID).Info("Cleaning up expired VM request")
			if err := s.DeleteVMRequest(ctx, request.ID); err != nil {
				s.logger.WithError(err).WithField("vmID", request.ID).Error("Failed to delete expired VM request")
			}
			if err := s.DeleteVMInfo(ctx, request.ID); err != nil {
				s.logger.WithError(err).WithField("vmID", request.ID).Error("Failed to delete expired VM info")
			}
		}
	}

	return nil
}
