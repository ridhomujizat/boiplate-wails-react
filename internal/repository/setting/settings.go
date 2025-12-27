package setting

import (
	models "onx-screen-record/internal/common/model"

	"gorm.io/gorm"
)

// Repository handles app settings database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Get retrieves a setting by key
func (r *Repository) Get(key string) (*models.AppSettings, error) {
	var setting models.AppSettings
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetValue retrieves only the value of a setting by key
func (r *Repository) GetValue(key string) (string, error) {
	setting, err := r.Get(key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Set creates or updates a setting
func (r *Repository) Set(key, value, valueType string) error {
	var setting models.AppSettings
	result := r.db.Where("key = ?", key).First(&setting)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new setting
		return r.db.Create(&models.AppSettings{
			Key:   key,
			Value: value,
			Type:  valueType,
		}).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Update existing setting
	return r.db.Model(&setting).Updates(map[string]interface{}{
		"value": value,
		"type":  valueType,
	}).Error
}

// SetValue updates only the value of a setting
func (r *Repository) SetValue(key, value string) error {
	return r.db.Model(&models.AppSettings{}).
		Where("key = ?", key).
		Update("value", value).Error
}

// GetAll retrieves all settings
func (r *Repository) GetAll() ([]models.AppSettings, error) {
	var settings []models.AppSettings
	err := r.db.Find(&settings).Error
	return settings, err
}

// Delete removes a setting by key
func (r *Repository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.AppSettings{}).Error
}

// GetAsMap returns all settings as a map
func (r *Repository) GetAsMap() (map[string]string, error) {
	settings, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}
