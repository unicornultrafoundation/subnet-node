package virtualbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var jobLog = logrus.WithField("service", "virtualbox-job-manager")

// JobManager manages background jobs and their progress
type JobManager struct {
	mu             sync.RWMutex
	requestChannel chan *vbtypes.VMRequest

	jobs map[string]*vbtypes.Job
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		requestChannel: make(chan *vbtypes.VMRequest, 100),
		jobs:           make(map[string]*vbtypes.Job),
	}
}

func (jm *JobManager) Start(ctx context.Context) error {
	go jm.startJobCleanup(ctx)
	return nil
}

// CreateJob creates a new job and returns its ID
func (jm *JobManager) CreateJob(ctx context.Context, jobType vbtypes.JobType, request map[string]interface{}, vmName string) (*vbtypes.Job, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jobID := generateJobID(jobType, vmName)

	job := &vbtypes.Job{
		ID:        jobID,
		Type:      jobType,
		Status:    vbtypes.JobStatusPending,
		Request:   request,
		CreatedAt: time.Now(),
		VMName:    vmName,
	}

	// Store job in memory
	jm.jobs[jobID] = job

	jobLog.WithFields(logrus.Fields{
		"jobID":  jobID,
		"type":   jobType,
		"vmName": vmName,
	}).Info("Created new job")

	return job, nil
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(ctx context.Context, jobID string) (*vbtypes.Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	if job, exists := jm.jobs[jobID]; exists {
		return job, nil
	}

	return nil, fmt.Errorf("job not found: %s", jobID)
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

		jobLog.WithFields(logrus.Fields{
			"jobID": jobID,
		}).Info("Job started")
	}

	return nil
}

// CompleteJob marks a job as completed with optional result
func (jm *JobManager) CompleteJob(ctx context.Context, jobID string, result map[string]interface{}) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = vbtypes.JobStatusCompleted
	job.Result = result
	now := time.Now()
	job.CompletedAt = &now

	jobLog.WithFields(logrus.Fields{
		"jobID": jobID,
		"type":  job.Type,
	}).Info("Job completed successfully")

	return nil
}

// FailJob marks a job as failed with an error message
func (jm *JobManager) FailJob(ctx context.Context, jobID string, errorMsg string) error {
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

	jobLog.WithFields(logrus.Fields{
		"jobID": jobID,
		"type":  job.Type,
		"error": errorMsg,
	}).Error("Job failed")

	return nil
}

// CancelJob cancels a job
func (jm *JobManager) CancelJob(ctx context.Context, jobID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == vbtypes.JobStatusCompleted || job.Status == vbtypes.JobStatusFailed {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	job.Status = vbtypes.JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	jobLog.WithFields(logrus.Fields{
		"jobID": jobID,
		"type":  job.Type,
	}).Info("Job cancelled")

	return nil
}

// ListJobs returns all jobs
func (jm *JobManager) ListJobs(ctx context.Context) ([]*vbtypes.Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	var jobs []*vbtypes.Job
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CleanupOldJobs removes completed/failed jobs older than the specified duration
func (jm *JobManager) CleanupOldJobs(ctx context.Context, maxAge time.Duration) error {
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

	if len(jobsToDelete) > 0 {
		jobLog.WithField("count", len(jobsToDelete)).Info("Cleaned up old jobs")
	}

	return nil
}

func generateJobID(jobType vbtypes.JobType, vmName string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%s-%d", jobType, vmName, timestamp)
}

// GetRequestChannel returns the request channel for VM operations
func (jm *JobManager) GetRequestChannel() chan *vbtypes.VMRequest {
	return jm.requestChannel
}

// SendRequest sends a VM request to the channel
func (jm *JobManager) SendRequest(request *vbtypes.VMRequest) error {
	select {
	case jm.requestChannel <- request:
		jobLog.WithFields(logrus.Fields{
			"requestType": request.Type,
			"vmName":      request.VMName,
		}).Info("VM request sent to channel")
		return nil
	default:
		return fmt.Errorf("request channel is full")
	}
}

// startJobCleanup starts the background job cleanup routine
func (jm *JobManager) startJobCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := jm.CleanupOldJobs(ctx, 24*time.Hour); err != nil {
				jobLog.WithError(err).Warn("Failed to cleanup old jobs")
			}
		case <-ctx.Done():
			return
		}
	}
}
