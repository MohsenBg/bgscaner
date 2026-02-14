package ip

import "net"

// ParseIP parses a single IP (NO CIDR).
// Returns nil if input is not a valid single IP.
func ParseIP(input string) net.IP {
	// Reject CIDR explicitly
	if _, _, err := net.ParseCIDR(input); err == nil {
		return nil
	}
	return net.ParseIP(input)
}

// ParseCIDR parses a CIDR string.
func ParseCIDR(input string) (*net.IPNet, error) {
	_, netw, err := net.ParseCIDR(input)
	return netw, err
}

func NormalizeIPOrCIDR(raw string) (string, bool) {
	if _, _, err := net.ParseCIDR(raw); err == nil {
		return raw, true
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return "", false
	}
	return ip.String(), true
}

