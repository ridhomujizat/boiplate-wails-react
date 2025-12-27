package setting

import (
	"onx-screen-record/internal/service/setting/dto"
)

const (
	KeyTenant  = "tenant"
	KeyBaseUrl = "baseurl"
	KeyMqtt    = "mqtt"
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
