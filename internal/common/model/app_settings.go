package models

import (
	"time"
)

// AppSettings represents application configuration settings
type AppSettings struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:255;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Type      string    `gorm:"size:50;default:'string'" json:"type"` // string, int, bool, json
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for AppSettings
func (AppSettings) TableName() string {
	return "app_settings"
}
