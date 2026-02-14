package ip

import "net"

// Strict single-IP validation (NO CIDR)
func ValidateSingleIP(input string) bool {
	// CIDR is NOT allowed here
	if _, _, err := net.ParseCIDR(input); err == nil {
		return false
	}

	return net.ParseIP(input) != nil
}

// Accepts IP or CIDR
func ValidateIPOrCIDR(input string) bool {
	if net.ParseIP(input) != nil {
		return true
	}
	if _, _, err := net.ParseCIDR(input); err == nil {
		return true
	}
	return false
}
