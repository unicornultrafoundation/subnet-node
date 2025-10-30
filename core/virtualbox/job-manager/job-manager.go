package job_manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

type JobManager struct {
	mu             sync.RWMutex
	requestChannel chan *vbtypes.VMRequest

	jobs map[string]*vbtypes.Job
}

// jobManager with 100 request channel buffer
func NewJobManager() *JobManager {
	return &JobManager{
		requestChannel: make(chan *vbtypes.VMRequest, 100),
		jobs:           make(map[string]*vbtypes.Job),
	}
}

func (jm *JobManager) GetRequestChannel() chan *vbtypes.VMRequest {
	return jm.requestChannel
}

func (jm *JobManager) Start(ctx context.Context) error {
	go jm.startJobCleanup(ctx)
	return nil
}

func (jm *JobManager) CreateJob(ctx context.Context, jobType vbtypes.VMEventType, request map[string]interface{}, vmName string) (*vbtypes.Job, error) {
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
