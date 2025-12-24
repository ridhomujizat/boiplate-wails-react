//go:build darwin
// +build darwin

package tray

import (
	"context"
)

type TrayManager struct {
	ctx       context.Context
	isVisible bool
	onShow    func()
	onHide    func()
	onExit    func()
}

type TrayOptions struct {
	Title    string
	Tooltip  string
	IconData []byte
	OnShow   func()
	OnHide   func()
	OnExit   func()
}

// NewTrayManager creates a new tray manager (no-op for Darwin)
func NewTrayManager(ctx context.Context, options TrayOptions) *TrayManager {
	return &TrayManager{
		ctx:       ctx,
		isVisible: true,
		onShow:    options.OnShow,
		onHide:    options.OnHide,
		onExit:    options.OnExit,
	}
}

// Initialize sets up the system tray (no-op for Darwin)
func (t *TrayManager) Initialize() {
	// macOS tray is not implemented
}

// ShowWindow shows the main window (no-op for Darwin)
func (t *TrayManager) ShowWindow() {
	t.isVisible = true
	if t.onShow != nil {
		t.onShow()
	}
}

// HideWindow hides the main window (no-op for Darwin)
func (t *TrayManager) HideWindow() {
	t.isVisible = false
	if t.onHide != nil {
		t.onHide()
	}
}

// ToggleWindow toggles window visibility (no-op for Darwin)
func (t *TrayManager) ToggleWindow() {
	if t.isVisible {
		t.HideWindow()
	} else {
		t.ShowWindow()
	}
}

// IsVisible returns visibility status
func (t *TrayManager) IsVisible() bool {
	return t.isVisible
}
