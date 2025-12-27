package recorder

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

//go:embed notif.mp3
var notifSound []byte

// SystemAudioRecorder handles system audio (loopback) recording using malgo
type SystemAudioRecorder struct {
	mu          sync.Mutex
	ctx         *malgo.AllocatedContext
	device      *malgo.Device
	outputPath  string
	isRecording bool
	sampleRate  uint32
	channels    uint32
	samples     []byte
}

// NewSystemAudioRecorder creates a new system audio recorder for loopback capture
func NewSystemAudioRecorder(outputPath string) (*SystemAudioRecorder, error) {
	return &SystemAudioRecorder{
		outputPath: outputPath,
		sampleRate: 44100,
		channels:   2,
		samples:    make([]byte, 0),
	}, nil
}

// playTriggerSound plays a short notification sound to wake up the loopback capture
func (s *SystemAudioRecorder) playTriggerSound() error {
	// Decode mp3 from embedded bytes
	reader := io.NopCloser(bytes.NewReader(notifSound))
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		return fmt.Errorf("failed to decode trigger sound: %w", err)
	}
	defer streamer.Close()

	// Initialize speaker with the format
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("failed to init speaker: %w", err)
	}

	// Lower volume to 10% (-20 dB)
	// Volume in beep uses decibels, -20 dB ≈ 10% volume
	volume := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   -20, // -20 dB = 10% volume
		Silent:   false,
	}

	// Create a channel to wait for playback to complete
	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	// Wait for sound to finish
	<-done

	return nil
}

// Start begins system audio recording (loopback capture)
func (s *SystemAudioRecorder) Start() error {
	s.mu.Lock()
	if s.isRecording {
		s.mu.Unlock()
		return fmt.Errorf("already recording")
	}
	s.isRecording = true
	s.samples = make([]byte, 0)
	s.mu.Unlock()

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to init context: %w", err)
	}
	s.ctx = ctx

	// Callback to receive audio samples
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.isRecording {
			s.samples = append(s.samples, pSample...)
		}
	}

	// Configure device for loopback capture (system audio)
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to init loopback device: %w", err)
	}

	s.device = device

	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to start loopback device: %w", err)
	}

	// Play trigger sound to wake up loopback capture and ensure sync
	// This runs in background so recording can start capturing immediately
	go func() {
		s.playTriggerSound()
	}()

	return nil
}

// Stop stops system audio recording and writes to file
func (s *SystemAudioRecorder) Stop() error {
	s.mu.Lock()
	s.isRecording = false
	samples := s.samples
	s.mu.Unlock()

	if s.device != nil {
		s.device.Stop()
		s.device.Uninit()
	}

	if s.ctx != nil {
		s.ctx.Uninit()
		s.ctx.Free()
	}

	// Write WAV file
	return s.writeWAV(samples)
}

// writeWAV writes the recorded samples to a WAV file
func (s *SystemAudioRecorder) writeWAV(samples []byte) error {
	file, err := os.Create(s.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	// WAV header
	numSamples := uint32(len(samples))
	bitsPerSample := uint16(16)
	byteRate := s.sampleRate * uint32(s.channels) * uint32(bitsPerSample) / 8
	blockAlign := uint16(s.channels) * bitsPerSample / 8

	// Write RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+numSamples)) // File size - 8
	file.WriteString("WAVE")

	// Write fmt subchunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))         // Subchunk1 size
	binary.Write(file, binary.LittleEndian, uint16(1))          // Audio format (PCM)
	binary.Write(file, binary.LittleEndian, uint16(s.channels)) // Num channels
	binary.Write(file, binary.LittleEndian, s.sampleRate)       // Sample rate
	binary.Write(file, binary.LittleEndian, byteRate)           // Byte rate
	binary.Write(file, binary.LittleEndian, blockAlign)         // Block align
	binary.Write(file, binary.LittleEndian, bitsPerSample)      // Bits per sample

	// Write data subchunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, numSamples) // Subchunk2 size
	file.Write(samples)

	return nil
}
