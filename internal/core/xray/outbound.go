package xray

import (
	"bgscan/internal/core/filemanager"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//
// --- Placeholder Replacement Logic --------------------------------------------------------
//

// replacePlaceholders recursively walks a decoded JSON structure
// (map[string]any, []any, string, etc.) and replaces any string
// whose entire value matches a key from the replacements map.
//
// This function does not modify the original JSON byte stream;
// instead, it operates on the Go values created by json.Unmarshal.
//
// Example:
//
//	Input:
//	  {"address": "$ADDRESS"}
//
//	Replacements:
//	  {"$ADDRESS": "1.2.3.4"}
//
//	Output:
//	  {"address": "1.2.3.4"}
//
// It supports arbitrarily nested structures.
func replacePlaceholders(data any, replacements map[string]string) any {
	switch v := data.(type) {

	case map[string]any:
		for key, val := range v {
			v[key] = replacePlaceholders(val, replacements)
		}
		return v

	case []any:
		for i, val := range v {
			v[i] = replacePlaceholders(val, replacements)
		}
		return v

	case string:
		if newVal, ok := replacements[v]; ok {
			return newVal
		}
		return v

	default:
		return v
	}
}

//
// --- Template Application Logic -----------------------------------------------------------
//

// applyOutboundTemplate loads an outbound JSON template,
// replaces placeholders with runtime execution values (currently only $ADDRESS),
// and returns a formatted JSON document.
//
// The template file must be valid JSON and contain at least one
// "address": "$ADDRESS" placeholder. The provided IP is validated
// before substitution.
//
// This function does *not* write anything to disk.
func applyOutboundTemplate(templatePath, ip string) (any, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse outbound template JSON: %w", err)
	}

	modified := replacePlaceholders(parsed, map[string]string{
		"$ADDRESS": ip,
	})

	return modified, nil
}

//
// --- Template Saving Logic ----------------------------------------------------------------
//

// SaveOutbound validates and stores a new outbound template in the
// template directory. It performs the following steps:
//
//  1. Ensures the source file exists and is not a directory.
//  2. Normalizes the outbound name to always have ".json" extension.
//  3. Prevents overwriting an existing template.
//  4. Loads the JSON and verifies it is well-formed.
//  5. Ensures the JSON contains at least one "address": "$ADDRESS" placeholder.
//  6. Validates the outbound content using ValidateOutbound().
//  7. Saves the JSON exactly as-is to the template directory.
//
// Returns a populated XrayOutboundsFile metadata object on success.
func SaveOutbound(src, name string) (*XrayOutboundsFile, error) {
	// 1. Check source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source outbound file does not exist: %s", src)
		}
		return nil, fmt.Errorf("cannot access source file %s: %w", src, err)
	}
	if srcInfo.IsDir() {
		return nil, fmt.Errorf("source path is a directory, expected file: %s", src)
	}

	// 2. Normalize extension
	if ext := filepath.Ext(name); ext != ".json" {
		name = strings.TrimSuffix(name, ext) + ".json"
	}

	// 3. Resolve destination
	dst := filepath.Join(templatePath, name)

	// 4. Prevent overwrite
	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("outbound template %q already exists", name)
	}

	// 5. Read file
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("cannot read source file %s: %w", src, err)
	}

	// 6. Unmarshal into generic structure
	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON in outbound template: %w", err)
	}

	// 7. Ensure placeholder exists
	if !containsAddressPlaceholder(jsonData) {
		return nil, fmt.Errorf("outbound template missing required placeholder: \"address\": \"$ADDRESS\"")
	}

	// 8. Save as-is
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("failed to save outbound template: %w", err)
	}

	// 9. Validate outbound content
	if err := ValidateOutbound(name); err != nil {
		os.Remove(dst)
		return nil, fmt.Errorf("outbound validation failed: %w", err)
	}

	return &XrayOutboundsFile{
		Name:        filemanager.StripExt(name),
		Path:        dst,
		CreatedTime: time.Now(),
	}, nil
}

//
// --- Template Retrieval -------------------------------------------------------------------
//

// GetOutboundTemplateByName finds an outbound template by name,
// automatically normalizing the filename to have a ".json" extension.
//
// Returns an XrayOutboundsFile describing the template.
func GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	if ext := filepath.Ext(name); ext != ".json" {
		name = strings.TrimSuffix(name, ext) + ".json"
	}

	path := filepath.Join(templatePath, name)
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read outbound template %s: %w", path, err)
	}

	return &XrayOutboundsFile{
		Name:        filemanager.StripExt(name),
		CreatedTime: info.ModTime(),
		Path:        path,
	}, nil
}

//
// --- Template Renaming --------------------------------------------------------------------
//

// RenameOutboundTemplate renames an existing outbound template file
// to a new name. It guarantees the following:
//
//   - The source template must exist.
//   - The new name is normalized to ".json".
//   - The destination must NOT already exist (no overwrite).
//   - The rename operation is atomic on supported filesystems.
//
// Returns updated metadata for the renamed template.
func RenameOutboundTemplate(oldName, newName string) (*XrayOutboundsFile, error) {
	oldFile, err := GetOutboundTemplateByName(oldName)
	if err != nil {
		return nil, err
	}

	// normalize new name
	if ext := filepath.Ext(newName); ext != ".json" {
		newName = strings.TrimSuffix(newName, ext) + ".json"
	}
	dst := filepath.Join(templatePath, newName)

	// check overwrite
	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("cannot rename: destination template %q already exists", newName)
	}

	// rename file
	if err := os.Rename(oldFile.Path, dst); err != nil {
		return nil, fmt.Errorf("failed to rename template: %w", err)
	}

	return &XrayOutboundsFile{
		Name:        filemanager.StripExt(newName),
		Path:        dst,
		CreatedTime: oldFile.CreatedTime,
	}, nil
}

//
// --- Outbound Validation ------------------------------------------------------------------
//

// ValidateOutbound embeds a raw outbound JSON snippet into a temporary
// Xray configuration, writes it to disk, and validates it using
// ValidateConfig(). The temporary file is always removed.
//
// Any validation or config-generation issue is returned as an error.
func ValidateOutbound(outbound string) error {
	configPath, err := GenerateConfig(outbound, "127.0.0.1", 40443)
	if err != nil {
		return err
	}
	defer os.Remove(configPath)

	return ValidateConfig(configPath)
}

//
// --- Placeholder Search Utility ------------------------------------------------------------
//

// containsAddressPlaceholder recursively searches a decoded JSON
// structure for at least one occurrence of:
//
//	"address": "$ADDRESS"
//
// The search is deep and supports nested JSON objects and arrays.
// Returns true immediately upon finding a match.
func containsAddressPlaceholder(v any) bool {
	switch val := v.(type) {

	case map[string]any:
		for k, v2 := range val {
			if k == "address" {
				if s, ok := v2.(string); ok && s == "$ADDRESS" {
					return true
				}
			}
			if containsAddressPlaceholder(v2) {
				return true
			}
		}

	case []any:
		for _, item := range val {
			if containsAddressPlaceholder(item) {
				return true
			}
		}
	}

	return false
}

func GetOutboundsTemplates() ([]XrayOutboundsFile, error) {
	filter := func(name string, info os.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(name, ".json")
	}

	files, err := filemanager.ListFiles(templatePath, filter)
	if err != nil {
		return nil, err
	}

	templates := make([]XrayOutboundsFile, 0, len(files))

	for _, f := range files {
		templates = append(templates, XrayOutboundsFile{
			Name:        filemanager.StripExt(f.Name),
			Path:        f.Path,
			CreatedTime: f.Info.ModTime(),
		})
	}

	return templates, nil
}

