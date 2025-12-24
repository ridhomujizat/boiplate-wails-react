package cronjob

import (
	"context"
	"fmt"
	"onx-screen-record/internal/pkg/logger"
	"sync"
	"time"
)

// JobFunc defines the function signature for cron jobs
type JobFunc func(ctx context.Context) error

// Job represents a scheduled job
type Job struct {
	Name     string
	Interval time.Duration
	Fn       JobFunc
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex
}

// Scheduler manages multiple cron jobs
type Scheduler struct {
	jobs []*Job
	mu   sync.RWMutex
	ctx  context.Context
}

// NewScheduler creates a new cron job scheduler
func NewScheduler(ctx context.Context) *Scheduler {
	return &Scheduler{
		jobs: make([]*Job, 0),
		ctx:  ctx,
	}
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(name string, interval time.Duration, fn JobFunc) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &Job{
		Name:     name,
		Interval: interval,
		Fn:       fn,
		stopChan: make(chan struct{}),
		running:  false,
	}

	s.jobs = append(s.jobs, job)
	logger.Info.Printf("Cron job '%s' added with interval: %v", name, interval)

	return job
}

// Start starts a specific job
func (s *Scheduler) Start(jobName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if job.Name == jobName {
			return job.Start(s.ctx)
		}
	}

	return fmt.Errorf("job '%s' not found", jobName)
}

// Stop stops a specific job
func (s *Scheduler) Stop(jobName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if job.Name == jobName {
			return job.Stop()
		}
	}

	return fmt.Errorf("job '%s' not found", jobName)
}

// StartAll starts all registered jobs
func (s *Scheduler) StartAll() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if err := job.Start(s.ctx); err != nil {
			logger.Error.Printf("Failed to start job '%s': %v", job.Name, err)
		}
	}
}

// StopAll stops all running jobs
func (s *Scheduler) StopAll() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if err := job.Stop(); err != nil {
			logger.Error.Printf("Failed to stop job '%s': %v", job.Name, err)
		}
	}
}

// GetJobStatus returns the status of a specific job
func (s *Scheduler) GetJobStatus(jobName string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if job.Name == jobName {
			return job.IsRunning(), nil
		}
	}

	return false, fmt.Errorf("job '%s' not found", jobName)
}

// ListJobs returns all registered jobs with their status
func (s *Scheduler) ListJobs() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobList := make([]map[string]interface{}, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobList = append(jobList, map[string]interface{}{
			"name":     job.Name,
			"interval": job.Interval.String(),
			"running":  job.IsRunning(),
		})
	}

	return jobList
}

// Start starts the job execution
func (j *Job) Start(ctx context.Context) error {
	j.mu.Lock()
	if j.running {
		j.mu.Unlock()
		return fmt.Errorf("job '%s' is already running", j.Name)
	}
	j.running = true
	j.stopChan = make(chan struct{})
	j.mu.Unlock()

	logger.Info.Printf("Starting cron job '%s' with interval: %v", j.Name, j.Interval)

	go j.run(ctx)

	return nil
}

// Stop stops the job execution
func (j *Job) Stop() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if !j.running {
		return fmt.Errorf("job '%s' is not running", j.Name)
	}

	logger.Info.Printf("Stopping cron job '%s'", j.Name)
	close(j.stopChan)
	j.running = false

	return nil
}

// IsRunning returns whether the job is currently running
func (j *Job) IsRunning() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.running
}

// run executes the job at specified intervals
func (j *Job) run(ctx context.Context) {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()

	// Execute immediately on start
	j.execute(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info.Printf("Context cancelled, stopping job '%s'", j.Name)
			return
		case <-j.stopChan:
			logger.Info.Printf("Job '%s' stopped", j.Name)
			return
		case <-ticker.C:
			j.execute(ctx)
		}
	}
}

// execute runs the job function and logs any errors
func (j *Job) execute(ctx context.Context) {
	logger.Debug.Printf("Executing job '%s'", j.Name)

	start := time.Now()
	err := j.Fn(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.Error.Printf("Job '%s' failed after %v: %v", j.Name, duration, err)
	} else {
		logger.Debug.Printf("Job '%s' completed successfully in %v", j.Name, duration)
	}
}
