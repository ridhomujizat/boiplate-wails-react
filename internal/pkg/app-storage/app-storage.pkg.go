package appstorage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type AppStorage struct {
	AppDir string
}

func NewAppStorage(appName string) (*AppStorage, error) {
	var baseDir string
	var err error

	// Menentukan lokasi folder data berdasarkan standar OS
	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA") // Roaming
	case "darwin":
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, "Library", "Application Support")
	case "linux":
		// Mengikuti standar XDG: ~/.local/share
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			baseDir = xdgData
		} else {
			home, _ := os.UserHomeDir()
			baseDir = filepath.Join(home, ".local", "share")
		}
	default:
		baseDir, err = os.UserConfigDir()
	}

	if err != nil || baseDir == "" {
		return nil, fmt.Errorf("tidak bisa menemukan direktori data")
	}

	appPath := filepath.Join(baseDir, appName)

	// Buat folder aplikasi jika belum ada
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return nil, err
	}

	return &AppStorage{AppDir: appPath}, nil
}

// GetPath memberikan path lengkap untuk nama file/folder yang diberikan
func (s *AppStorage) GetPath(subPath string) string {
	return filepath.Join(s.AppDir, subPath)
}

// EnsureSubDir berguna jika Anda ingin membuat sub-folder seperti 'logs' atau 'uploads'
func (s *AppStorage) EnsureSubDir(dirName string) (string, error) {
	path := filepath.Join(s.AppDir, dirName)
	err := os.MkdirAll(path, 0755)
	return path, err
}
