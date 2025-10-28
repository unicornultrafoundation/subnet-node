package vbox_service

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/cmd_exec"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/config"
	job_manager "github.com/unicornultrafoundation/subnet-node/core/virtualbox/job-manager"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var serviceLog = logrus.WithField("service", "vbox_service")

var (
	reVMNameUUID      = regexp.MustCompile(`"(.+)" {([0-9a-f-]+)}`)
	reVMInfoLine      = regexp.MustCompile(`(?:"(.+)"|(.+))=(?:"(.*)"|(.*))`)
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reMachineNotFound = regexp.MustCompile(`Could not find a registered machine named '(.+)'`)
)

var vBoxCmd cmd_exec.Command

type VBoxService struct {
	cfg *config.VBoxConfig
	ds  datastore.Datastore

	jobManager *job_manager.JobManager

	lock sync.Mutex

	syncTicker *time.Ticker
	stopChan   chan struct{}

	orderToVMMap map[string]string
}

func NewVboxService(cfg *config.VBoxConfig, ds datastore.Datastore) *VBoxService {

	vBoxCmd = cmd_exec.GetVBoxCmd()

	return &VBoxService{
		cfg:        cfg,
		ds:         ds,
		jobManager: job_manager.NewJobManager(),
	}
}

func (s *VBoxService) Start(ctx context.Context) error {

	if err := s.loadAllOrderMappingsFromDatastore(ctx); err != nil {
		serviceLog.Warnf("Failed to load order mappings from datastore: %v", err)
		return err
	}

	// Sync VM in Vbox into datastore every 30 seconds
	s.syncTicker = time.NewTicker(30 * time.Second)
	go s.syncVMToDatastore(ctx)
	return nil
}

// VBOX Functions

// Create VM
// check resource availability
// create VM
// store VM in datastore
func (s *VBoxService) CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error) {

	// Create a job for tracking progress
	requestData := map[string]interface{}{
		"name":         req.Name,
		"cpu_cores":    req.CPUCores,
		"memory_mb":    req.MemoryMB,
		"disk_size_gb": req.DiskSizeGB,
		"os":           req.OS,
		"version":      req.Version,
		"username":     req.Username,
		"password":     req.Password,
		"order_id":     req.OrderId,
	}

	job, err := s.jobManager.CreateJob(ctx, vbtypes.VMEventCreateVM, requestData, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return &vbtypes.JobCreateResponse{
		JobID: job.ID,
	}, nil
}
