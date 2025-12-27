package app

import (
	"onx-screen-record/internal/pkg/audio"
	"onx-screen-record/internal/pkg/permission"
	"onx-screen-record/internal/pkg/recorder"
	dtoSetting "onx-screen-record/internal/service/setting/dto"
)

// SaveSettingsResponse is the response type after saving settings (exported for Wails)
type SaveSettingsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PermissionStatus represents the status of a system permission
type PermissionStatus struct {
	Granted bool   `json:"granted"`
	Message string `json:"message"`
}

// AudioDevice represents an available audio device
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetSettings retrieves all settings from the database
func (a *App) GetSettings() dtoSetting.SettingResponse {

	result, err := a.setting.GetSettings()
	if err != nil {
		return dtoSetting.SettingResponse{}
	}

	return *result
}

// SaveSettings saves settings to the database
func (a *App) SaveSettings(req dtoSetting.SettingRequest) SaveSettingsResponse {

	result, err := a.setting.SaveSettings(req)

	if err != nil {
		return SaveSettingsResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	return SaveSettingsResponse{
		Success: result.Success,
		Message: result.Message,
	}
}

// CheckScreenPermission checks if screen recording permission is granted
func (a *App) CheckScreenPermission() PermissionStatus {
	pm := permission.NewPermissionManager()
	status := pm.CheckScreenPermission()
	return PermissionStatus{
		Granted: status.Granted,
		Message: status.Message,
	}
}

// RequestScreenPermission requests screen recording permission
func (a *App) RequestScreenPermission() bool {
	pm := permission.NewPermissionManager()
	return pm.RequestScreenPermission()
}

// CheckAccessibilityPermission checks if accessibility permission is granted
func (a *App) CheckAccessibilityPermission() PermissionStatus {
	pm := permission.NewPermissionManager()
	status := pm.CheckAccessibilityPermission()
	return PermissionStatus{
		Granted: status.Granted,
		Message: status.Message,
	}
}

// RequestAccessibilityPermission requests accessibility permission
func (a *App) RequestAccessibilityPermission() bool {
	pm := permission.NewPermissionManager()
	return pm.RequestAccessibilityPermission()
}

// GetAudioDevices returns all available audio devices
func (a *App) GetAudioDevices() []AudioDevice {
	am := audio.NewAudioManager()
	devices, err := am.GetAllDevices()
	if err != nil {
		return []AudioDevice{}
	}

	result := make([]AudioDevice, 0, len(devices))
	for _, d := range devices {
		result = append(result, AudioDevice{
			ID:   d.ID,
			Name: d.Name,
			Type: d.Type,
		})
	}
	return result
}

// GetCaptureDevices returns available microphone/input devices
func (a *App) GetCaptureDevices() []AudioDevice {
	am := audio.NewAudioManager()
	devices, err := am.GetCaptureDevices()
	if err != nil {
		return []AudioDevice{}
	}

	result := make([]AudioDevice, 0, len(devices))
	for _, d := range devices {
		result = append(result, AudioDevice{
			ID:   d.ID,
			Name: d.Name,
			Type: d.Type,
		})
	}
	return result
}

// GetAudioSettings retrieves audio settings from the database
func (a *App) GetAudioSettings() dtoSetting.AudioSettingResponse {
	result, err := a.setting.GetAudioSettings()
	if err != nil {
		return dtoSetting.AudioSettingResponse{}
	}
	return *result
}

// SaveAudioSettings saves audio settings to the database
func (a *App) SaveAudioSettings(req dtoSetting.AudioSettingRequest) SaveSettingsResponse {
	result, err := a.setting.SaveAudioSettings(req)
	if err != nil {
		return SaveSettingsResponse{
			Success: false,
			Message: err.Error(),
		}
	}
	return SaveSettingsResponse{
		Success: result.Success,
		Message: result.Message,
	}
}

// RecordingStatus represents the current recording state
type RecordingStatus struct {
	State    string `json:"state"`
	Duration int64  `json:"duration"`
	FilePath string `json:"filePath"`
	Error    string `json:"error"`
}

// StartRecordingResponse is the response after starting recording
type StartRecordingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// StopRecordingResponse is the response after stopping recording
type StopRecordingResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FilePath string `json:"filePath"`
}

// StartRecording starts screen and audio recording
func (a *App) StartRecording() StartRecordingResponse {
	// Get audio settings to configure microphone
	audioSettings, _ := a.setting.GetAudioSettings()

	// Get paths for recording
	outputDir, _ := a.path.GetStreamDataDir()
	tempDir, _ := a.path.GetTempDataDir()

	// Update recorder config with current settings
	a.recorder.UpdateConfig(recorder.RecordingConfig{
		MicrophoneID:       audioSettings.MicrophoneID,
		SystemAudioEnabled: audioSettings.SystemAudioEnabled,
		OutputDir:          outputDir,
		TempDir:            tempDir,
	})

	if err := a.recorder.StartRecording(); err != nil {
		return StartRecordingResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	return StartRecordingResponse{
		Success: true,
		Message: "Recording started",
	}
}

// StopRecording stops recording and returns the file path
func (a *App) StopRecording() StopRecordingResponse {
	filePath, err := a.recorder.StopRecording()
	if err != nil {
		return StopRecordingResponse{
			Success: false,
			Message: err.Error(),
		}
	}
	return StopRecordingResponse{
		Success:  true,
		Message:  "Recording saved",
		FilePath: filePath,
	}
}

// GetRecordingStatus returns the current recording status
func (a *App) GetRecordingStatus() RecordingStatus {
	status := a.recorder.GetStatus()
	return RecordingStatus{
		State:    string(status.State),
		Duration: status.Duration,
		FilePath: status.FilePath,
		Error:    status.Error,
	}
}
