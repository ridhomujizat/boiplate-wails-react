package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"onx-screen-record/internal/common/enum"
	models "onx-screen-record/internal/common/model"
	"onx-screen-record/internal/pkg/helper"
	"onx-screen-record/internal/pkg/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

var uploadRetrySchedule = []time.Duration{
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	6 * time.Hour,
}

const blockedAuthRetryDelay = 24 * time.Hour

func (a *App) startUploadWorker() {
	if a.uploadWake == nil {
		a.uploadWake = make(chan struct{}, 1)
	}
	if a.uploadStop == nil {
		a.uploadStop = make(chan struct{})
	}

	now := time.Now()
	if err := a.rp.UploadJob.ResetInFlightJobs(now); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to reset in-flight jobs: %v", err)
	}

	go a.uploadWorkerLoop()
	a.wakeUploadWorker()
}

func (a *App) stopUploadWorker() {
	a.uploadStopOnce.Do(func() {
		if a.uploadStop != nil {
			close(a.uploadStop)
		}
	})
}

func (a *App) wakeUploadWorker() {
	if a.uploadWake == nil {
		return
	}
	select {
	case a.uploadWake <- struct{}{}:
	default:
	}
}

func (a *App) resumeBlockedUploads() {
	if err := a.rp.UploadJob.MarkBlockedAuthReady(time.Now()); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to resume blocked uploads: %v", err)
		return
	}
	a.wakeUploadWorker()
}

func (a *App) uploadWorkerLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		a.processReadyUploadJobs()

		select {
		case <-ticker.C:
		case <-a.uploadWake:
		case <-a.uploadStop:
			return
		}
	}
}

func (a *App) processReadyUploadJobs() {
	for {
		job, err := a.rp.UploadJob.GetNextReady(time.Now())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		if err != nil {
			logger.Error.Printf("[UploadQueue] Failed to fetch ready job: %v", err)
			return
		}

		a.processUploadJob(job)
	}
}

func (a *App) enqueueUploadJob(filePath, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("session ID is required for upload")
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	now := time.Now()
	existing, err := a.rp.UploadJob.FindBySessionAndPath(sessionID, filePath)
	if err == nil {
		existing.Status = models.UploadJobStatusPending
		existing.NextAttemptAt = &now
		existing.LastError = ""
		existing.LastHTTPStatus = 0
		existing.GCSUploadedAt = nil
		existing.ConfirmedAt = nil
		existing.UploadID = ""
		existing.SignedURL = ""
		existing.SignedURLExpiresAt = nil
		existing.Filename = filepath.Base(filePath)
		existing.ContentType = detectUploadContentType(filePath)
		existing.FileSize = fileInfo.Size()
		if saveErr := a.rp.UploadJob.Save(existing); saveErr != nil {
			return saveErr
		}
		a.emitUploadJobEvent(existing, "queued", nil, "")
		a.wakeUploadWorker()
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	job := &models.UploadJob{
		SessionID:     sessionID,
		FilePath:      filePath,
		Filename:      filepath.Base(filePath),
		ContentType:   detectUploadContentType(filePath),
		FileSize:      fileInfo.Size(),
		Status:        models.UploadJobStatusPending,
		NextAttemptAt: &now,
	}
	if err := a.rp.UploadJob.Create(job); err != nil {
		return err
	}

	a.emitUploadJobEvent(job, "queued", nil, "")
	a.wakeUploadWorker()
	return nil
}

func (a *App) processUploadJob(job *models.UploadJob) {
	now := time.Now()
	job.AttemptCount++
	job.LastAttemptAt = &now
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to update attempt count for job %d: %v", job.ID, err)
		return
	}

	if err := a.refreshUploadJobMetadata(job); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			a.markUploadMissingFile(job, err.Error())
			return
		}
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return
	}

	if a.authToken == "" {
		a.markUploadBlockedAuth(job, "Auth token is not available")
		return
	}

	settings, err := a.setting.GetSettings()
	if err != nil {
		a.scheduleUploadRetry(job, "Failed to load settings", 0, nil, false)
		return
	}
	if strings.TrimSpace(settings.BaseUrl) == "" {
		a.scheduleUploadRetry(job, "BaseUrl is not configured", 0, nil, false)
		return
	}

	if job.Status == models.UploadJobStatusUploadedUnconfirmed && job.UploadID != "" {
		a.emitUploadJobEvent(job, "confirming", nil, "")
		a.completeUploadJob(job, settings.BaseUrl)
		return
	}

	if job.UploadID == "" || job.SignedURL == "" || signedURLExpired(job) {
		if !a.initUploadJob(job, settings.BaseUrl) {
			return
		}
	}

	if !a.putUploadJob(job) {
		return
	}

	a.completeUploadJob(job, settings.BaseUrl)
}

func (a *App) refreshUploadJobMetadata(job *models.UploadJob) error {
	info, err := os.Stat(job.FilePath)
	if err != nil {
		return err
	}

	job.Filename = filepath.Base(job.FilePath)
	job.ContentType = detectUploadContentType(job.FilePath)
	job.FileSize = info.Size()
	return a.rp.UploadJob.Save(job)
}

func (a *App) initUploadJob(job *models.UploadJob, baseURL string) bool {
	job.Status = models.UploadJobStatusIniting
	job.NextAttemptAt = nil
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to mark job %d as initing: %v", job.ID, err)
		return false
	}

	a.emitUploadJobEvent(job, "initing", nil, "")

	initBody := map[string]interface{}{
		"session_id":   job.SessionID,
		"filename":     job.Filename,
		"content_type": job.ContentType,
		"file_size":    job.FileSize,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	initURL := fmt.Sprintf("%s/api/record/upload/init", baseURL)
	resp, err := helper.HTTPRequest(
		&helper.HTTPRequestPayload{
			Method: enum.POST,
			URL:    initURL,
			Body:   initBody,
		},
		&helper.HTTPRequestConfig{
			Ctx: ctx,
			Headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + a.authToken},
			},
		},
	)
	if err != nil {
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return false
	}

	if resp.StatusCode == http.StatusUnauthorized {
		a.markUploadBlockedAuth(job, getResponseMessage(resp.Data))
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		a.scheduleUploadRetry(job, getResponseMessage(resp.Data), resp.StatusCode, resp.Data, false)
		return false
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.markUploadDead(job, "Init upload failed", resp.StatusCode, resp.Data)
		return false
	}

	uploadURL, uploadID, expiresAt, err := parseInitUploadResponse(resp.Data)
	if err != nil {
		a.markUploadDead(job, err.Error(), resp.StatusCode, resp.Data)
		return false
	}

	job.UploadID = uploadID
	job.SignedURL = uploadURL
	job.SignedURLExpiresAt = expiresAt
	job.LastError = ""
	job.LastHTTPStatus = 0
	job.Status = models.UploadJobStatusPending
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to persist init result for job %d: %v", job.ID, err)
		return false
	}

	a.emitUploadJobEvent(job, "init_ok", resp.Data, "")
	return true
}

func (a *App) putUploadJob(job *models.UploadJob) bool {
	job.Status = models.UploadJobStatusPutting
	job.NextAttemptAt = nil
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to mark job %d as putting: %v", job.ID, err)
		return false
	}

	a.emitUploadJobEvent(job, "uploading", nil, "")

	file, err := os.Open(job.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			a.markUploadMissingFile(job, err.Error())
			return false
		}
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return false
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, job.SignedURL, file)
	if err != nil {
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return false
	}
	req.Header.Set("Content-Type", job.ContentType)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return false
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyText := strings.TrimSpace(string(bodyBytes))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now()
		job.Status = models.UploadJobStatusUploadedUnconfirmed
		job.GCSUploadedAt = &now
		job.LastError = ""
		job.LastHTTPStatus = 0
		if err := a.rp.UploadJob.Save(job); err != nil {
			logger.Error.Printf("[UploadQueue] Failed to persist successful PUT for job %d: %v", job.ID, err)
			return false
		}
		a.emitUploadJobEvent(job, "uploaded_unconfirmed", bodyText, "")
		return true
	}

	if resp.StatusCode == http.StatusForbidden {
		job.UploadID = ""
		job.SignedURL = ""
		job.SignedURLExpiresAt = nil
		job.GCSUploadedAt = nil
		a.scheduleUploadRetry(job, "Signed URL rejected by storage", resp.StatusCode, bodyText, false)
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		a.scheduleUploadRetry(job, bodyText, resp.StatusCode, bodyText, false)
		return false
	}

	a.markUploadDead(job, "GCS upload failed", resp.StatusCode, bodyText)
	return false
}

func (a *App) completeUploadJob(job *models.UploadJob, baseURL string) {
	doneBody := map[string]interface{}{
		"session_id": job.SessionID,
		"upload_id":  job.UploadID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	doneURL := fmt.Sprintf("%s/api/record/upload/done", baseURL)
	resp, err := helper.HTTPRequest(
		&helper.HTTPRequestPayload{
			Method: enum.POST,
			URL:    doneURL,
			Body:   doneBody,
		},
		&helper.HTTPRequestConfig{
			Ctx: ctx,
			Headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + a.authToken},
			},
		},
	)
	if err != nil {
		a.scheduleUploadRetry(job, err.Error(), 0, nil, false)
		return
	}

	message := getResponseMessage(resp.Data)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now()
		job.Status = models.UploadJobStatusDone
		job.ConfirmedAt = &now
		job.NextAttemptAt = nil
		job.LastError = ""
		job.LastHTTPStatus = 0
		if err := a.rp.UploadJob.Save(job); err != nil {
			logger.Error.Printf("[UploadQueue] Failed to persist completed upload for job %d: %v", job.ID, err)
			return
		}

		uploadSettings, settingsErr := a.setting.GetUploadSettings()
		if settingsErr == nil && uploadSettings.DeleteAfterUpload {
			if removeErr := os.Remove(job.FilePath); removeErr != nil {
				logger.Error.Printf("[UploadQueue] Failed to delete local file for job %d: %v", job.ID, removeErr)
			}
		}

		a.emitUploadJobEvent(job, "success", resp.Data, "")
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		a.markUploadBlockedAuth(job, message)
		return
	}
	if resp.StatusCode == http.StatusBadRequest {
		lowerMessage := strings.ToLower(message)
		switch {
		case strings.Contains(lowerMessage, "file not found in storage"):
			a.scheduleUploadRetry(job, message, resp.StatusCode, resp.Data, true)
			return
		case strings.Contains(lowerMessage, "upload_id does not match"):
			job.UploadID = ""
			job.SignedURL = ""
			job.SignedURLExpiresAt = nil
			job.GCSUploadedAt = nil
			a.scheduleUploadRetry(job, message, resp.StatusCode, resp.Data, false)
			return
		}
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		a.scheduleUploadRetry(job, message, resp.StatusCode, resp.Data, false)
		return
	}

	a.markUploadDead(job, "Upload confirmation failed", resp.StatusCode, resp.Data)
}

func (a *App) scheduleUploadRetry(job *models.UploadJob, reason string, statusCode int, response interface{}, fast bool) {
	now := time.Now()
	delay := nextUploadRetryDelay(job.AttemptCount, fast)
	nextAttempt := now.Add(delay)

	job.Status = models.UploadJobStatusRetryWait
	job.NextAttemptAt = &nextAttempt
	job.LastError = fallbackMessage(reason, "Upload will be retried")
	job.LastHTTPStatus = statusCode
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to schedule retry for job %d: %v", job.ID, err)
		return
	}

	a.emitUploadJobEvent(job, "retrying", response, job.LastError)
}

func (a *App) markUploadBlockedAuth(job *models.UploadJob, reason string) {
	nextAttempt := time.Now().Add(blockedAuthRetryDelay)
	job.Status = models.UploadJobStatusBlockedAuth
	job.NextAttemptAt = &nextAttempt
	job.LastError = fallbackMessage(reason, "Upload waiting for authentication")
	job.LastHTTPStatus = http.StatusUnauthorized
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to block job %d on auth: %v", job.ID, err)
		return
	}

	a.emitUploadJobEvent(job, "blocked_auth", nil, job.LastError)
}

func (a *App) markUploadMissingFile(job *models.UploadJob, reason string) {
	job.Status = models.UploadJobStatusMissingFile
	job.NextAttemptAt = nil
	job.LastError = fallbackMessage(reason, "Local file not found")
	job.LastHTTPStatus = 0
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to mark missing file for job %d: %v", job.ID, err)
		return
	}

	a.emitUploadJobEvent(job, "failed_permanent", nil, job.LastError)
}

func (a *App) markUploadDead(job *models.UploadJob, reason string, statusCode int, response interface{}) {
	job.Status = models.UploadJobStatusDead
	job.NextAttemptAt = nil
	job.LastError = fallbackMessage(reason, "Upload failed permanently")
	job.LastHTTPStatus = statusCode
	if err := a.rp.UploadJob.Save(job); err != nil {
		logger.Error.Printf("[UploadQueue] Failed to mark dead job %d: %v", job.ID, err)
		return
	}

	a.emitUploadJobEvent(job, "failed_permanent", response, job.LastError)
}

func (a *App) emitUploadJobEvent(job *models.UploadJob, phase string, response interface{}, errMsg string) {
	if a.ctx == nil || job == nil {
		return
	}

	event := map[string]interface{}{
		"success":        job.Status == models.UploadJobStatusDone,
		"status":         string(job.Status),
		"phase":          phase,
		"jobId":          job.ID,
		"filePath":       job.FilePath,
		"sessionId":      job.SessionID,
		"attemptCount":   job.AttemptCount,
		"lastHttpStatus": job.LastHTTPStatus,
	}
	if job.NextAttemptAt != nil {
		event["nextAttemptAt"] = job.NextAttemptAt.Format(time.RFC3339)
	}
	if response != nil {
		event["response"] = response
	}
	if errMsg != "" {
		event["error"] = errMsg
	}

	runtime.EventsEmit(a.ctx, "recording-uploaded", event)
}

func detectUploadContentType(filePath string) string {
	if strings.HasSuffix(strings.ToLower(filePath), ".mp4") {
		return "video/mp4"
	}
	return "video/webm"
}

func signedURLExpired(job *models.UploadJob) bool {
	if job.SignedURL == "" {
		return true
	}
	if job.SignedURLExpiresAt == nil {
		return false
	}
	return time.Now().After(job.SignedURLExpiresAt.Add(-1 * time.Minute))
}

func parseInitUploadResponse(data interface{}) (string, string, *time.Time, error) {
	respMap, ok := data.(map[string]interface{})
	if !ok {
		return "", "", nil, fmt.Errorf("unexpected init response format")
	}

	nestedData, ok := respMap["data"].(map[string]interface{})
	if !ok {
		return "", "", nil, fmt.Errorf("missing data in init response")
	}

	uploadURL, _ := nestedData["upload_url"].(string)
	uploadID, _ := nestedData["upload_id"].(string)
	if uploadURL == "" || uploadID == "" {
		return "", "", nil, fmt.Errorf("missing upload_url or upload_id")
	}

	var expiresAt *time.Time
	if rawExpiresAt, ok := nestedData["expires_at"].(string); ok && strings.TrimSpace(rawExpiresAt) != "" {
		if parsed, err := time.Parse(time.RFC3339, rawExpiresAt); err == nil {
			expiresAt = &parsed
		}
	}

	return uploadURL, uploadID, expiresAt, nil
}

func getResponseMessage(data interface{}) string {
	switch value := data.(type) {
	case map[string]interface{}:
		if message, ok := value["message"].(string); ok && strings.TrimSpace(message) != "" {
			return message
		}
		if nestedMessage, ok := value["error"].(string); ok && strings.TrimSpace(nestedMessage) != "" {
			return nestedMessage
		}
	case string:
		return value
	case []byte:
		return string(value)
	}

	serialized, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(serialized)
}

func fallbackMessage(primary, fallback string) string {
	primary = strings.TrimSpace(primary)
	if primary != "" && primary != "null" {
		return primary
	}
	return fallback
}

func nextUploadRetryDelay(attempt int, fast bool) time.Duration {
	if fast {
		base := 10 * time.Second
		if attempt > 3 {
			base = 30 * time.Second
		}
		return addRetryJitter(base)
	}

	if attempt < 1 {
		attempt = 1
	}
	index := attempt - 1
	if index >= len(uploadRetrySchedule) {
		index = len(uploadRetrySchedule) - 1
	}
	return addRetryJitter(uploadRetrySchedule[index])
}

func addRetryJitter(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}

	jitterWindow := int64(base / 5)
	if jitterWindow <= 0 {
		return base
	}

	offset := rand.Int63n((jitterWindow * 2) + 1)
	return base - time.Duration(jitterWindow) + time.Duration(offset)
}
