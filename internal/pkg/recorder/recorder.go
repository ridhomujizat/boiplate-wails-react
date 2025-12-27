package recorder

import (
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
	MicrophoneID       string
	SystemAudioEnabled bool
	OutputDir          string
	TempDir            string
}

// RecorderManager manages screen and audio recording
type RecorderManager struct {
	mu            sync.Mutex
	status        RecordingStatus
	config        RecordingConfig
	stopChan      chan struct{}
	screenCmd     interface{} // *exec.Cmd, platform specific
	audioRecorder *AudioRecorder
	tempVideoPath string
	tempAudioPath string
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
