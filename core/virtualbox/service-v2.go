package virtualbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	job_manager "github.com/unicornultrafoundation/subnet-node/core/virtualbox/job-manager"
	vbox_service "github.com/unicornultrafoundation/subnet-node/core/virtualbox/service"
	storage_v2 "github.com/unicornultrafoundation/subnet-node/core/virtualbox/storage_v2"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

type IVirtualboxService interface {
	Start(ctx context.Context) error
	CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error)
	GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	GetVMs(ctx context.Context) ([]*vbtypes.VM, error)
	GetVMByOrderId(ctx context.Context, orderId string) (*vbtypes.VM, error)
	GetVMIdByOrderId(orderId string) (string, bool)
	DeleteVM(ctx context.Context, vmID string) error
}

type virtualboxService struct {
	vBoxService    *vbox_service.VBoxService
	jobManager     *job_manager.JobManager
	storageManager *storage_v2.ImageStorage
	ds             datastore.Datastore

	orderToVMMap map[string]string
	orderMapMu   sync.RWMutex

	syncTicker *time.Ticker
	stopChan   chan struct{}
}

func NewServiceV2(ds datastore.Datastore, cfg VirtualBoxConfig) IVirtualboxService {
	storageManager := storage_v2.NewImageStorage(cfg.BaseFolder, cfg.ImagesDir)
	vboxService := vbox_service.NewVBoxService()

	orderToVMMap := make(map[string]string)
	jobManager := job_manager.NewJobManager(storageManager, vboxService, &orderToVMMap, ds)
	return &virtualboxService{
		vBoxService:    vboxService,
		jobManager:     jobManager,
		storageManager: storageManager,
		ds:             ds,
		orderToVMMap:   orderToVMMap,
		orderMapMu:     sync.RWMutex{},
		stopChan:       make(chan struct{}),
	}
}

func (s *virtualboxService) Start(ctx context.Context) error {
	if err := s.loadAllOrderMappingsFromDatastore(ctx); err != nil {
		serviceLog.Warnf("Failed to load order mappings from datastore: %v", err)
	}

	return nil
}

// CreateVM will last long time to create a VM,
// Add the request into the job manager and return the job id
// The job manager will handle the request and create the VM in the background
func (s *virtualboxService) CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error) {
	fmt.Println("createVM new flow")

	job, err := s.jobManager.CreateJob(ctx, vbtypes.VMEventCreateVM, req, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return &vbtypes.JobCreateResponse{
		JobID: job.ID,
	}, nil
}

func (s *virtualboxService) GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {

	fmt.Println("getVM new flow")

	vm, err := s.vBoxService.GetVM(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	return vm, nil
}

func (s *virtualboxService) GetVMs(ctx context.Context) ([]*vbtypes.VM, error) {
	vms, err := s.vBoxService.ListVMs()
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}
	return vms, nil
}

func (s *virtualboxService) loadAllOrderMappingsFromDatastore(ctx context.Context) error {
	q := query.Query{Prefix: "virtualbox/order_mapping/"}
	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return err
	}
	defer results.Close()

	for result := range results.Next() {
		if result.Error != nil {
			continue
		}
		var mapping map[string]string
		if err := json.Unmarshal(result.Value, &mapping); err != nil {
			continue
		}
		orderId := mapping["orderId"]
		vmId := mapping["vmId"]
		if orderId != "" && vmId != "" {
			s.orderToVMMap[orderId] = vmId
		}
	}

	serviceLog.Infof("Loaded %d order mappings from datastore", len(s.orderToVMMap))
	return nil
}

func (s *virtualboxService) GetVMByOrderId(ctx context.Context, orderId string) (*vbtypes.VM, error) {

	vmId, exists := s.orderToVMMap[orderId]
	if !exists {
		return nil, fmt.Errorf("VM not found for this orderId")
	}
	return s.GetVM(ctx, vmId)
}

func (s *virtualboxService) GetVMIdByOrderId(orderId string) (string, bool) {
	s.orderMapMu.RLock()
	defer s.orderMapMu.RUnlock()
	vmId, exists := s.orderToVMMap[orderId]
	return vmId, exists
}

func (s *virtualboxService) DeleteVM(ctx context.Context, vmID string) error {
	err := s.vBoxService.DeleteVM(vmID)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}
	return nil
}
