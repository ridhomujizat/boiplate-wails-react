package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetPath returns the path to the ffmpeg binary.
// On Windows, it first checks for a bundled ffmpeg.exe next to the application executable,
// then falls back to exec.LookPath.
// On other platforms, it uses exec.LookPath only.
func GetPath() (string, error) {
	if runtime.GOOS == "windows" {
		// Check for bundled ffmpeg.exe next to the application executable
		exePath, err := os.Executable()
		if err == nil {
			bundledPath := filepath.Join(filepath.Dir(exePath), "ffmpeg.exe")
			if _, err := os.Stat(bundledPath); err == nil {
				return bundledPath, nil
			}
		}
	}

	// Fall back to PATH lookup
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}
	return path, nil
}

// CheckAvailability runs ffmpeg -version and returns the version string or an error.
func CheckAvailability() (string, error) {
	ffmpegPath, err := GetPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ffmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed to execute: %w", err)
	}

	// Extract first line which contains the version info
	lines := strings.SplitN(string(output), "\n", 2)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return strings.TrimSpace(string(output)), nil
}

// Command creates an *exec.Cmd for ffmpeg with the given arguments.
// This is a drop-in replacement for exec.Command("ffmpeg", args...).
func Command(args ...string) *exec.Cmd {
	ffmpegPath, err := GetPath()
	if err != nil {
		// Fall back to bare "ffmpeg" so the error surfaces when the command runs
		ffmpegPath = "ffmpeg"
	}
	return exec.Command(ffmpegPath, args...)
}
