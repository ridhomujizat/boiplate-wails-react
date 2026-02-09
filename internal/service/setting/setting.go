package setting

import (
	"strconv"

	"onx-screen-record/internal/service/setting/dto"
)

const (
	KeyTenant                  = "tenant"
	KeyBaseUrl                 = "baseurl"
	KeyMqtt                    = "mqtt"
	KeyMicrophone              = "microphone"
	KeySystemAudio             = "systemaudio"
	KeyActivityPollingInterval = "activity_polling_interval"
	KeyActivityAFKThreshold    = "activity_afk_threshold"
	KeyMaxRecordingTimeEnabled = "max_recording_time_enabled"
	KeyMaxRecordingTimeSeconds = "max_recording_time_seconds"
)

// GetSettings retrieves all settings from the database
func (s *Service) GetSettings() (*dto.SettingResponse, error) {
	settingsMap, err := s.rp.Setting.GetAsMap()
	if err != nil {
		return nil, err
	}

	return &dto.SettingResponse{
		TenantCode: settingsMap[KeyTenant],
		BaseUrl:    settingsMap[KeyBaseUrl],
		MqttBroker: settingsMap[KeyMqtt],
	}, nil
}

// SaveSettings saves settings to the database
func (s *Service) SaveSettings(req dto.SettingRequest) (*dto.SaveSettingResponse, error) {
	// Save tenant code
	if err := s.rp.Setting.SetValue(KeyTenant, req.TenantCode); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save tenant code: " + err.Error(),
		}, err
	}

	// Save base URL
	if err := s.rp.Setting.SetValue(KeyBaseUrl, req.BaseUrl); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save base URL: " + err.Error(),
		}, err
	}

	// Save MQTT broker
	if err := s.rp.Setting.SetValue(KeyMqtt, req.MqttBroker); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save MQTT broker: " + err.Error(),
		}, err
	}

	return &dto.SaveSettingResponse{
		Success: true,
		Message: "Settings saved successfully",
	}, nil
}

// GetAudioSettings retrieves audio settings from the database
func (s *Service) GetAudioSettings() (*dto.AudioSettingResponse, error) {
	settingsMap, err := s.rp.Setting.GetAsMap()
	if err != nil {
		return nil, err
	}

	systemAudioEnabled := settingsMap[KeySystemAudio] == "true"

	return &dto.AudioSettingResponse{
		MicrophoneID:       settingsMap[KeyMicrophone],
		SystemAudioEnabled: systemAudioEnabled,
	}, nil
}

// SaveAudioSettings saves audio settings to the database
func (s *Service) SaveAudioSettings(req dto.AudioSettingRequest) (*dto.SaveSettingResponse, error) {
	// Save microphone ID
	if err := s.rp.Setting.SetValue(KeyMicrophone, req.MicrophoneID); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save microphone: " + err.Error(),
		}, err
	}

	// Save system audio enabled status
	systemAudioValue := "false"
	if req.SystemAudioEnabled {
		systemAudioValue = "true"
	}
	if err := s.rp.Setting.SetValue(KeySystemAudio, systemAudioValue); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save system audio setting: " + err.Error(),
		}, err
	}

	return &dto.SaveSettingResponse{
		Success: true,
		Message: "Audio settings saved successfully",
	}, nil
}

func (s *Service) GetActivitySettings() (*dto.ActivitySettingResponse, error) {
	settingsMap, err := s.rp.Setting.GetAsMap()
	if err != nil {
		return nil, err
	}

	pollingInterval := 5
	if v := settingsMap[KeyActivityPollingInterval]; v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			pollingInterval = parsed
		}
	}

	afkThreshold := 180
	if v := settingsMap[KeyActivityAFKThreshold]; v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			afkThreshold = parsed
		}
	}

	return &dto.ActivitySettingResponse{
		PollingInterval: pollingInterval,
		AFKThreshold:    afkThreshold,
	}, nil
}

func (s *Service) SaveActivitySettings(req dto.ActivitySettingRequest) (*dto.SaveSettingResponse, error) {
	if req.PollingInterval < 1 {
		req.PollingInterval = 5
	}
	if req.AFKThreshold < 10 {
		req.AFKThreshold = 180
	}

	if err := s.rp.Setting.Set(KeyActivityPollingInterval, strconv.Itoa(req.PollingInterval), "int"); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save polling interval: " + err.Error(),
		}, err
	}

	if err := s.rp.Setting.Set(KeyActivityAFKThreshold, strconv.Itoa(req.AFKThreshold), "int"); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save AFK threshold: " + err.Error(),
		}, err
	}

	return &dto.SaveSettingResponse{
		Success: true,
		Message: "Activity settings saved successfully",
	}, nil
}

func (s *Service) GetRecordingSettings() (*dto.RecordingSettingResponse, error) {
	settingsMap, err := s.rp.Setting.GetAsMap()
	if err != nil {
		return nil, err
	}

	enabled := false
	if v := settingsMap[KeyMaxRecordingTimeEnabled]; v == "true" {
		enabled = true
	}

	seconds := 3600 // default 1 hour
	if v := settingsMap[KeyMaxRecordingTimeSeconds]; v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 30 && parsed <= 36000 {
			seconds = parsed
		}
	}

	return &dto.RecordingSettingResponse{
		MaxRecordingTimeEnabled: enabled,
		MaxRecordingTimeSeconds: seconds,
	}, nil
}

func (s *Service) SaveRecordingSettings(req dto.RecordingSettingRequest) (*dto.SaveSettingResponse, error) {
	// Validate seconds range
	if req.MaxRecordingTimeSeconds < 30 {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Maximum recording time must be at least 30 seconds",
		}, nil
	}
	if req.MaxRecordingTimeSeconds > 36000 {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Maximum recording time cannot exceed 36000 seconds (10 hours)",
		}, nil
	}

	// Save enabled status
	enabledValue := "false"
	if req.MaxRecordingTimeEnabled {
		enabledValue = "true"
	}
	if err := s.rp.Setting.Set(KeyMaxRecordingTimeEnabled, enabledValue, "bool"); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save max recording time enabled: " + err.Error(),
		}, err
	}

	// Save seconds
	if err := s.rp.Setting.Set(KeyMaxRecordingTimeSeconds, strconv.Itoa(req.MaxRecordingTimeSeconds), "int"); err != nil {
		return &dto.SaveSettingResponse{
			Success: false,
			Message: "Failed to save max recording time seconds: " + err.Error(),
		}, err
	}

	return &dto.SaveSettingResponse{
		Success: true,
		Message: "Recording settings saved successfully",
	}, nil
}
