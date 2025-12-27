//go:build darwin
// +build darwin

package app

import (
	"onx-screen-record/internal/pkg/tray"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// setupSystemTray initializes the system tray (no-op for macOS)
func (a *App) setupSystemTray() {
	trayOptions := tray.TrayOptions{
		Title:   "Screen Recorder",
		Tooltip: "Screen Recorder Application",
	}

	a.trayManager = tray.NewTrayManager(a.ctx, trayOptions)
	// Tray is not initialized on macOS
}

// ShowWindow shows the main window (no-op for macOS)
func (a *App) ShowWindow() {
	if a.trayManager != nil {
		a.trayManager.ShowWindow()
	}
}

// HideWindow hides the main window to system tray (no-op for macOS)
func (a *App) HideWindow() {
	runtime.Hide(a.ctx)
}

// ToggleWindow toggles window visibility (no-op for macOS)
func (a *App) ToggleWindow() {
	if a.trayManager != nil {
		a.trayManager.ToggleWindow()
	}
}

// MinimizeToTray minimizes window to system tray (no-op for macOS)
func (a *App) MinimizeToTray() {
	a.HideWindow()
}

// IsWindowVisible returns current window visibility
func (a *App) IsWindowVisible() bool {
	if a.trayManager != nil {
		return a.trayManager.IsVisible()
	}
	return a.isVisible
}
