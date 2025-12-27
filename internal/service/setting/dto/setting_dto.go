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
