//go:build windows
// +build windows

package recorder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// stdinPipe holds the stdin pipe for FFmpeg process
var ffmpegStdin io.WriteCloser

// StartRecording starts screen and audio recording on Windows
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
	r.tempSystemAudioPath = filepath.Join(r.config.TempDir, fmt.Sprintf("system_audio_%s.wav", timestamp))

	// Ensure temp directory exists
	if err := os.MkdirAll(r.config.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Start FFmpeg screen capture using gdigrab
	screenCmd := exec.Command("ffmpeg",
		"-f", "gdigrab",
		"-framerate", "30",
		"-i", "desktop",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y",
		r.tempVideoPath,
	)

	// Redirect stderr to suppress FFmpeg output
	screenCmd.Stderr = nil
	screenCmd.Stdout = nil

	// Create stdin pipe for graceful shutdown (send 'q' to stop)
	stdin, err := screenCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	ffmpegStdin = stdin

	if err := screenCmd.Start(); err != nil {
		return fmt.Errorf("failed to start screen recording: %w", err)
	}
	r.screenCmd = screenCmd

	// Start microphone recording if microphone is selected
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

	// Start system audio recording if enabled
	if r.config.SystemAudioEnabled {
		systemAudioRecorder, err := NewSystemAudioRecorder(r.tempSystemAudioPath)
		if err != nil {
			// Stop other recordings if system audio fails
			if r.audioRecorder != nil {
				r.audioRecorder.Stop()
			}
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to start system audio recording: %w", err)
		}
		r.systemAudioRecorder = systemAudioRecorder
		go r.systemAudioRecorder.Start()
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

	// Stop screen recording gracefully by sending 'q' to FFmpeg
	if cmd, ok := r.screenCmd.(*exec.Cmd); ok && cmd.Process != nil {
		// Send 'q' to FFmpeg stdin for graceful shutdown
		// This allows FFmpeg to properly finalize the video file
		if ffmpegStdin != nil {
			ffmpegStdin.Write([]byte("q"))
			ffmpegStdin.Close()
			ffmpegStdin = nil
		}
		// Wait for FFmpeg to finish gracefully
		cmd.Wait()
	}

	// Small delay to ensure file handles are released
	time.Sleep(500 * time.Millisecond)

	// Stop microphone audio recording
	if r.audioRecorder != nil {
		r.audioRecorder.Stop()
	}

	// Stop system audio recording
	if r.systemAudioRecorder != nil {
		r.systemAudioRecorder.Stop()
	}

	// Generate output file path
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(r.config.OutputDir, fmt.Sprintf("recording_%s.mp4", timestamp))

	// Ensure output directory exists
	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine which audio sources are available
	hasMicAudio := r.audioRecorder != nil && r.tempAudioPath != ""
	hasSystemAudio := r.systemAudioRecorder != nil && r.tempSystemAudioPath != ""

	// Mux video and audio based on available sources
	var err error
	if hasMicAudio && hasSystemAudio {
		// Mix both audio sources and mux with video
		mixedAudioPath := filepath.Join(r.config.TempDir, fmt.Sprintf("mixed_audio_%s.wav", timestamp))
		err = r.mixAudioFiles(r.tempAudioPath, r.tempSystemAudioPath, mixedAudioPath)
		if err == nil {
			err = r.muxVideoAudio(r.tempVideoPath, mixedAudioPath, outputPath)
			os.Remove(mixedAudioPath)
		}
	} else if hasMicAudio {
		err = r.muxVideoAudio(r.tempVideoPath, r.tempAudioPath, outputPath)
	} else if hasSystemAudio {
		err = r.muxVideoAudio(r.tempVideoPath, r.tempSystemAudioPath, outputPath)
	} else {
		// Just copy video if no audio - use retry mechanism for Windows file locking
		err = r.renameWithRetry(r.tempVideoPath, outputPath, 5, 100*time.Millisecond)
	}

	if err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", err
	}

	// Cleanup temp files
	os.Remove(r.tempVideoPath)
	os.Remove(r.tempAudioPath)
	os.Remove(r.tempSystemAudioPath)

	// Reset recorders
	r.audioRecorder = nil
	r.systemAudioRecorder = nil

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

// mixAudioFiles mixes two audio files into one using FFmpeg's amix filter
func (r *RecorderManager) mixAudioFiles(audio1Path, audio2Path, outputPath string) error {
	// Use FFmpeg amix filter to mix both audio streams
	cmd := exec.Command("ffmpeg",
		"-i", audio1Path,
		"-i", audio2Path,
		"-filter_complex", "amix=inputs=2:duration=longest:dropout_transition=0",
		"-c:a", "pcm_s16le",
		"-y",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mix audio files: %w", err)
	}

	return nil
}

// renameWithRetry attempts to rename a file with exponential backoff retry
// This handles Windows file locking issues where FFmpeg may hold the file briefly after termination
func (r *RecorderManager) renameWithRetry(src, dst string, maxRetries int, initialDelay time.Duration) error {
	var lastErr error
	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		lastErr = os.Rename(src, dst)
		if lastErr == nil {
			return nil
		}

		// Wait before retrying with exponential backoff
		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to rename file after %d retries: %w", maxRetries, lastErr)
}
