package filemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ═══════════════════════════════════════════════════════════
// JSON File Operations
// ═══════════════════════════════════════════════════════════

// WriteJSONFile writes data to a JSON file at the specified path
// Creates directory if it doesn't exist, overwrites if file exists
func WriteJSONFile(path string, data any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// WriteJSONFileIfNotExist writes JSON only if file doesn't exist
func WriteJSONFileIfNotExist(path string, data any) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteJSONFile(path, data)
}

// GetJSONFile reads a JSON file and unmarshals it into the provided destination
// Returns error if file doesn't exist or JSON is invalid
func GetJSONFile(path string, dest any) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

// GetJSONFileOrDefault attempts to load JSON from file, falls back to default on error
// This is useful for config loading where you want graceful degradation
func GetJSONFileOrDefault(path string, dest any, defaultValue any) error {
	err := GetJSONFile(path, dest)
	if err != nil {
		// Copy default value to dest
		defaultJSON, marshalErr := json.Marshal(defaultValue)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal default value: %w", marshalErr)
		}

		if unmarshalErr := json.Unmarshal(defaultJSON, dest); unmarshalErr != nil {
			return fmt.Errorf("failed to unmarshal default value: %w", unmarshalErr)
		}

		// Optionally write the default to file for next time
		_ = WriteJSONFile(path, defaultValue)

		return nil
	}

	return nil
}

// UpdateJSONFile updates a JSON file atomically
func UpdateJSONFile(path string, dest any, updateFn func(any) error) error {
	if err := GetJSONFile(path, dest); err != nil {
		return fmt.Errorf("failed to read existing file: %w", err)
	}

	if err := updateFn(dest); err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	if err := WriteJSONFile(path, dest); err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	return nil
}
