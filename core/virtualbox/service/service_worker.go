package vbox_service

import (
	"context"
	"encoding/json"
	"fmt"

	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func (s *VBoxService) StartWorker(ctx context.Context) {
	go s.processRequests(ctx)
}

// processRequests processes VM requests from the channel
func (s *VBoxService) processRequests(ctx context.Context) {
	for {
		select {
		case request := <-s.jobManager.GetRequestChannel():
			go s.handleVMRequest(ctx, request)
		case <-s.stopChan:
			serviceLog.Info("Request processing stopped")
			return
		case <-ctx.Done():
			serviceLog.Info("Context cancelled, stopping request processing")
			return
		}
	}
}

func (s *VBoxService) handleVMRequest(ctx context.Context, request *vbtypes.VMRequest) {

	switch request.Type {
	case vbtypes.VMEventCreateVM:
		s.handleCreateVMFromImageRequest(ctx, request)
	default:
		serviceLog.WithField("requestType", request.Type).Error("Unknown request type")
	}
}

func (s *VBoxService) handleCreateVMFromImageRequest(ctx context.Context, request *vbtypes.VMRequest) {

	// change status of job into in progress
	if err := s.jobManager.StartJob(ctx, request.Data["jobID"].(string)); err != nil {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to start job")
		return
	}

	reqData, err := extractVMCreateFromImageRequest(request.Data["request"])
	if err != nil {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract create request from request data")
		if err := s.jobManager.MarkJobAsFailed(ctx, request.Data["jobID"].(string), err.Error()); err != nil {
		}
		return
	}

	s.performCreateVMFromImage(ctx, *reqData)
	// TODO: more implementation

}

func (s *VBoxService) performCreateVMFromImage(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.VM, error) {

	// init storage manager
	// osStorage := storage.NewImageStorage(s.cfg.BaseFolder, s.cfg.ImagesDir)

	// // get or create VDI file
	// vdiPath, err := osStorage.GetOrCreateVDI(ctx, req.Version, runtime.GOARCH, req.Name)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get or create VDI: %w", err)
	// }

	// // generate cloud-init ISO
	// cloudInitISO, err := osStorage.GenerateCloudInitISO(ctx, req.Name, req.Username, req.Password)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	// }
	return nil, nil

}

func extractVMCreateFromImageRequest(data interface{}) (*vbtypes.VMCreateFromImageRequest, error) {
	switch v := data.(type) {
	case vbtypes.VMCreateFromImageRequest:
		return &v, nil
	case *vbtypes.VMCreateFromImageRequest:
		return v, nil
	case map[string]interface{}:
		// Convert map to JSON bytes, then unmarshal to struct
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
		}
		var req vbtypes.VMCreateFromImageRequest
		if err := json.Unmarshal(jsonBytes, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON to VMCreateFromImageRequest: %w", err)
		}
		return &req, nil
	default:
		return nil, fmt.Errorf("unsupported type for VMCreateFromImageRequest: %T", data)
	}
}
