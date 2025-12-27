//go:build darwin
// +build darwin

package permission

import (
	"os/exec"
)

// CheckScreenPermission checks if screen recording permission is granted
// Note: macOS doesn't provide a direct API to check this programmatically
// The app will know when it tries to capture and fails
func (p *PermissionManager) CheckScreenPermission() PermissionStatus {
	return PermissionStatus{
		Granted: false,
		Message: "Please grant screen recording permission in System Preferences",
	}
}

// RequestScreenPermission opens System Preferences to the screen recording section
func (p *PermissionManager) RequestScreenPermission() bool {
	// Open System Preferences > Privacy & Security > Screen Recording
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
	err := cmd.Run()
	return err == nil
}

// CheckAccessibilityPermission checks if accessibility permission is granted
func (p *PermissionManager) CheckAccessibilityPermission() PermissionStatus {
	return PermissionStatus{
		Granted: false,
		Message: "Please grant accessibility permission in System Preferences",
	}
}

// RequestAccessibilityPermission opens System Preferences to the accessibility section
func (p *PermissionManager) RequestAccessibilityPermission() bool {
	// Open System Preferences > Privacy & Security > Accessibility
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")
	err := cmd.Run()
	return err == nil
}
