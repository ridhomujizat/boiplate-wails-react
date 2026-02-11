package recorder

import (
	"fmt"
	"sync"
	"time"
)

// RecordingState represents the current state of recording
type RecordingState string

const (
	StateIdle       RecordingState = "idle"
	StateRecording  RecordingState = "recording"
	StateProcessing RecordingState = "processing"
	StateError      RecordingState = "error"
)

// RecordingStatus represents the current recording status
type RecordingStatus struct {
	State     RecordingState `json:"state"`
	Duration  int64          `json:"duration"` // seconds
	FilePath  string         `json:"filePath"`
	Error     string         `json:"error"`
	StartTime time.Time      `json:"-"`
}

// RecordingConfig holds recording configuration
type RecordingConfig struct {
	MicrophoneID            string
	SystemAudioEnabled      bool
	OutputDir               string
	TempDir                 string
	MaxRecordingTimeEnabled bool
	MaxRecordingTimeSeconds int
	SessionId               string // Used for output filename; falls back to timestamp if empty
}

// RecorderManager manages screen and audio recording
type RecorderManager struct {
	mu                  sync.Mutex
	status              RecordingStatus
	config              RecordingConfig
	stopChan            chan struct{}
	autoStopTimer       *time.Timer
	screenCmd           interface{} // *exec.Cmd, platform specific
	audioRecorder       *AudioRecorder
	systemAudioRecorder *SystemAudioRecorder
	tempVideoPath       string
	tempAudioPath       string
	tempSystemAudioPath string
}

// NewRecorderManager creates a new recorder manager
func NewRecorderManager(config RecordingConfig) *RecorderManager {
	return &RecorderManager{
		status: RecordingStatus{
			State: StateIdle,
		},
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// GetStatus returns the current recording status
func (r *RecorderManager) GetStatus() RecordingStatus {
	r.mu.Lock()
	defer r.mu.Unlock()

	status := r.status
	if status.State == StateRecording && !status.StartTime.IsZero() {
		status.Duration = int64(time.Since(status.StartTime).Seconds())
	}
	return status
}

// UpdateConfig updates the recording configuration
func (r *RecorderManager) UpdateConfig(config RecordingConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = config
}

// setupAutoStopTimer sets up a timer to automatically stop recording after the configured duration
func (r *RecorderManager) setupAutoStopTimer(onAutoStop func()) {
	if !r.config.MaxRecordingTimeEnabled || r.config.MaxRecordingTimeSeconds <= 0 {
		return
	}
	duration := time.Duration(r.config.MaxRecordingTimeSeconds) * time.Second
	r.autoStopTimer = time.AfterFunc(duration, func() {
		if onAutoStop != nil {
			onAutoStop()
		}
	})
}

// cancelAutoStopTimer stops and clears the auto-stop timer
func (r *RecorderManager) cancelAutoStopTimer() {
	fmt.Println("Cancelling auto-stop timer")
	if r.autoStopTimer != nil {
		r.autoStopTimer.Stop()
		r.autoStopTimer = nil
	}
}
