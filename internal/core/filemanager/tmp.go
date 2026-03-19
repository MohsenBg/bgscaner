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

// CreateTmpFile creates a temporary file using the provided pattern.
// It returns the open file, its absolute path, and an error if any.
func CreateTmpFile(pattern string) (*os.File, string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, "", fmt.Errorf("create temp file: %w", err)
	}

	absPath, err := filepath.Abs(f.Name())
	if err != nil {
		f.Close()
		_ = os.Remove(f.Name())
		return nil, "", fmt.Errorf("resolve temp file path: %w", err)
	}

	return f, absPath, nil
}

// CreateTmpJSONFile creates a temporary JSON file containing the provided data.
// The file is closed before returning, and its absolute path is returned.
func CreateTmpJSONFile(pattern string, data any) (string, error) {
	f, path, err := CreateTmpFile(pattern)
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("encode json to temp file: %w", err)
	}

	return path, nil
}
