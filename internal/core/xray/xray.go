package xray

import (
	"bgscan/internal/core/filemanager"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	// configPath is the directory where generated Xray configuration
	// files are written. Each scan target produces a dedicated config.
	configPath = "assets/xray/configs"

	// templatePath is the directory containing outbound configuration
	// templates used to construct Xray configs dynamically.
	templatePath = "assets/xray/templates"
)

// GenerateConfig builds a complete Xray configuration from a template
// and writes it to disk.
//
// The function performs the following steps:
//
//  1. Validates the provided IP address.
//  2. Loads the specified outbound template.
//  3. Replaces template placeholders (e.g. $ADDRESS).
//  4. Injects a scanner-generated inbound proxy.
//  5. Writes the final configuration file to disk.
//
// Each generated config contains:
//
//   - a local SOCKS inbound used by the scanner probes
//   - a single outbound derived from the selected template
//
// The returned value is the path to the generated configuration file,
// which can then be passed to an Xray process.
func GenerateConfig(templateName, ip string, port int) (string, error) {

	// Validate IP
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP: %s", ip)
	}

	// Get template file path
	tmplPath, err := GetTemplateByName(templateName)
	if err != nil {
		return "", err
	}

	// Read template
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template JSON
	var outbound any
	if err := json.Unmarshal(raw, &outbound); err != nil {
		return "", fmt.Errorf("failed to parse template JSON: %w", err)
	}

	// Replace placeholders
	replacements := map[string]string{
		"$ADDRESS": ip,
	}
	outbound = replacePlaceholders(outbound, replacements)

	// Build full config
	config := XrayConfig{
		Inbounds:  []Inbound{getInbound(port)},
		Outbounds: []any{outbound},
	}

	// Generate output path
	outputPath := getNewXrayConfigName(ip)

	// Write config file
	if err := filemanager.WriteJSONFile(outputPath, config); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return outputPath, nil
}

// GetTemplateByName resolves a template file from the template directory.
//
// If the provided name does not include a ".json" extension, it is
// automatically appended. The function verifies that the template file
// exists before returning its path.
func GetTemplateByName(name string) (string, error) {
	if ext := filepath.Ext(name); ext != ".json" {
		name = strings.TrimSuffix(name, ext) + ".json"
	}

	path := filepath.Join(templatePath, name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("cannot read template file %s: %w", path, err)
	}

	return path, nil
}

// getNewXrayConfigName returns the file path for a generated
// Xray configuration associated with the given IP address.
//
// Each target IP produces a dedicated configuration file stored
// inside the configPath directory.
func getNewXrayConfigName(ip string) string {
	filename := fmt.Sprintf("%s.json", ip)
	return filepath.Join(configPath, filename)
}
