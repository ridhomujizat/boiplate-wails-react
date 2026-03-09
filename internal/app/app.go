package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"onx-screen-record/internal/common/enum"
	types "onx-screen-record/internal/common/type"
	"onx-screen-record/internal/pkg/helper"
	"onx-screen-record/internal/pkg/logger"
	pathHelper "onx-screen-record/internal/pkg/path-file"
	"onx-screen-record/internal/pkg/recorder"
	"onx-screen-record/internal/pkg/tray"
	"onx-screen-record/internal/repository"
	activityService "onx-screen-record/internal/service/activity"
	"onx-screen-record/internal/service/auth"
	"onx-screen-record/internal/service/integration"
	mqttService "onx-screen-record/internal/service/mqtt"
	"onx-screen-record/internal/service/setting"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	AppName string
	ctx     context.Context
	path    *pathHelper.PathHelper

	trayManager *tray.TrayManager
	httpServer  *integration.Server
	isVisible   bool

	rp repository.IRepository

	setting         setting.IService
	auth            auth.IService
	mqtt            mqttService.IService
	recorder        *recorder.RecorderManager
	activityTracker *activityService.Tracker

	initialDeepLink string // Store initial deep link URL
	authToken       string // Store auth token for API calls
}

func NewApp() *App {
	return &App{
		AppName: "onx-screen-record",
	}
}

func (a *App) Startup(ctx context.Context) {
	logger.Setup()
	a.ctx = ctx

	a.path = pathHelper.NewPathHelper(a.AppName)

	if err := a.initializeDatabase(); err != nil {
		logger.Error.Printf("Failed to initialize database: %v", err)
		runtime.Quit(ctx)
		return
	}

	a.setupSystemTray()

	a.startHTTPServer()

	a.setting = setting.NewService(a.ctx, a.rp)

	// Initialize auth service with baseURL getter
	a.auth = auth.NewService(a.ctx, func() (string, error) {
		settings, err := a.setting.GetSettings()
		if err != nil {
			return "", err
		}
		return settings.BaseUrl, nil
	})

	// Initialize MQTT service (do not connect yet)
	a.mqtt = mqttService.NewService(a.ctx, a.auth, a.rp, a.setting)
	logger.Info.Printf("MQTT service initialized")

	outputDir, _ := a.path.GetStreamDataDir()
	tempDir, _ := a.path.GetTempDataDir()
	a.recorder = recorder.NewRecorderManager(recorder.RecordingConfig{
		OutputDir: outputDir,
		TempDir:   tempDir,
	})

	activitySettings, _ := a.setting.GetActivitySettings()
	a.activityTracker = activityService.NewTracker(&a.rp, activityService.TrackerConfig{
		PollingInterval: time.Duration(activitySettings.PollingInterval) * time.Second,
		AFKThreshold:    time.Duration(activitySettings.AFKThreshold) * time.Second,
		Enabled:         true,
	})
	// a.activityTracker.Start()

	// Process initial deep link if present
	if a.initialDeepLink != "" {
		a.HandleDeepLink(a.initialDeepLink)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Login performs user authentication via API
func (a *App) Login(email string, password string) interface{} {
	response, err := a.auth.Login(email, password)
	if err != nil {
		logger.Error.Printf("Login error: %v", err)
		return response
	}

	// Check if login was successful by validating token presence
	if response.Data.Token != "" {
		a.authToken = response.Data.Token
		logger.Info.Printf("Login successful, connecting to MQTT...")

		// Connect to MQTT in goroutine to avoid blocking login response
		go a.connectMQTT()
	}

	return response
}

// connectMQTT establishes MQTT connection with message handler
func (a *App) connectMQTT() {
	// Check if already connected
	if a.mqtt != nil && a.mqtt.IsConnected() {
		logger.Info.Printf("[MQTT] Already connected, skipping reconnection")
		return
	}

	// Check settings before attempting connection
	settings, err := a.setting.GetSettings()
	if err != nil {
		logger.Error.Printf("[MQTT] Cannot get settings: %v", err)
		return
	}

	// Log settings for debugging (without sensitive data)
	logger.Info.Printf("[MQTT] Attempting connection with Broker: %s, Tenant: %s",
		settings.MqttBroker, settings.TenantCode)

	if settings.MqttBroker == "" {
		logger.Error.Printf("[MQTT] ✗ MQTT Broker URL is not configured in settings")
		return
	}
	if settings.TenantCode == "" {
		logger.Error.Printf("[MQTT] ✗ Tenant Code is not configured in settings")
		return
	}

	// Define message handler for incoming MQTT messages
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		logger.Info.Printf("Pesan diterima dari topic %s:\n%+v\n", msg.Topic(), msg.Payload())

		// Emit event to frontend for real-time message handling
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "mqtt-message", map[string]interface{}{
				"topic":   msg.Topic(),
				"payload": string(msg.Payload()),
			})
		}

		// Handle recording commands from MQTT
		a.handleMQTTRecordingCommand(msg.Payload())
	}

	// Attempt connection (errors logged but don't prevent app usage)
	if err := a.mqtt.Connect(messageHandler); err != nil {
		logger.Error.Printf("[MQTT] ✗ Failed to connect: %v", err)
		return
	}

	logger.Info.Printf("[MQTT] Connection initiated successfully")
}

// handleMQTTRecordingCommand processes MQTT messages for recording control
func (a *App) handleMQTTRecordingCommand(payload []byte) {
	var msg types.RecordMQTTPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		logger.Error.Printf("[MQTT] Failed to parse recording command: %v", err)
		return
	}

	logger.Info.Printf("[MQTT] Recording command received: action=%s, session_id=%s, client_id=%s",
		msg.Action, msg.SessionId, msg.ClientID)

	switch msg.Action {
	case "start":
		if msg.SessionId == "" {
			logger.Error.Printf("[MQTT] Cannot start recording: session_id is empty")
			return
		}

		// If already recording, stop existing recording first
		status := a.recorder.GetStatus()
		if status.State == recorder.StateRecording {
			logger.Info.Printf("[MQTT] Stopping existing recording before starting new one")
			if _, err := a.recorder.StopRecording(); err != nil {
				logger.Error.Printf("[MQTT] Failed to stop existing recording: %v", err)
			}
		}

		// Start new recording with session ID
		resp := a.startRecordingWithSession(msg.SessionId)
		if !resp.Success {
			logger.Error.Printf("[MQTT] Failed to start recording: %s", resp.Message)
		} else {
			logger.Info.Printf("[MQTT] Recording started with session_id=%s", msg.SessionId)
		}

		// Emit recording state to frontend
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "recording-state", map[string]interface{}{
				"state":      "recording",
				"session_id": msg.SessionId,
			})
		}

	case "stop":
		status := a.recorder.GetStatus()
		if status.State != recorder.StateRecording {
			logger.Error.Printf("[MQTT] Cannot stop recording: no recording in progress")
			return
		}

		resp := a.StopRecording()
		if !resp.Success {
			logger.Error.Printf("[MQTT] Failed to stop recording: %s", resp.Message)
		} else {
			logger.Info.Printf("[MQTT] Recording stopped, saved to: %s", resp.FilePath)
			go a.uploadRecording(resp.FilePath, msg.SessionId)
		}

		// Emit recording state to frontend
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "recording-state", map[string]interface{}{
				"state":     "idle",
				"file_path": resp.FilePath,
			})
		}

	default:
		logger.Info.Printf("[MQTT] Unhandled action: %s", msg.Action)
	}
}

// uploadRecording uploads the recording file using the signed-URL flow:
// 1. POST /api/record/upload/init → get signed PUT URL
// 2. PUT file directly to GCS via signed URL
// 3. POST /api/record/upload/done → confirm upload
func (a *App) uploadRecording(filePath, sessionId string) {
	logger.Info.Printf("[Upload] Starting signed-URL upload for file: %s, session: %s", filePath, sessionId)

	// Get settings
	settings, err := a.setting.GetSettings()
	if err != nil {
		logger.Error.Printf("[Upload] Failed to get settings: %v", err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, "Failed to get settings")
		return
	}

	if settings.BaseUrl == "" {
		logger.Error.Printf("[Upload] BaseUrl is not configured")
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, "BaseUrl is not configured")
		return
	}

	if a.authToken == "" {
		logger.Error.Printf("[Upload] Auth token is not available")
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, "Auth token is not available")
		return
	}

	// Read the file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error.Printf("[Upload] Failed to read file %s: %v", filePath, err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, err.Error())
		return
	}

	fileSize := len(fileBytes)
	filename := filepath.Base(filePath)
	contentType := "video/webm"
	if strings.HasSuffix(strings.ToLower(filename), ".mp4") {
		contentType = "video/mp4"
	}

	logger.Info.Printf("[Upload] File size: %.2f MB, filename: %s, content_type: %s",
		float64(fileSize)/(1024*1024), filename, contentType)

	// ── Step 1: POST /api/record/upload/init ──
	logger.Info.Printf("[Upload] Step 1: Initiating upload...")

	initBody := map[string]interface{}{
		"session_id":   sessionId,
		"filename":     filename,
		"content_type": contentType,
		"file_size":    fileSize,
	}

	initURL := fmt.Sprintf("%s/api/record/upload/init", settings.BaseUrl)
	initResp, err := helper.HTTPRequest(
		&helper.HTTPRequestPayload{
			Method: enum.POST,
			URL:    initURL,
			Body:   initBody,
		},
		&helper.HTTPRequestConfig{
			Ctx: context.Background(),
			Headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + a.authToken},
			},
		},
	)
	if err != nil {
		logger.Error.Printf("[Upload] Step 1 failed: %v", err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, err.Error())
		return
	}

	initDataJSON, _ := json.Marshal(initResp.Data)
	logger.Info.Printf("[Upload] Step 1 response (status %d): %s", initResp.StatusCode, string(initDataJSON))

	if initResp.StatusCode < 200 || initResp.StatusCode >= 300 {
		logger.Error.Printf("[Upload] Step 1 failed with status %d", initResp.StatusCode)
		a.emitUploadEvent(false, filePath, sessionId, initResp.StatusCode, initResp.Data, "Init upload failed")
		return
	}

	// Parse init response to get upload_url and upload_id
	initDataMap, ok := initResp.Data.(map[string]interface{})
	if !ok {
		logger.Error.Printf("[Upload] Step 1: unexpected response format")
		a.emitUploadEvent(false, filePath, sessionId, initResp.StatusCode, initResp.Data, "Unexpected init response format")
		return
	}

	// Extract data from nested "data" field
	dataField, ok := initDataMap["data"]
	if !ok {
		logger.Error.Printf("[Upload] Step 1: missing 'data' field in response")
		a.emitUploadEvent(false, filePath, sessionId, initResp.StatusCode, initResp.Data, "Missing data in init response")
		return
	}

	dataMap, ok := dataField.(map[string]interface{})
	if !ok {
		logger.Error.Printf("[Upload] Step 1: unexpected 'data' field format")
		a.emitUploadEvent(false, filePath, sessionId, initResp.StatusCode, initResp.Data, "Unexpected data format in init response")
		return
	}

	uploadURL, _ := dataMap["upload_url"].(string)
	uploadID, _ := dataMap["upload_id"].(string)

	if uploadURL == "" || uploadID == "" {
		logger.Error.Printf("[Upload] Step 1: missing upload_url or upload_id")
		a.emitUploadEvent(false, filePath, sessionId, initResp.StatusCode, initResp.Data, "Missing upload_url or upload_id")
		return
	}

	logger.Info.Printf("[Upload] Step 1 ✓ Got upload_id=%s, upload_url=%s...", uploadID, uploadURL[:min(80, len(uploadURL))])

	// ── Step 2: PUT file to GCS signed URL ──
	logger.Info.Printf("[Upload] Step 2: Uploading file to GCS...")

	putReq, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		logger.Error.Printf("[Upload] Step 2: failed to create PUT request: %v", err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, err.Error())
		return
	}
	putReq.Header.Set("Content-Type", contentType)

	httpClient := &http.Client{Timeout: 10 * time.Minute}
	putResp, err := httpClient.Do(putReq)
	if err != nil {
		logger.Error.Printf("[Upload] Step 2: PUT to GCS failed: %v", err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, err.Error())
		return
	}
	defer putResp.Body.Close()

	putBody, _ := io.ReadAll(putResp.Body)
	logger.Info.Printf("[Upload] Step 2 response: status=%d, body=%s", putResp.StatusCode, string(putBody))

	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		logger.Error.Printf("[Upload] Step 2: GCS upload failed with status %d", putResp.StatusCode)
		a.emitUploadEvent(false, filePath, sessionId, putResp.StatusCode, string(putBody), "GCS upload failed")
		return
	}

	logger.Info.Printf("[Upload] Step 2 ✓ File uploaded to GCS successfully")

	// ── Step 3: POST /api/record/upload/done ──
	logger.Info.Printf("[Upload] Step 3: Confirming upload...")

	doneBody := map[string]interface{}{
		"session_id": sessionId,
		"upload_id":  uploadID,
	}

	doneURL := fmt.Sprintf("%s/api/record/upload/done", settings.BaseUrl)
	doneResp, err := helper.HTTPRequest(
		&helper.HTTPRequestPayload{
			Method: enum.POST,
			URL:    doneURL,
			Body:   doneBody,
		},
		&helper.HTTPRequestConfig{
			Ctx: context.Background(),
			Headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + a.authToken},
			},
		},
	)
	if err != nil {
		logger.Error.Printf("[Upload] Step 3 failed: %v", err)
		a.emitUploadEvent(false, filePath, sessionId, 0, nil, err.Error())
		return
	}

	doneDataJSON, _ := json.Marshal(doneResp.Data)
	logger.Info.Printf("[Upload] Step 3 response (status %d): %s", doneResp.StatusCode, string(doneDataJSON))

	if doneResp.StatusCode >= 200 && doneResp.StatusCode < 300 {
		logger.Info.Printf("[Upload] ✓ Upload completed successfully (file: %s, session: %s)", filename, sessionId)

		// Check if we should delete the local file
		uploadSettings, err := a.setting.GetUploadSettings()
		if err == nil && uploadSettings.DeleteAfterUpload {
			if err := os.Remove(filePath); err != nil {
				logger.Error.Printf("[Upload] Failed to delete local file: %v", err)
			} else {
				logger.Info.Printf("[Upload] ✓ Local file deleted: %s", filePath)
			}
		}

		a.emitUploadEvent(true, filePath, sessionId, doneResp.StatusCode, doneResp.Data, "")
	} else {
		logger.Error.Printf("[Upload] ✗ Upload confirmation failed (status: %d, file: %s, session: %s)",
			doneResp.StatusCode, filename, sessionId)
		a.emitUploadEvent(false, filePath, sessionId, doneResp.StatusCode, doneResp.Data, "Upload confirmation failed")
	}
}

// emitUploadEvent sends upload status to the frontend
func (a *App) emitUploadEvent(success bool, filePath, sessionId string, statusCode int, response interface{}, errMsg string) {
	if a.ctx == nil {
		return
	}
	event := map[string]interface{}{
		"success":   success,
		"filePath":  filePath,
		"sessionId": sessionId,
	}
	if statusCode > 0 {
		event["statusCode"] = statusCode
	}
	if response != nil {
		event["response"] = response
	}
	if errMsg != "" {
		event["error"] = errMsg
	}
	runtime.EventsEmit(a.ctx, "recording-uploaded", event)
}

// ConnectMQTT is an exported method to connect MQTT (callable from frontend)
func (a *App) ConnectMQTT() map[string]interface{} {
	if a.mqtt == nil {
		return map[string]interface{}{
			"success": false,
			"message": "MQTT service not initialized",
		}
	}

	// Check if already connected
	if a.mqtt.IsConnected() {
		return map[string]interface{}{
			"success": true,
			"message": "Already connected to MQTT",
		}
	}

	// Connect in background
	go a.connectMQTT()

	return map[string]interface{}{
		"success": true,
		"message": "MQTT connection initiated",
	}
}

// Logout performs user logout
func (a *App) Logout(token string) interface{} {
	// Disconnect MQTT before logging out
	if a.mqtt != nil && a.mqtt.IsConnected() {
		logger.Info.Printf("Disconnecting MQTT on logout...")
		a.mqtt.Disconnect()
		logger.Info.Printf("MQTT disconnected")
	}

	// Clear stored auth token
	a.authToken = ""

	response, err := a.auth.Logout(token)
	if err != nil {
		logger.Error.Printf("Logout error: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": "Logout failed",
		}
	}
	return response
}

// Requirement represents a status requirement
func (a *App) Quit() {
	// Disconnect MQTT if connected
	if a.mqtt != nil && a.mqtt.IsConnected() {
		logger.Info.Printf("Disconnecting MQTT on quit...")
		a.mqtt.Disconnect()
	}

	if a.activityTracker != nil {
		a.activityTracker.Stop()
	}

	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

func (a *App) OnWindowClose() {
	// Instead of closing, minimize to tray
	// a.MinimizeToTray()
}

// HandleDeepLink processes incoming deep link URL and authenticates user
func (a *App) HandleDeepLink(deepLinkURL string) {
	fmt.Println("=== TEST DEEP LINK AUTH ===")
	fmt.Println("Deep Link URL:", deepLinkURL)
	fmt.Println("=========================")
	// Normalize deep link URL: handle "onxrecord://email=..." (missing '?')
	// by inserting '?' after the scheme so url.Parse can extract query params.
	const scheme = "onxrecord://"
	if strings.HasPrefix(deepLinkURL, scheme) {
		rest := deepLinkURL[len(scheme):]
		// If the part after scheme doesn't start with '?' and contains '=',
		// it's likely query params without the '?' separator.
		if rest != "" && !strings.HasPrefix(rest, "?") && strings.Contains(rest, "=") {
			deepLinkURL = scheme + "?" + rest
		}
	}

	// Parse the URL to extract data/token
	parsed, err := url.Parse(deepLinkURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		fmt.Println("=========================")
		return
	}

	// Extract query parameters
	query := parsed.Query()
	email := query.Get("email")
	token := query.Get("token")

	// Store token for API calls
	if token != "" {
		a.authToken = token
	}

	// Validate required parameters
	if email == "" || token == "" {
		logger.Error.Printf("Missing required parameters in deep link. Email: %s, Token: %s", email, token)
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "deep-link-auth-error", map[string]interface{}{
				"message": "Missing required parameters (email or token)",
			})
		}
		return
	}

	// Emit loading state to frontend
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "deep-link-auth-loading", map[string]interface{}{
			"message": "Authenticating...",
			"email":   email,
		})
	}

	// Ensure auth service is initialized
	if a.auth == nil {
		logger.Error.Printf("Auth service not initialized, initializing...")
		if a.setting == nil {
			logger.Error.Printf("Setting service not initialized")
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "deep-link-auth-error", map[string]interface{}{
					"message": "Setting service not initialized",
				})
			}
			return
		}
		ctx := a.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		a.auth = auth.NewService(ctx, func() (string, error) {
			settings, err := a.setting.GetSettings()
			if err != nil {
				return "", err
			}
			return settings.BaseUrl, nil
		})
	}

	// Check baseurl before making API call
	settings, err := a.setting.GetSettings()
	if err != nil {
		logger.Error.Printf("Failed to get settings: %v", err)
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "deep-link-auth-error", map[string]interface{}{
				"message": "Failed to get settings",
			})
		}
		return
	}

	if settings.BaseUrl == "" {
		logger.Error.Printf("BaseURL not configured in settings")
		fmt.Println("ERROR: BaseURL not configured. Please set it in Settings first.")
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "deep-link-auth-error", map[string]interface{}{
				"message": "Base URL not configured. Please set it in Settings first.",
			})
		}
		return
	}

	// Authenticate via deep link
	fmt.Println("Calling DeepLinkAuth API...")
	response, err := a.auth.DeepLinkAuth(email, token)
	if err != nil {
		logger.Error.Printf("Deep link auth failed: %v", err)
		fmt.Printf("ERROR: Deep link auth failed: %v\n", err)
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "deep-link-auth-error", map[string]interface{}{
				"message": fmt.Sprintf("Authentication failed: %v", err),
			})
		}
		return
	}

	fmt.Printf("Auth Response Message: %s\n", response.Message)
	fmt.Printf("User Data: ID=%d, Email=%s\n", response.Data.ID, response.Data.Email)

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "deep-link-auth-success", map[string]interface{}{
			"message": response.Message,
			"user":    response.Data,
		})
		a.ShowWindow()
	}
}

// SetInitialDeepLink stores the initial deep link for processing after startup
func (a *App) SetInitialDeepLink(deepLinkURL string) {
	a.initialDeepLink = deepLinkURL
}

// TestDeepLinkAuth is a helper method to test deep link auth from frontend
// Call this with: onxrecord://auth?email=test@example.com&token=test123
func (a *App) TestDeepLinkAuth(email string, token string) interface{} {
	fmt.Println("=== TEST DEEP LINK AUTH ===")
	fmt.Printf("Email: %s, Token: %s\n", email, token)

	if a.auth == nil {
		if a.setting == nil {
			return map[string]interface{}{
				"success": false,
				"message": "Setting service not initialized",
			}
		}
		ctx := a.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		a.auth = auth.NewService(ctx, func() (string, error) {
			settings, err := a.setting.GetSettings()
			if err != nil {
				return "", err
			}
			return settings.BaseUrl, nil
		})
	}

	response, err := a.auth.DeepLinkAuth(email, token)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": response.Message,
		"data":    response.Data,
	}
}
