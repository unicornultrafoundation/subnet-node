package job_manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"

	vbox_service "github.com/unicornultrafoundation/subnet-node/core/virtualbox/service"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/storage"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

type JobManager struct {
	mu             sync.RWMutex
	requestChannel chan *vbtypes.VMRequest

	jobs map[string]*vbtypes.Job

	orderToVMMap *map[string]string
	orderMapMu   sync.RWMutex

	datastore datastore.Datastore

	storageManager *storage.ImageStorage
	vboxService    *vbox_service.VBoxService

	stopChan chan struct{}
}

// jobManager with 100 request channel buffer
func NewJobManager(storageManager *storage.ImageStorage, vboxService *vbox_service.VBoxService, orderToVMMap *map[string]string, datastore datastore.Datastore) *JobManager {

	jobManager := JobManager{
		requestChannel: make(chan *vbtypes.VMRequest, 100),
		jobs:           make(map[string]*vbtypes.Job),
		orderToVMMap:   orderToVMMap,
		orderMapMu:     sync.RWMutex{},
		datastore:      datastore,
		stopChan:       make(chan struct{}),
		storageManager: storageManager,
		vboxService:    vboxService,
	}
	jobManager.Start(context.Background())
	return &jobManager
}

func (jm *JobManager) Start(ctx context.Context) error {
	go jm.startJobCleanup(ctx)
	go jm.startWorker(ctx)
	return nil
}

func (jm *JobManager) CreateJob(ctx context.Context, jobType vbtypes.VMEventType, request interface{}, vmName string) (*vbtypes.Job, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jobID := generateJobID(jobType, vmName)

	job := &vbtypes.Job{
		ID:        jobID,
		EventType: jobType,
		Status:    vbtypes.JobStatusPending,
		CreatedAt: time.Now(),
		VMName:    vmName,
	}

	// store job in memory
	jm.jobs[jobID] = job

	// vmRequest
	vmRequest := &vbtypes.VMRequest{
		Type:      jobType,
		VMID:      "",
		VMName:    vmName,
		VMStatus:  vbtypes.Stopped,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request": request,
			"jobID":   jobID,
		},
	}

	select {
	case jm.requestChannel <- vmRequest:
		return job, nil
	default:
		return nil, fmt.Errorf("request channel is full")
	}

}

// StartJob marks a job as running
func (jm *JobManager) StartJob(ctx context.Context, jobID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update status to running
	if job.Status == vbtypes.JobStatusPending {
		job.Status = vbtypes.JobStatusRunning
		now := time.Now()
		job.StartedAt = &now

	}

	return nil
}

func (jm *JobManager) MarkJobAsFailed(ctx context.Context, jobID string, errorMsg string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = vbtypes.JobStatusFailed
	job.Error = errorMsg
	now := time.Now()
	job.CompletedAt = &now

	return nil
}

func (jm *JobManager) ListJobs(ctx context.Context) ([]*vbtypes.Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*vbtypes.Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (jm *JobManager) GetJobProgress(ctx context.Context, jobID string) (*vbtypes.Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	return job, nil
}

func (jm *JobManager) CompleteJob(ctx context.Context, jobID string, vmID string, result map[string]interface{}) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = vbtypes.JobStatusCompleted
	job.VMID = vmID
	now := time.Now()
	job.CompletedAt = &now

	return nil
}

// startJobCleanup starts the background job cleanup routine
func (jm *JobManager) startJobCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := jm.cleanupOldJobs(24 * time.Hour); err != nil {
			}
		case <-ctx.Done():
			return
		}
	}
}

// CleanupOldJobs removes completed/failed jobs older than the specified duration
func (jm *JobManager) cleanupOldJobs(maxAge time.Duration) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var jobsToDelete []string

	for jobID, job := range jm.jobs {
		if (job.Status == vbtypes.JobStatusCompleted || job.Status == vbtypes.JobStatusFailed || job.Status == vbtypes.JobStatusCancelled) &&
			job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			jobsToDelete = append(jobsToDelete, jobID)
		}
	}

	for _, jobID := range jobsToDelete {
		delete(jm.jobs, jobID)
	}

	return nil
}

func generateJobID(jobType vbtypes.VMEventType, vmName string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%s-%d", jobType, vmName, timestamp)
}
