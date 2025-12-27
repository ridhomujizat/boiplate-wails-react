package app

import (
	dtoSetting "onx-screen-record/internal/service/setting/dto"
)

// SaveSettingsResponse is the response type after saving settings (exported for Wails)
type SaveSettingsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
