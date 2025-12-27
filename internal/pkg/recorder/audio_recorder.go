package recorder

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/gen2brain/malgo"
)

// AudioRecorder handles audio recording using malgo
type AudioRecorder struct {
	mu           sync.Mutex
	device       *malgo.Device
	deviceConfig malgo.DeviceConfig
	outputPath   string
	file         *os.File
	isRecording  bool
	sampleRate   uint32
	channels     uint32
	samples      []byte
}

// NewAudioRecorder creates a new audio recorder for the specified device
func NewAudioRecorder(deviceID string, outputPath string) (*AudioRecorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init context: %w", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// Configure device for capture
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100

	// Parse device ID if provided
	if deviceID != "" {
		// Device ID is expected to be passed; we'll use default if empty
		// In production, you might want to match the deviceID to actual devices
	}

	return &AudioRecorder{
		deviceConfig: deviceConfig,
		outputPath:   outputPath,
		sampleRate:   44100,
		channels:     2,
		samples:      make([]byte, 0),
	}, nil
}

// Start begins audio recording
func (a *AudioRecorder) Start() error {
	a.mu.Lock()
	if a.isRecording {
		a.mu.Unlock()
		return fmt.Errorf("already recording")
	}
	a.isRecording = true
	a.samples = make([]byte, 0)
	a.mu.Unlock()

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to init context: %w", err)
	}

	// Callback to receive audio samples
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.isRecording {
			a.samples = append(a.samples, pSample...)
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to init device: %w", err)
	}

	a.device = device

	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to start device: %w", err)
	}

	return nil
}

// Stop stops audio recording and writes to file
func (a *AudioRecorder) Stop() error {
	a.mu.Lock()
	a.isRecording = false
	samples := a.samples
	a.mu.Unlock()

	if a.device != nil {
		a.device.Stop()
		a.device.Uninit()
	}

	// Write WAV file
	return a.writeWAV(samples)
}

// writeWAV writes the recorded samples to a WAV file
func (a *AudioRecorder) writeWAV(samples []byte) error {
	file, err := os.Create(a.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	// WAV header
	numSamples := uint32(len(samples))
	bitsPerSample := uint16(16)
	byteRate := a.sampleRate * uint32(a.channels) * uint32(bitsPerSample) / 8
	blockAlign := uint16(a.channels) * bitsPerSample / 8

	// Write RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+numSamples)) // File size - 8
	file.WriteString("WAVE")

	// Write fmt subchunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))         // Subchunk1 size
	binary.Write(file, binary.LittleEndian, uint16(1))          // Audio format (PCM)
	binary.Write(file, binary.LittleEndian, uint16(a.channels)) // Num channels
	binary.Write(file, binary.LittleEndian, a.sampleRate)       // Sample rate
	binary.Write(file, binary.LittleEndian, byteRate)           // Byte rate
	binary.Write(file, binary.LittleEndian, blockAlign)         // Block align
	binary.Write(file, binary.LittleEndian, bitsPerSample)      // Bits per sample

	// Write data subchunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, numSamples) // Subchunk2 size
	file.Write(samples)

	return nil
}
