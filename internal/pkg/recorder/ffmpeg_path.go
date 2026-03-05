//go:build darwin
// +build darwin

package recorder

import (
	"os"
	"os/exec"
	"path/filepath"
)

// GetFFmpegPath returns the path to the ffmpeg binary.
// Priority: 1) Bundled in .app Resources, 2) System PATH
func GetFFmpegPath() string {
	execPath, err := os.Executable()
	if err == nil {
		// execPath = .../onx-screen-record.app/Contents/MacOS/onx-screen-record
		bundled := filepath.Join(filepath.Dir(execPath), "..", "Resources", "ffmpeg")
		if _, err := os.Stat(bundled); err == nil {
			return bundled
		}
	}

	// Fallback to system PATH (for development)
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path
	}

	return "ffmpeg" // last resort
}
