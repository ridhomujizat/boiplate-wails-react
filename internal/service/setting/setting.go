package setting

import (
	"onx-screen-record/internal/service/setting/dto"
)

const (
	KeyTenant      = "tenant"
	KeyBaseUrl     = "baseurl"
	KeyMqtt        = "mqtt"
	KeyMicrophone  = "microphone"
	KeySystemAudio = "systemaudio"
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
