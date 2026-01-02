package activity

import (
	"log"
	"runtime"
	"sync"
	"time"

	models "onx-screen-record/internal/common/model"
	"onx-screen-record/internal/pkg/activity"
	"onx-screen-record/internal/repository"
)

const (
	DefaultPollingInterval = 5 * time.Second
	DefaultAFKThreshold    = 1 * time.Minute
)

type TrackerConfig struct {
	PollingInterval time.Duration
	AFKThreshold    time.Duration
	Enabled         bool
}

type Tracker struct {
	mu            sync.RWMutex
	config        TrackerConfig
	repo          *repository.IRepository
	currentEvent  *models.ActivityEvent
	isRunning     bool
	stopChan      chan struct{}
	lastWindowKey string
}

func NewTracker(repo *repository.IRepository, config TrackerConfig) *Tracker {
	if config.PollingInterval == 0 {
		config.PollingInterval = DefaultPollingInterval
	}
	if config.AFKThreshold == 0 {
		config.AFKThreshold = DefaultAFKThreshold
	}

	return &Tracker{
		config:   config,
		repo:     repo,
		stopChan: make(chan struct{}),
	}
}

func (t *Tracker) Start() {
	t.mu.Lock()
	if t.isRunning {
		t.mu.Unlock()
		return
	}
	t.isRunning = true
	t.stopChan = make(chan struct{})
	t.mu.Unlock()

	go t.trackingLoop()
	log.Println("[INFO] Activity tracker started")
}

func (t *Tracker) Stop() {
	t.mu.Lock()
	if !t.isRunning {
		t.mu.Unlock()
		return
	}
	t.isRunning = false
	close(t.stopChan)
	t.mu.Unlock()

	t.finalizeCurrentEvent()
	log.Println("[INFO] Activity tracker stopped")
}

func (t *Tracker) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isRunning
}

func (t *Tracker) trackingLoop() {
	ticker := time.NewTicker(t.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case <-ticker.C:
			t.captureActivity()
		}
	}
}

func (t *Tracker) captureActivity() {
	windowInfo := activity.GetActiveWindow()
	idleTimeMs := activity.GetIdleTimeMs()

	isAFK := idleTimeMs >= t.config.AFKThreshold.Milliseconds()

	windowKey := windowInfo.AppName + "|" + windowInfo.WindowTitle

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	if t.currentEvent != nil {
		windowChanged := t.lastWindowKey != windowKey
		statusChanged := (t.currentEvent.Status == models.StatusAFK) != isAFK

		if windowChanged || statusChanged {
			t.currentEvent.EndTime = now
			t.currentEvent.Duration = int64(now.Sub(t.currentEvent.StartTime).Seconds())
			if err := t.repo.Activity.Update(t.currentEvent); err != nil {
				log.Printf("[ERROR] Failed to update activity event: %v", err)
			}
			t.currentEvent = nil
		}
	}

	if t.currentEvent == nil && windowInfo.AppName != "" {
		status := models.StatusActive
		if isAFK {
			status = models.StatusAFK
		}

		event := &models.ActivityEvent{
			AppName:     windowInfo.AppName,
			BundleID:    windowInfo.BundleID,
			WindowTitle: windowInfo.WindowTitle,
			StartTime:   now,
			Status:      status,
			OS:          runtime.GOOS,
		}

		if err := t.repo.Activity.Create(event); err != nil {
			log.Printf("[ERROR] Failed to create activity event: %v", err)
			return
		}

		t.currentEvent = event
		t.lastWindowKey = windowKey
	}
}

func (t *Tracker) finalizeCurrentEvent() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentEvent != nil {
		now := time.Now()
		t.currentEvent.EndTime = now
		t.currentEvent.Duration = int64(now.Sub(t.currentEvent.StartTime).Seconds())
		if err := t.repo.Activity.Update(t.currentEvent); err != nil {
			log.Printf("[ERROR] Failed to finalize activity event: %v", err)
		}
		t.currentEvent = nil
	}
}

func (t *Tracker) UpdateConfig(pollingIntervalSec, afkThresholdSec int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if pollingIntervalSec > 0 {
		t.config.PollingInterval = time.Duration(pollingIntervalSec) * time.Second
	}
	if afkThresholdSec > 0 {
		t.config.AFKThreshold = time.Duration(afkThresholdSec) * time.Second
	}
}
