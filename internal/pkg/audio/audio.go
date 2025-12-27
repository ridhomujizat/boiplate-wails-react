package audio

import (
	"github.com/gen2brain/malgo"
)

// AudioDevice represents an audio device
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "capture" or "playback"
}

// AudioManager handles audio device enumeration
type AudioManager struct{}

// NewAudioManager creates a new audio manager
func NewAudioManager() *AudioManager {
	return &AudioManager{}
}

// GetCaptureDevices returns a list of available capture (microphone) devices
func (a *AudioManager) GetCaptureDevices() ([]AudioDevice, error) {
	return a.getDevices(malgo.Capture, "capture")
}

// GetPlaybackDevices returns a list of available playback (speaker) devices
func (a *AudioManager) GetPlaybackDevices() ([]AudioDevice, error) {
	return a.getDevices(malgo.Playback, "playback")
}

// GetAllDevices returns all audio devices (capture and playback)
func (a *AudioManager) GetAllDevices() ([]AudioDevice, error) {
	captureDevices, err := a.GetCaptureDevices()
	if err != nil {
		return nil, err
	}

	playbackDevices, err := a.GetPlaybackDevices()
	if err != nil {
		return nil, err
	}

	return append(captureDevices, playbackDevices...), nil
}

// getDevices retrieves devices of a specific type
func (a *AudioManager) getDevices(deviceType malgo.DeviceType, typeStr string) ([]AudioDevice, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	devices, err := ctx.Devices(deviceType)
	if err != nil {
		return nil, err
	}

	result := make([]AudioDevice, 0, len(devices))
	for _, device := range devices {
		result = append(result, AudioDevice{
			ID:   device.ID.String(),
			Name: device.Name(),
			Type: typeStr,
		})
	}

	return result, nil
}
