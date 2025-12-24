//go:build windows
// +build windows

package tray

// ShowWindow shows the main window
func (t *TrayManager) ShowWindow() {
	t.isVisible = true
	t.menuShow.Hide()
	t.menuHide.Show()

	if t.onShow != nil {
		t.onShow()
	}
}

// HideWindow hides the main window
func (t *TrayManager) HideWindow() {
	t.isVisible = false
	t.menuHide.Hide()
	t.menuShow.Show()

	if t.onHide != nil {
		t.onHide()
	}
}

// ToggleWindow toggles window visibility
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
