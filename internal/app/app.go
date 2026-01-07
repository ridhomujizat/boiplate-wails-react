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
	"onx-screen-record/internal/service/setting"

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
	a.activityTracker.Start()

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
	}
	return response
}

// Requirement represents a status requirement
type Requirement struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

// GetRequirements returns the list of requirements
func (a *App) GetRequirements() []Requirement {
	return []Requirement{
		{ID: "1", Title: "User Authentication", Status: "completed", Progress: 100},
		{ID: "2", Title: "Dashboard Layout", Status: "completed", Progress: 100},
		{ID: "3", Title: "API Integration", Status: "pending", Progress: 45},
		{ID: "4", Title: "Data Validation", Status: "warning", Progress: 20},
		{ID: "5", Title: "Testing Coverage", Status: "pending", Progress: 60},
	}
}

func (a *App) Quit() {
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

// HandleDeepLink processes incoming deep link URL and prints data to terminal
func (a *App) HandleDeepLink(deepLinkURL string) {
	fmt.Println("=== DEEP LINK RECEIVED ===")
	fmt.Printf("Full URL: %s\n", deepLinkURL)

	// Parse the URL to extract data/token
	parsed, err := url.Parse(deepLinkURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		fmt.Println("=========================")
		return
	}

	fmt.Printf("Scheme: %s\n", parsed.Scheme)
	fmt.Printf("Host: %s\n", parsed.Host)
	fmt.Printf("Path: %s\n", parsed.Path)

	// Print query parameters
	for key, values := range parsed.Query() {
		fmt.Printf("Query[%s]: %v\n", key, values)
	}
	fmt.Println("=========================")
}

// SetInitialDeepLink stores the initial deep link for processing after startup
func (a *App) SetInitialDeepLink(deepLinkURL string) {
	a.initialDeepLink = deepLinkURL
}
