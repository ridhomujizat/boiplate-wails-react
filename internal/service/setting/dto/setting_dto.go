package dto

// SettingRequest represents the request body for saving settings
type SettingRequest struct {
	TenantCode string `json:"tenantCode"`
	BaseUrl    string `json:"baseUrl"`
	MqttBroker string `json:"mqttBroker"`
}

// SettingResponse represents the response body for getting settings
type SettingResponse struct {
	TenantCode string `json:"tenantCode"`
	BaseUrl    string `json:"baseUrl"`
	MqttBroker string `json:"mqttBroker"`
}

// SaveSettingResponse represents the response after saving settings
type SaveSettingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AudioSettingRequest represents the request body for saving audio settings
type AudioSettingRequest struct {
	MicrophoneID       string `json:"microphoneId"`
	SystemAudioEnabled bool   `json:"systemAudioEnabled"`
}

// AudioSettingResponse represents the response body for getting audio settings
type AudioSettingResponse struct {
	MicrophoneID       string `json:"microphoneId"`
	SystemAudioEnabled bool   `json:"systemAudioEnabled"`
}
