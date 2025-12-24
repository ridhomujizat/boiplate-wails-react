package helpers

import (
	"os"
	"path/filepath"
	"runtime"
)

type PathHelper struct {
	appName string
}

func NewPathHelper(appName string) *PathHelper {
	return &PathHelper{appName: appName}
}

// GetAppDataDir returns app data directory (for database, config, logs, etc)
func (p *PathHelper) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(os.Getenv("APPDATA"), p.appName)
	case "darwin":
		dir = filepath.Join(home, "Library", "Application Support", p.appName)
	default: // linux
		dir = filepath.Join(home, ".local", "share", p.appName)
	}

	return dir, os.MkdirAll(dir, 0755)
}

// GetStreamDataDir returns directory for stream data (video, image, document, etc)
func (p *PathHelper) GetStreamDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(home, "Documents", p.appName)
	case "darwin":
		dir = filepath.Join(home, "Documents", p.appName)
	default: // linux
		dir = filepath.Join(home, "Documents", p.appName)
	}

	return dir, os.MkdirAll(dir, 0755)
}

// GetTempDataDir returns directory for temporary data
func (p *PathHelper) GetTempDataDir() (string, error) {
	dir := filepath.Join(os.TempDir(), p.appName)
	return dir, os.MkdirAll(dir, 0755)
}

// GetPath returns full path for a file in specified directory type
func (p *PathHelper) GetPath(dirType, filename string) (string, error) {
	var dir string
	var err error

	switch dirType {
	case "appdata":
		dir, err = p.GetAppDataDir()
	case "stream":
		dir, err = p.GetStreamDataDir()
	case "temp":
		dir, err = p.GetTempDataDir()
	default:
		dir, err = p.GetAppDataDir()
	}

	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}
