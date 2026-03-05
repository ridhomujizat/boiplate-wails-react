//go:build darwin
// +build darwin

package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var recordingStartTime time.Time
var recordingMutex sync.Mutex
var sckRecorderInstance *SCKAudioRecorder

func (r *RecorderManager) StartRecording() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status.State == StateRecording {
		return fmt.Errorf("recording already in progress")
	}

	timestamp := time.Now().Format("20060102_150405")
	r.tempVideoPath = filepath.Join(r.config.TempDir, fmt.Sprintf("video_%s.mp4", timestamp))
	r.tempAudioPath = filepath.Join(r.config.TempDir, fmt.Sprintf("audio_%s.wav", timestamp))
	r.tempSystemAudioPath = filepath.Join(r.config.TempDir, fmt.Sprintf("system_audio_%s.wav", timestamp))

	if err := os.MkdirAll(r.config.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	width, height, err := GetDisplayBounds(30)
	if err != nil {
		return fmt.Errorf("failed to get display bounds: %w", err)
	}

	scaleFilter := fmt.Sprintf("scale=%d:%d", width, height)
	screenCmd := exec.Command(GetFFmpegPath(),
		"-f", "avfoundation",
		"-capture_cursor", "1",
		"-framerate", "15",
		"-i", "Capture screen 0:none",
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "28",
		"-vf", scaleFilter,
		"-pix_fmt", "yuv420p",
		"-y",
		r.tempVideoPath,
	)

	screenCmd.Stderr = nil
	screenCmd.Stdout = nil

	if err := screenCmd.Start(); err != nil {
		return fmt.Errorf("failed to start screen recording: %w", err)
	}
	r.screenCmd = screenCmd

	recordingMutex.Lock()
	recordingStartTime = time.Now()
	recordingMutex.Unlock()

	if r.config.MicrophoneID != "" {
		audioRecorder, err := NewAudioRecorder(r.config.MicrophoneID, r.tempAudioPath)
		if err != nil {
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to create audio recorder: %w", err)
		}
		r.audioRecorder = audioRecorder
		if err := r.audioRecorder.Start(); err != nil {
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to start audio recording: %w", err)
		}
	}

	if r.config.SystemAudioEnabled && IsSystemAudioSupported() {
		sckRecorder, err := NewSCKAudioRecorder(r.tempSystemAudioPath)
		if err != nil {
			if r.audioRecorder != nil {
				r.audioRecorder.Stop()
			}
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to create system audio recorder: %w", err)
		}
		r.systemAudioRecorder = &SystemAudioRecorder{
			outputPath: r.tempSystemAudioPath,
			sampleRate: 48000,
			channels:   2,
		}
		if err := sckRecorder.Start(); err != nil {
			if r.audioRecorder != nil {
				r.audioRecorder.Stop()
			}
			screenCmd.Process.Kill()
			return fmt.Errorf("failed to start system audio recording: %w", err)
		}
		sckRecorderInstance = sckRecorder
	}

	r.status = RecordingStatus{
		State:     StateRecording,
		StartTime: time.Now(),
	}
	r.stopChan = make(chan struct{})

	// Setup auto-stop timer with callback
	r.setupAutoStopTimer(func() {
		// Timer expired, stop recording
		r.StopRecording()
	})

	return nil
}

func (r *RecorderManager) StopRecording() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status.State != StateRecording {
		return "", fmt.Errorf("no recording in progress")
	}

	r.cancelAutoStopTimer()
	r.status.State = StateProcessing

	if cmd, ok := r.screenCmd.(*exec.Cmd); ok && cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}

	if r.audioRecorder != nil {
		r.audioRecorder.Stop()
	}

	if sckRecorderInstance != nil {
		sckRecorderInstance.Stop()
		sckRecorderInstance = nil
	}

	// Use SessionId as filename if available, otherwise fallback to timestamp
	var outputName string
	if r.config.SessionId != "" {
		outputName = r.config.SessionId
	} else {
		outputName = fmt.Sprintf("recording_%s", time.Now().Format("20060102_150405"))
	}
	outputPath := filepath.Join(r.config.OutputDir, outputName+".mp4")

	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	hasMicAudio := r.audioRecorder != nil && r.tempAudioPath != "" && fileExists(r.tempAudioPath)
	hasSystemAudio := r.config.SystemAudioEnabled && IsSystemAudioSupported() && fileExists(r.tempSystemAudioPath)

	var err error
	if hasMicAudio && hasSystemAudio {
		mixedAudioPath := filepath.Join(r.config.TempDir, fmt.Sprintf("mixed_audio_%s.wav", outputName))
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
		err = os.Rename(r.tempVideoPath, outputPath)
	}

	if err != nil {
		r.status = RecordingStatus{State: StateError, Error: err.Error()}
		return "", err
	}

	os.Remove(r.tempVideoPath)
	os.Remove(r.tempAudioPath)
	os.Remove(r.tempSystemAudioPath)

	r.audioRecorder = nil
	r.systemAudioRecorder = nil

	r.status = RecordingStatus{
		State:    StateIdle,
		FilePath: outputPath,
	}

	return outputPath, nil
}

func (r *RecorderManager) muxVideoAudio(videoPath, audioPath, outputPath string) error {
	cmd := exec.Command(GetFFmpegPath(),
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-af", "aresample=async=1:first_pts=0",
		"-async", "1",
		"-shortest",
		"-y",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mux video and audio: %w", err)
	}

	return nil
}

func (r *RecorderManager) mixAudioFiles(audio1Path, audio2Path, outputPath string) error {
	cmd := exec.Command(GetFFmpegPath(),
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func convertToWebM(inputPath string) string {
	webmPath := inputPath[:len(inputPath)-4] + ".webm"

	cmd := exec.Command(GetFFmpegPath(),
		"-i", inputPath,
		"-c:v", "libvpx-vp9",
		"-crf", "40",
		"-b:v", "0",
		"-c:a", "libopus",
		"-b:a", "64k",
		"-y",
		webmPath,
	)

	cmd.Stderr = nil
	cmd.Stdout = nil

	if err := cmd.Run(); err != nil {
		return ""
	}

	return webmPath
}
