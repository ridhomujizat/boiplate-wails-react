package activity

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

func (r *Repository) Create(event *models.ActivityEvent) error {
	return r.db.Create(event).Error
}

func (r *Repository) Update(event *models.ActivityEvent) error {
	return r.db.Save(event).Error
}

func (r *Repository) GetByID(id uint) (*models.ActivityEvent, error) {
	var event models.ActivityEvent
	err := r.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *Repository) GetLastEvent() (*models.ActivityEvent, error) {
	var event models.ActivityEvent
	err := r.db.Order("id DESC").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *Repository) GetTimeline(startDate, endDate time.Time) ([]models.ActivityEvent, error) {
	var events []models.ActivityEvent
	err := r.db.Where("start_time >= ? AND start_time <= ?", startDate, endDate).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}

func (r *Repository) GetTopApplications(startDate, endDate time.Time, limit int) ([]AppUsageStat, error) {
	var stats []AppUsageStat
	err := r.db.Model(&models.ActivityEvent{}).
		Select("app_name, SUM(duration) as total_duration, COUNT(*) as session_count").
		Where("start_time >= ? AND start_time <= ? AND status = ?", startDate, endDate, models.StatusActive).
		Group("app_name").
		Order("total_duration DESC").
		Limit(limit).
		Scan(&stats).Error
	return stats, err
}

func (r *Repository) GetActivitySummary(startDate, endDate time.Time) ([]AppActivitySummary, error) {
	var summaries []AppActivitySummary
	err := r.db.Model(&models.ActivityEvent{}).
		Select(`
			app_name,
			SUM(CASE WHEN status = 'active' THEN duration ELSE 0 END) as active_time,
			SUM(CASE WHEN status = 'afk' THEN duration ELSE 0 END) as afk_time,
			COUNT(*) as session_count
		`).
		Where("start_time >= ? AND start_time <= ?", startDate, endDate).
		Group("app_name").
		Order("active_time DESC").
		Scan(&summaries).Error
	return summaries, err
}

func (r *Repository) GetTotalActiveTime(startDate, endDate time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&models.ActivityEvent{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("start_time >= ? AND start_time <= ? AND status = ?", startDate, endDate, models.StatusActive).
		Scan(&total).Error
	return total, err
}

func (r *Repository) GetTotalAFKTime(startDate, endDate time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&models.ActivityEvent{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("start_time >= ? AND start_time <= ? AND status = ?", startDate, endDate, models.StatusAFK).
		Scan(&total).Error
	return total, err
}

type AppUsageStat struct {
	AppName       string `json:"app_name"`
	TotalDuration int64  `json:"total_duration"`
	SessionCount  int    `json:"session_count"`
}

type AppActivitySummary struct {
	AppName      string `json:"app_name"`
	ActiveTime   int64  `json:"active_time"`
	AFKTime      int64  `json:"afk_time"`
	SessionCount int    `json:"session_count"`
}
