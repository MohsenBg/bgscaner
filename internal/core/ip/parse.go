package ip

import (
	"fmt"
	"net"
	"strings"
)

// ParseIP parses a single IP address (no CIDR allowed).
// It explicitly rejects inputs that contain a slash (/), e.g. "192.168.0.0/24".
// Returns nil if the input is invalid.
func ParseIP(input string) net.IP {
	input = strings.TrimSpace(input)

	// Quick reject CIDR-form input
	if strings.Contains(input, "/") {
		return nil
	}

	ip := net.ParseIP(input)
	return ip
}

// ParseCIDR parses a CIDR string and returns its *net.IPNet.
// It returns a detailed error if parsing fails.
func ParseCIDR(input string) (*net.IPNet, error) {
	_, netw, err := net.ParseCIDR(strings.TrimSpace(input))
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", input, err)
	}
	return netw, nil
}

// NormalizeIPOrCIDR validates and normalizes an IP or CIDR input string.
// Returns the normalized string (canonical form) and a boolean indicating validity.
//
// Examples:
//
//	NormalizeIPOrCIDR("192.168.001.001") → "192.168.1.1", true
//	NormalizeIPOrCIDR("10.0.0.0/8") → "10.0.0.0/8", true
//	NormalizeIPOrCIDR("invalid") → "", false
func NormalizeIPOrCIDR(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)

	// Try CIDR first
	if _, _, err := net.ParseCIDR(raw); err == nil {
		return raw, true
	}

	// Try plain IP
	ip := net.ParseIP(raw)
	if ip == nil {
		return "", false
	}

	return ip.String(), true
}
