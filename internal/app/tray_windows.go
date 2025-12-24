//go:build windows
// +build windows

package app

import (
	"onx-screen-record/assets"
	"onx-screen-record/internal/pkg/tray"
)

// setupSystemTray initializes the system tray
func (a *App) setupSystemTray() {
	trayOptions := tray.TrayOptions{
		Title:    "Screen Recorder",
		Tooltip:  "Screen Recorder Application",
		IconData: assets.IconData,
		OnShow: func() {
			a.isVisible = true
		},
		OnHide: func() {
			a.isVisible = false
		},
		OnExit: func() {
			a.Quit()
		},
	}

	a.trayManager = tray.NewTrayManager(a.ctx, trayOptions)

	// Initialize tray in a separate goroutine to avoid blocking
	go func() {
		a.trayManager.Initialize()
	}()
}

// ShowWindow shows the main window
func (a *App) ShowWindow() {
	if a.trayManager != nil {
		a.trayManager.ShowWindow()
	}
}

// HideWindow hides the main window to system tray
func (a *App) HideWindow() {
	if a.trayManager != nil {
		a.trayManager.HideWindow()
	}
}

// ToggleWindow toggles window visibility
func (a *App) ToggleWindow() {
	if a.trayManager != nil {
		a.trayManager.ToggleWindow()
	}
}

// MinimizeToTray minimizes window to system tray
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
