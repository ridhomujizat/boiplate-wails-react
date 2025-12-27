//go:build windows
// +build windows

package permission

// CheckScreenPermission checks if screen recording permission is granted
// Windows typically doesn't require explicit screen recording permission
func (p *PermissionManager) CheckScreenPermission() PermissionStatus {
	return PermissionStatus{
		Granted: true,
		Message: "Screen recording is available",
	}
}

// RequestScreenPermission on Windows - no action needed
func (p *PermissionManager) RequestScreenPermission() bool {
	return true
}

// CheckAccessibilityPermission checks if accessibility permission is granted
// Windows typically doesn't require explicit accessibility permission
func (p *PermissionManager) CheckAccessibilityPermission() PermissionStatus {
	return PermissionStatus{
		Granted: true,
		Message: "Accessibility is available",
	}
}

// RequestAccessibilityPermission on Windows - no action needed
func (p *PermissionManager) RequestAccessibilityPermission() bool {
	return true
}
