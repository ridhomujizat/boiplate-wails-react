package helper

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetDeviceID returns the unique machine ID for the current platform
func GetDeviceID() (string, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", "(Get-ItemProperty -Path 'HKLM:\\SOFTWARE\\Microsoft\\Cryptography').MachineGuid")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil

	case "darwin":
		cmd := exec.Command("bash", "-c", "ioreg -rd1 -c IOPlatformExpertDevice | awk '/IOPlatformUUID/ { print $3; }'")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		result := strings.ReplaceAll(strings.TrimSpace(string(out)), "\"", "")
		return result, nil

	case "linux":
		cmd := exec.Command("cat", "/etc/machine-id")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil

	default:
		return "", fmt.Errorf("unsupported OS")
	}
}
