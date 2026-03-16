package uploadjob

import (
	"time"

	models "onx-screen-record/internal/common/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(job *models.UploadJob) error {
	return r.db.Create(job).Error
}

func (r *Repository) Save(job *models.UploadJob) error {
	return r.db.Save(job).Error
}

func (r *Repository) GetByID(id uint) (*models.UploadJob, error) {
	var job models.UploadJob
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Repository) FindBySessionAndPath(sessionID, filePath string) (*models.UploadJob, error) {
	var job models.UploadJob
	err := r.db.Where("session_id = ? AND file_path = ?", sessionID, filePath).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Repository) GetNextReady(now time.Time) (*models.UploadJob, error) {
	var job models.UploadJob
	err := r.db.
		Where("status IN ?", []models.UploadJobStatus{
			models.UploadJobStatusPending,
			models.UploadJobStatusRetryWait,
			models.UploadJobStatusBlockedAuth,
			models.UploadJobStatusUploadedUnconfirmed,
			models.UploadJobStatusIniting,
			models.UploadJobStatusPutting,
		}).
		Where("next_attempt_at IS NULL OR next_attempt_at <= ?", now).
		Order("CASE status WHEN 'uploaded_unconfirmed' THEN 0 ELSE 1 END, created_at ASC").
		First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Repository) ResetInFlightJobs(now time.Time) error {
	return r.db.Model(&models.UploadJob{}).
		Where("status IN ?", []models.UploadJobStatus{
			models.UploadJobStatusIniting,
			models.UploadJobStatusPutting,
		}).
		Updates(map[string]interface{}{
			"status":          models.UploadJobStatusRetryWait,
			"next_attempt_at": now,
			"last_error":      "Upload resumed after app restart",
		}).Error
}

func (r *Repository) MarkBlockedAuthReady(now time.Time) error {
	return r.db.Model(&models.UploadJob{}).
		Where("status = ?", models.UploadJobStatusBlockedAuth).
		Update("next_attempt_at", now).Error
}

func (r *Repository) ListAll() ([]models.UploadJob, error) {
	var jobs []models.UploadJob
	err := r.db.Order("created_at DESC").Order("id DESC").Find(&jobs).Error
	return jobs, err
}
