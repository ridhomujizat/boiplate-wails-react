package app

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
	}

	// Attempt connection (errors logged but don't prevent app usage)
	if err := a.mqtt.Connect(messageHandler); err != nil {
		logger.Error.Printf("[MQTT] ✗ Failed to connect: %v", err)
		return
	}

	logger.Info.Printf("[MQTT] Connection initiated successfully")
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

	// Validate required parameters
	if email == "" || token == "" {
		logger.Error.Printf("Missing required parameters in deep link. Email: %s, Token: %s", email, token)
		return
	}

	// Ensure auth service is initialized
	if a.auth == nil {
		logger.Error.Printf("Auth service not initialized, initializing...")
		if a.setting == nil {
			logger.Error.Printf("Setting service not initialized")
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
		return
	}

	if settings.BaseUrl == "" {
		logger.Error.Printf("BaseURL not configured in settings")
		fmt.Println("ERROR: BaseURL not configured. Please set it in Settings first.")
		return
	}

	// Authenticate via deep link
	fmt.Println("Calling DeepLinkAuth API...")
	response, err := a.auth.DeepLinkAuth(email, token)
	if err != nil {
		logger.Error.Printf("Deep link auth failed: %v", err)
		fmt.Printf("ERROR: Deep link auth failed: %v\n", err)
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
