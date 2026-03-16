package app

import (
	"time"

	models "onx-screen-record/internal/common/model"
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

// MQTTStatus represents the current MQTT connection status
type MQTTStatus struct {
	Connected bool   `json:"connected"`
	State     string `json:"state"` // disconnected, connecting, connected, reconnecting
	Message   string `json:"message"`
}

type UploadHistoryItem struct {
	ID                 uint   `json:"id"`
	SessionID          string `json:"sessionId"`
	FilePath           string `json:"filePath"`
	Filename           string `json:"filename"`
	ContentType        string `json:"contentType"`
	FileSize           int64  `json:"fileSize"`
	Status             string `json:"status"`
	AttemptCount       int    `json:"attemptCount"`
	LastError          string `json:"lastError"`
	LastHTTPStatus     int    `json:"lastHttpStatus"`
	UploadID           string `json:"uploadId"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	LastAttemptAt      string `json:"lastAttemptAt"`
	NextAttemptAt      string `json:"nextAttemptAt"`
	SignedURLExpiresAt string `json:"signedUrlExpiresAt"`
	GCSUploadedAt      string `json:"gcsUploadedAt"`
	ConfirmedAt        string `json:"confirmedAt"`
}

// AuthMe validates the current token and returns the active user session.
func (a *App) AuthMe(token string) interface{} {
	response, err := a.auth.AuthMe(token)
	if err != nil {
		return response
	}

	if response.Success {
		a.authToken = token
	}

	if response.StatusCode == 401 {
		a.authToken = ""
		if a.mqtt != nil && a.mqtt.IsConnected() {
			a.mqtt.Disconnect()
		}
	}

	return response
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

// StartRecording starts screen and audio recording (called from frontend).
func (a *App) StartRecording() StartRecordingResponse {
	return a.startRecordingWithSession("")
}

// startRecordingWithSession starts recording with an optional session ID for the output filename.
func (a *App) startRecordingWithSession(sessionId string) StartRecordingResponse {
	// Get audio settings to configure microphone
	audioSettings, _ := a.setting.GetAudioSettings()

	// Get recording settings for max time
	recordingSettings, _ := a.setting.GetRecordingSettings()

	// Get paths for recording
	outputDir, _ := a.path.GetStreamDataDir()
	tempDir, _ := a.path.GetTempDataDir()

	// Update recorder config with current settings
	a.recorder.UpdateConfig(recorder.RecordingConfig{
		MicrophoneID:            audioSettings.MicrophoneID,
		SystemAudioEnabled:      audioSettings.SystemAudioEnabled,
		OutputDir:               outputDir,
		TempDir:                 tempDir,
		MaxRecordingTimeEnabled: recordingSettings.MaxRecordingTimeEnabled,
		MaxRecordingTimeSeconds: recordingSettings.MaxRecordingTimeSeconds,
		SessionId:               sessionId,
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

// GetMQTTStatus returns the current MQTT connection status
func (a *App) GetMQTTStatus() MQTTStatus {
	if a.mqtt == nil {
		return MQTTStatus{
			Connected: false,
			State:     "disconnected",
			Message:   "MQTT service not initialized",
		}
	}

	state := a.mqtt.GetConnectionState()
	connected := a.mqtt.IsConnected()

	var message string
	switch state {
	case "connected":
		message = "Connected to MQTT broker"
	case "connecting":
		message = "Connecting to MQTT broker..."
	case "reconnecting":
		message = "Reconnecting to MQTT broker..."
	default:
		message = "Not connected"
	}

	return MQTTStatus{
		Connected: connected,
		State:     state,
		Message:   message,
	}
}

func (a *App) GetUploadHistory() []UploadHistoryItem {
	jobs, err := a.rp.UploadJob.ListAll()
	if err != nil {
		return []UploadHistoryItem{}
	}

	items := make([]UploadHistoryItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, mapUploadHistoryItem(job))
	}
	return items
}

func (a *App) GetActivitySettings() dtoSetting.ActivitySettingResponse {
	result, err := a.setting.GetActivitySettings()
	if err != nil {
		return dtoSetting.ActivitySettingResponse{
			PollingInterval: 5,
			AFKThreshold:    180,
		}
	}
	return *result
}

func (a *App) SaveActivitySettings(req dtoSetting.ActivitySettingRequest) SaveSettingsResponse {
	result, err := a.setting.SaveActivitySettings(req)
	if err != nil {
		return SaveSettingsResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	if a.activityTracker != nil {
		a.activityTracker.UpdateConfig(req.PollingInterval, req.AFKThreshold)
	}

	return SaveSettingsResponse{
		Success: result.Success,
		Message: result.Message,
	}
}

func (a *App) GetRecordingSettings() dtoSetting.RecordingSettingResponse {
	result, err := a.setting.GetRecordingSettings()
	if err != nil {
		return dtoSetting.RecordingSettingResponse{
			MaxRecordingTimeEnabled: false,
			MaxRecordingTimeSeconds: 3600,
		}
	}
	return *result
}

func (a *App) SaveRecordingSettings(req dtoSetting.RecordingSettingRequest) SaveSettingsResponse {
	result, err := a.setting.SaveRecordingSettings(req)
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

func (a *App) GetUploadSettings() dtoSetting.UploadSettingResponse {
	result, err := a.setting.GetUploadSettings()
	if err != nil {
		return dtoSetting.UploadSettingResponse{
			DeleteAfterUpload: false,
		}
	}
	return *result
}

func (a *App) SaveUploadSettings(req dtoSetting.UploadSettingRequest) SaveSettingsResponse {
	result, err := a.setting.SaveUploadSettings(req)
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

func mapUploadHistoryItem(job models.UploadJob) UploadHistoryItem {
	return UploadHistoryItem{
		ID:                 job.ID,
		SessionID:          job.SessionID,
		FilePath:           job.FilePath,
		Filename:           job.Filename,
		ContentType:        job.ContentType,
		FileSize:           job.FileSize,
		Status:             string(job.Status),
		AttemptCount:       job.AttemptCount,
		LastError:          job.LastError,
		LastHTTPStatus:     job.LastHTTPStatus,
		UploadID:           job.UploadID,
		CreatedAt:          formatUploadHistoryTime(job.CreatedAt),
		UpdatedAt:          formatUploadHistoryTime(job.UpdatedAt),
		LastAttemptAt:      formatUploadHistoryTimePtr(job.LastAttemptAt),
		NextAttemptAt:      formatUploadHistoryTimePtr(job.NextAttemptAt),
		SignedURLExpiresAt: formatUploadHistoryTimePtr(job.SignedURLExpiresAt),
		GCSUploadedAt:      formatUploadHistoryTimePtr(job.GCSUploadedAt),
		ConfirmedAt:        formatUploadHistoryTimePtr(job.ConfirmedAt),
	}
}

func formatUploadHistoryTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatUploadHistoryTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatUploadHistoryTime(*t)
}
