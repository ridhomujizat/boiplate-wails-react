package app

import (
	"time"

	"onx-screen-record/internal/repository/activity"
)

type DateRange struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type TimelineEvent struct {
	ID          uint   `json:"id"`
	AppName     string `json:"appName"`
	WindowTitle string `json:"windowTitle"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Duration    int64  `json:"duration"`
	Status      string `json:"status"`
}

type TopApplication struct {
	AppName       string  `json:"appName"`
	TotalDuration int64   `json:"totalDuration"`
	SessionCount  int     `json:"sessionCount"`
	Percentage    float64 `json:"percentage"`
}

type ActivitySummary struct {
	AppName      string  `json:"appName"`
	ActiveTime   int64   `json:"activeTime"`
	AFKTime      int64   `json:"afkTime"`
	SessionCount int     `json:"sessionCount"`
	Percentage   float64 `json:"percentage"`
}

type DashboardStats struct {
	TotalActiveTime int64  `json:"totalActiveTime"`
	TotalAFKTime    int64  `json:"totalAfkTime"`
	TotalApps       int    `json:"totalApps"`
	TopApp          string `json:"topApp"`
}

func (a *App) GetActivityTimeline(dateRange DateRange) []TimelineEvent {
	startDate, err := time.Parse("2006-01-02", dateRange.StartDate)
	if err != nil {
		startDate = time.Now().Truncate(24 * time.Hour)
	}

	endDate, err := time.Parse("2006-01-02", dateRange.EndDate)
	if err != nil {
		endDate = time.Now()
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	events, err := a.rp.Activity.GetTimeline(startDate, endDate)
	if err != nil {
		return []TimelineEvent{}
	}

	result := make([]TimelineEvent, 0, len(events))
	for _, e := range events {
		result = append(result, TimelineEvent{
			ID:          e.ID,
			AppName:     e.AppName,
			WindowTitle: e.WindowTitle,
			StartTime:   e.StartTime.Format(time.RFC3339),
			EndTime:     e.EndTime.Format(time.RFC3339),
			Duration:    e.Duration,
			Status:      string(e.Status),
		})
	}
	return result
}

func (a *App) GetTopApplications(dateRange DateRange, limit int) []TopApplication {
	startDate, err := time.Parse("2006-01-02", dateRange.StartDate)
	if err != nil {
		startDate = time.Now().Truncate(24 * time.Hour)
	}

	endDate, err := time.Parse("2006-01-02", dateRange.EndDate)
	if err != nil {
		endDate = time.Now()
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	if limit <= 0 {
		limit = 10
	}

	stats, err := a.rp.Activity.GetTopApplications(startDate, endDate, limit)
	if err != nil {
		return []TopApplication{}
	}

	var totalDuration int64
	for _, s := range stats {
		totalDuration += s.TotalDuration
	}

	result := make([]TopApplication, 0, len(stats))
	for _, s := range stats {
		percentage := float64(0)
		if totalDuration > 0 {
			percentage = float64(s.TotalDuration) / float64(totalDuration) * 100
		}
		result = append(result, TopApplication{
			AppName:       s.AppName,
			TotalDuration: s.TotalDuration,
			SessionCount:  s.SessionCount,
			Percentage:    percentage,
		})
	}
	return result
}

func (a *App) GetActivitySummary(dateRange DateRange) []ActivitySummary {
	startDate, err := time.Parse("2006-01-02", dateRange.StartDate)
	if err != nil {
		startDate = time.Now().Truncate(24 * time.Hour)
	}

	endDate, err := time.Parse("2006-01-02", dateRange.EndDate)
	if err != nil {
		endDate = time.Now()
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	summaries, err := a.rp.Activity.GetActivitySummary(startDate, endDate)
	if err != nil {
		return []ActivitySummary{}
	}

	var totalTime int64
	for _, s := range summaries {
		totalTime += s.ActiveTime + s.AFKTime
	}

	result := make([]ActivitySummary, 0, len(summaries))
	for _, s := range summaries {
		percentage := float64(0)
		if totalTime > 0 {
			percentage = float64(s.ActiveTime) / float64(totalTime) * 100
		}
		result = append(result, ActivitySummary{
			AppName:      s.AppName,
			ActiveTime:   s.ActiveTime,
			AFKTime:      s.AFKTime,
			SessionCount: s.SessionCount,
			Percentage:   percentage,
		})
	}
	return result
}

func (a *App) GetDashboardStats(dateRange DateRange) DashboardStats {
	startDate, err := time.Parse("2006-01-02", dateRange.StartDate)
	if err != nil {
		startDate = time.Now().Truncate(24 * time.Hour)
	}

	endDate, err := time.Parse("2006-01-02", dateRange.EndDate)
	if err != nil {
		endDate = time.Now()
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	totalActive, _ := a.rp.Activity.GetTotalActiveTime(startDate, endDate)
	totalAFK, _ := a.rp.Activity.GetTotalAFKTime(startDate, endDate)

	topApps, _ := a.rp.Activity.GetTopApplications(startDate, endDate, 1)
	topApp := ""
	if len(topApps) > 0 {
		topApp = topApps[0].AppName
	}

	summaries, _ := a.rp.Activity.GetActivitySummary(startDate, endDate)

	return DashboardStats{
		TotalActiveTime: totalActive,
		TotalAFKTime:    totalAFK,
		TotalApps:       len(summaries),
		TopApp:          topApp,
	}
}

func (a *App) StartActivityTracking() bool {
	if a.activityTracker == nil {
		return false
	}
	a.activityTracker.Start()
	return true
}

func (a *App) StopActivityTracking() bool {
	if a.activityTracker == nil {
		return false
	}
	a.activityTracker.Stop()
	return true
}

func (a *App) IsActivityTrackingActive() bool {
	if a.activityTracker == nil {
		return false
	}
	return a.activityTracker.IsRunning()
}

func convertToTopApplication(stats []activity.AppUsageStat, totalDuration int64) []TopApplication {
	result := make([]TopApplication, 0, len(stats))
	for _, s := range stats {
		percentage := float64(0)
		if totalDuration > 0 {
			percentage = float64(s.TotalDuration) / float64(totalDuration) * 100
		}
		result = append(result, TopApplication{
			AppName:       s.AppName,
			TotalDuration: s.TotalDuration,
			SessionCount:  s.SessionCount,
			Percentage:    percentage,
		})
	}
	return result
}
