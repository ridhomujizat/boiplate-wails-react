//go:build darwin
// +build darwin

package permission

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=10.15
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -mmacosx-version-min=10.15

#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>

int checkScreenCapturePermission() {
    if (@available(macOS 10.15, *)) {
        return CGPreflightScreenCaptureAccess() ? 1 : 0;
    }
    return 1;
}

int requestScreenCapturePermission() {
    if (@available(macOS 10.15, *)) {
        return CGRequestScreenCaptureAccess() ? 1 : 0;
    }
    return 1;
}

int isSystemAudioSupported() {
    if (@available(macOS 13.0, *)) {
        return 1;
    }
    return 0;
}
*/
import "C"

import (
	"os/exec"
)

func (p *PermissionManager) CheckScreenPermission() PermissionStatus {
	if C.checkScreenCapturePermission() == 1 {
		return PermissionStatus{
			Granted: true,
			Message: "Screen recording permission granted",
		}
	}
	return PermissionStatus{
		Granted: false,
		Message: "Please grant screen recording permission in System Preferences",
	}
}

func (p *PermissionManager) RequestScreenPermission() bool {
	if C.requestScreenCapturePermission() == 1 {
		return true
	}
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
	return cmd.Run() == nil
}

func (p *PermissionManager) CheckAccessibilityPermission() PermissionStatus {
	return PermissionStatus{
		Granted: false,
		Message: "Please grant accessibility permission in System Preferences",
	}
}

func (p *PermissionManager) RequestAccessibilityPermission() bool {
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")
	return cmd.Run() == nil
}

func (p *PermissionManager) IsSystemAudioSupported() bool {
	return C.isSystemAudioSupported() == 1
}

func (p *PermissionManager) RequestMicrophonePermission() bool {
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone")
	return cmd.Run() == nil
}
