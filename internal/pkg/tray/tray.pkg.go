//go:build windows
// +build windows

package tray

import (
	"context"
	"log"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TrayManager struct {
	ctx       context.Context
	iconData  []byte
	isVisible bool
	onShow    func()
	onHide    func()
	onExit    func()

	// Menu items
	menuShow *systray.MenuItem
	menuHide *systray.MenuItem
	menuExit *systray.MenuItem
}

type TrayOptions struct {
	Title    string
	Tooltip  string
	IconData []byte
	OnShow   func()
	OnHide   func()
	OnExit   func()
}

// NewTrayManager creates a new tray manager
func NewTrayManager(ctx context.Context, options TrayOptions) *TrayManager {
	return &TrayManager{
		ctx:       ctx,
		isVisible: true,
		onShow:    options.OnShow,
		onHide:    options.OnHide,
		onExit:    options.OnExit,
	}
}

// Initialize sets up the system tray
func (t *TrayManager) Initialize() {
	systray.Run(t.onReady(), t.handleTrayExit)
}

// onReady is called when systray is ready
func (t *TrayManager) onReady() func() {
	return func() {
		// Set tray icon and tooltip
		if len(t.iconData) > 0 {
			systray.SetIcon(t.iconData)
		} else {
			// Default icon data (simple dot)
			systray.SetIcon(getDefaultIcon())
		}

		systray.SetTitle("Screen Recorder")
		systray.SetTooltip("Screen Recorder - Click to show/hide")

		// Create menu items
		t.menuShow = systray.AddMenuItem("Show Window", "Show the main window")
		t.menuHide = systray.AddMenuItem("Hide Window", "Hide the main window")
		systray.AddSeparator()
		t.menuExit = systray.AddMenuItem("Exit", "Exit the application")

		// Initially hide the "Show" option since window is visible
		t.menuShow.Hide()

		// Handle menu clicks
		go t.handleMenuClicks()
	}
}

// getDefaultIcon returns a default icon as byte array
func getDefaultIcon() []byte {
	// This is a simple 16x16 black dot icon in ICO format
	// In a real application, you would load this from an embedded file
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10,
		0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x68, 0x04,
		0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

// handleMenuClicks processes menu item clicks
func (t *TrayManager) handleMenuClicks() {
	for {
		select {
		case <-t.menuShow.ClickedCh:
			t.ShowWindow()

		case <-t.menuHide.ClickedCh:
			t.HideWindow()

		case <-t.menuExit.ClickedCh:
			t.ExitApplication()
			return
		}
	}
}

// ShowWindow shows the main application window
func (t *TrayManager) ShowWindow() {
	if t.ctx != nil {
		runtime.Show(t.ctx)
		runtime.WindowUnminimise(t.ctx)
		t.isVisible = true
		t.menuShow.Hide()
		t.menuHide.Show()

		if t.onShow != nil {
			t.onShow()
		}
	}
}

// HideWindow hides the main application window
func (t *TrayManager) HideWindow() {
	if t.ctx != nil {
		runtime.Hide(t.ctx)
		t.isVisible = false
		t.menuHide.Hide()
		t.menuShow.Show()

		if t.onHide != nil {
			t.onHide()
		}
	}
}

// ExitApplication exits the application completely
func (t *TrayManager) ExitApplication() {
	if t.onExit != nil {
		t.onExit()
	}

	if t.ctx != nil {
		runtime.Quit(t.ctx)
	}

	systray.Quit()
}

// ToggleWindow toggles window visibility
func (t *TrayManager) ToggleWindow() {
	if t.isVisible {
		t.HideWindow()
	} else {
		t.ShowWindow()
	}
}

// IsVisible returns current window visibility status
func (t *TrayManager) IsVisible() bool {
	return t.isVisible
}

// UpdateIcon updates the tray icon
func (t *TrayManager) UpdateIcon(iconData []byte) {
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	}
}

// UpdateTooltip updates the tray tooltip
func (t *TrayManager) UpdateTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

// Quit stops the system tray
func (t *TrayManager) Quit() {
	systray.Quit()
}

// handleTrayExit is called when systray exits
func (t *TrayManager) handleTrayExit() {
	log.Println("System tray exited")
}

// getDe
