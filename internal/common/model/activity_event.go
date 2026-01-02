package models

import (
	"time"
)

// ActivityStatus represents whether the user was active or AFK
type ActivityStatus string

const (
	StatusActive ActivityStatus = "active"
	StatusAFK    ActivityStatus = "afk"
)

// ActivityEvent represents a single activity event with window/app information
type ActivityEvent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AppName     string         `gorm:"size:255;not null;index" json:"app_name"`
	BundleID    string         `gorm:"size:255" json:"bundle_id"` // macOS bundle ID or Windows process name
	WindowTitle string         `gorm:"size:512" json:"window_title"`
	TabTitle    string         `gorm:"size:512" json:"tab_title"` // Browser tab title if applicable
	URL         string         `gorm:"size:1024" json:"url"`      // Optional URL for browser tabs
	StartTime   time.Time      `gorm:"not null;index" json:"start_time"`
	EndTime     time.Time      `gorm:"index" json:"end_time"`
	Duration    int64          `gorm:"default:0" json:"duration"` // Duration in seconds
	Status      ActivityStatus `gorm:"size:20;default:'active'" json:"status"`
	OS          string         `gorm:"size:20" json:"os"` // darwin, windows, linux
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for ActivityEvent
func (ActivityEvent) TableName() string {
	return "activity_events"
}

// AggregatedStats represents aggregated activity statistics per app per day
type AggregatedStats struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AppName         string    `gorm:"size:255;not null;index:idx_app_date" json:"app_name"`
	Date            time.Time `gorm:"type:date;not null;index:idx_app_date" json:"date"`
	TotalActiveTime int64     `gorm:"default:0" json:"total_active_time"` // Total active time in seconds
	TotalAFKTime    int64     `gorm:"default:0" json:"total_afk_time"`    // Total AFK time in seconds
	SessionCount    int       `gorm:"default:0" json:"session_count"`     // Number of sessions
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for AggregatedStats
func (AggregatedStats) TableName() string {
	return "aggregated_stats"
}
