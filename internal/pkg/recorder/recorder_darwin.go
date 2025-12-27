//go:build darwin
// +build darwin

package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// StartRecording starts screen and audio recording on macOS
func (r *RecorderManager) StartRecording() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status.State == StateRecording {
		return fmt.Errorf("recording already in progress")
	}

	// Generate temp file paths
	timestamp := time.Now().Format("20060102_150405")
	r.tempVideoPath = filepath.Join(r.config.TempDir, fmt.Sprintf("video_%s.mp4", timestamp))
	r.tempAudioPath = filepath.Join(r.config.TempDir, fmt.Sprintf("audio_%s.wav", timestamp))

	// Ensure temp directory exists
	if err := os.MkdirAll(r.config.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Start FFmpeg screen capture using avfoundation
	// Use "Capture screen 0:none" for screen capture (not camera)
	// Index 0 is typically the main screen on macOS
	screenCmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-capture_cursor", "1",
		"-framerate", "30",
		"-i", "Capture screen 0:none",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y",
		r.tempVideoPath,
	)

	// Redirect stderr to suppress FFmpeg output
	screenCmd.Stderr = nil
	screenCmd.Stdout = nil

	if err := screenCmd.Start(); err != nil {
		return fmt.Errorf("failed to start screen recording: %w", err)
	}
	r.screenCmd = screenCmd

	// Start audio recording if microphone is selected
	if r.config.MicrophoneID != "" {
		audioRecorder, err := NewAudioRecorder(r.config.MicrophoneID, r.tempAudioPath)
		if err != nil {
			// Stop screen recording if audio fails
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to start audio recording: %w", err)
		}
		r.audioRecorder = audioRecorder
		go r.audioRecorder.Start()
	}

	r.status = RecordingStatus{
		State:     StateRecording,
		StartTime: time.Now(),
	}
	r.stopChan = make(chan struct{})

	return nil
}

// StopRecording stops recording and muxes video + audio
func (r *RecorderManager) StopRecording() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status.State != StateRecording {
		return "", fmt.Errorf("no recording in progress")
	}

	r.status.State = StateProcessing

	// Stop screen recording
	if cmd, ok := r.screenCmd.(*exec.Cmd); ok && cmd.Process != nil {
		// Send 'q' to FFmpeg to gracefully stop
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}

	// Stop audio recording
	if r.audioRecorder != nil {
		r.audioRecorder.Stop()
	}

	// Generate output file path
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(r.config.OutputDir, fmt.Sprintf("recording_%s.mp4", timestamp))

	// Ensure output directory exists
	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Mux video and audio if audio was recorded
	var err error
	if r.audioRecorder != nil && r.tempAudioPath != "" {
		err = r.muxVideoAudio(r.tempVideoPath, r.tempAudioPath, outputPath)
	} else {
		// Just copy video if no audio
		err = os.Rename(r.tempVideoPath, outputPath)
	}

	if err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", err
	}

	// Cleanup temp files
	os.Remove(r.tempVideoPath)
	os.Remove(r.tempAudioPath)

	r.status = RecordingStatus{
		State:    StateIdle,
		FilePath: outputPath,
	}

	return outputPath, nil
}

// muxVideoAudio combines video and audio using FFmpeg
func (r *RecorderManager) muxVideoAudio(videoPath, audioPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-shortest",
		"-y",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mux video and audio: %w", err)
	}

	return nil
}
