package xray

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// replacePlaceholders recursively traverses a parsed JSON structure
// and replaces string values that match placeholder keys.
//
// The function supports arbitrary JSON structures including nested
// objects and arrays. When a string value matches a key in the
// replacements map, it is substituted with the corresponding value.
//
// Example:
//
//	Input JSON:
//	{
//	  "address": "$ADDRESS"
//	}
//
//	Replacements:
//	map[string]string{
//	    "$ADDRESS": "1.2.3.4",
//	}
//
//	Output:
//	{
//	  "address": "1.2.3.4"
//	}
//
// The traversal operates on generic decoded JSON types
// (map[string]any, []any, string, etc.) produced by json.Unmarshal.
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

// applyTemplate loads an Xray configuration template and injects
// runtime values into placeholder fields.
//
// The template must be a valid JSON document. Placeholders inside
// the template are replaced using a simple string substitution
// mechanism during JSON traversal.
//
// Currently supported placeholders:
//
//	$ADDRESS  → replaced with the provided IP address
//
// The resulting configuration is returned as formatted JSON ready
// to be written to a temporary config file and passed to Xray.
//
// The function validates the provided IP address to prevent
// accidental injection of invalid or malformed values.
func applyTemplate(templatePath string, ip string) ([]byte, error) {

	// Validate IP first (don’t inject garbage)
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	replacements := map[string]string{
		"$ADDRESS": ip,
	}

	modified := replacePlaceholders(parsed, replacements)

	return json.MarshalIndent(modified, "", "  ")
}
