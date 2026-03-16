package models

import "time"

type UploadJobStatus string

const (
	UploadJobStatusPending             UploadJobStatus = "pending"
	UploadJobStatusIniting             UploadJobStatus = "initing"
	UploadJobStatusPutting             UploadJobStatus = "putting"
	UploadJobStatusUploadedUnconfirmed UploadJobStatus = "uploaded_unconfirmed"
	UploadJobStatusRetryWait           UploadJobStatus = "retry_wait"
	UploadJobStatusBlockedAuth         UploadJobStatus = "blocked_auth"
	UploadJobStatusDone                UploadJobStatus = "done"
	UploadJobStatusMissingFile         UploadJobStatus = "missing_file"
	UploadJobStatusDead                UploadJobStatus = "dead"
)

// UploadJob stores durable upload state so uploads can continue after restart.
type UploadJob struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	SessionID          string          `gorm:"size:255;not null;index" json:"session_id"`
	FilePath           string          `gorm:"type:text;not null" json:"file_path"`
	Filename           string          `gorm:"size:512" json:"filename"`
	ContentType        string          `gorm:"size:255" json:"content_type"`
	FileSize           int64           `gorm:"default:0" json:"file_size"`
	Status             UploadJobStatus `gorm:"size:64;not null;default:'pending';index" json:"status"`
	AttemptCount       int             `gorm:"default:0" json:"attempt_count"`
	NextAttemptAt      *time.Time      `gorm:"index" json:"next_attempt_at"`
	LastAttemptAt      *time.Time      `json:"last_attempt_at"`
	LastError          string          `gorm:"type:text" json:"last_error"`
	LastHTTPStatus     int             `gorm:"default:0" json:"last_http_status"`
	UploadID           string          `gorm:"size:255" json:"upload_id"`
	SignedURL          string          `gorm:"type:text" json:"signed_url"`
	SignedURLExpiresAt *time.Time      `json:"signed_url_expires_at"`
	GCSUploadedAt      *time.Time      `json:"gcs_uploaded_at"`
	ConfirmedAt        *time.Time      `json:"confirmed_at"`
	CreatedAt          time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UploadJob) TableName() string {
	return "upload_jobs"
}
