package filemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ═══════════════════════════════════════════════════════════
// Temporary File Operations
// ═══════════════════════════════════════════════════════════

// CreateTmpFile creates a temporary file with the given pattern
func CreateTmpFile(pattern string) (*os.File, string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return tmpFile, absPath, nil
}

// CreateTmpJSONFile creates a temporary JSON file with data
func CreateTmpJSONFile(pattern string, data any) (string, error) {
	tmpFile, absPath, err := CreateTmpFile(pattern)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		_ = os.Remove(absPath)
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	if _, err := tmpFile.Write(jsonData); err != nil {
		_ = os.Remove(absPath)
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}

	return absPath, nil
}
