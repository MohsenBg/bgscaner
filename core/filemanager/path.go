package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ═══════════════════════════════════════════════════════════
// Path Operations
// ═══════════════════════════════════════════════════════════

// GetCurrentPath returns the absolute path of the current working directory
func GetCurrentPath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current path: %w", err)
	}
	return path, nil
}

// CheckFileExists checks if a file exists at the given path
func CheckFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func StripExt(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}

func HasExt(name, ext string) bool {
	return strings.EqualFold(filepath.Ext(name), ext)
}
