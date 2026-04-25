package iplist

import "strings"

// IPList represents a single IP address or CIDR range entry
// with an enable/disable flag.
//
// The IP field may contain either:
//
//   - Single IP:   "192.168.1.1"
//   - CIDR range:  "10.0.0.0/24"
//
// When Enable is false the entry is ignored by scanners.
type IPList struct {
	IP     string
	Enable bool
}

// New creates an IPList entry from a raw enable flag.
// Any value greater than zero is treated as enabled.
func New(ip string, enable int) IPList {
	return IPList{
		IP:     ip,
		Enable: enable > 0,
	}
}

// NewEnabled creates an enabled IPList entry.
func NewEnabled(ip string) IPList {
	return IPList{
		IP:     ip,
		Enable: true,
	}
}

// NewDisabled creates a disabled IPList entry.
func NewDisabled(ip string) IPList {
	return IPList{
		IP:     ip,
		Enable: false,
	}
}

// IsCIDR reports whether the entry represents a CIDR range.
func (e IPList) IsCIDR() bool {
	return strings.IndexByte(e.IP, '/') >= 0
}

// EncodeCSV returns the CSV representation of the entry.
//
// Format:
//
//	ip,enable
//
// Example:
//
//	192.168.1.1,1
//	10.0.0.0/24,0
func (e IPList) EncodeCSV() []string {
	enable := "0"
	if e.Enable {
		enable = "1"
	}

	return []string{e.IP, enable}
}
