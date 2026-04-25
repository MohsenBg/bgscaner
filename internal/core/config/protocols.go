package config

import (
	"strings"
	"time"
)

// ICMPConfig defines configuration for ICMP probing.
type ICMPConfig struct {
	Workers      int        `toml:"workers"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	ShuffleIPs   bool       `toml:"shuffle_ips"`
	PrefixOutput string     `toml:"prefix_output"`
}

// Normalize validates ICMP configuration values and adjusts them to allowed ranges.
func (i *ICMPConfig) Normalize(rep *ValidationReport) {
	def := DefaultICMPConfig()

	normalizeInt("ICMP.Workers", &i.Workers, 1, 10000, def.Workers, rep)

	normalizeDuration(
		"ICMP.Timeout",
		&i.Timeout,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(30*time.Second),
		def.Timeout,
		rep,
	)

	normalizeUint16("ICMP.Tries", &i.Tries, 1, 10, def.Tries, rep)

	normalizeString("ICMP.PrefixOutput", &i.PrefixOutput, def.PrefixOutput, rep)
}

// TCPConfig defines configuration for TCP probing.
type TCPConfig struct {
	Workers      int        `toml:"workers"`
	Port         int        `toml:"port"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	ShuffleIPs   bool       `toml:"shuffle_ips"`
	PrefixOutput string     `toml:"prefix_output"`
}

// Normalize validates TCP configuration values and adjusts them to allowed ranges.
func (t *TCPConfig) Normalize(rep *ValidationReport) {
	def := DefaultTCPConfig()

	normalizeInt("TCP.Workers", &t.Workers, 1, 10000, def.Workers, rep)
	normalizeInt("TCP.Port", &t.Port, 1, 65535, def.Port, rep)

	normalizeUint16("TCP.Tries", &t.Tries, 1, 10, def.Tries, rep)

	normalizeDuration(
		"TCP.Timeout",
		&t.Timeout,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(30*time.Second),
		def.Timeout,
		rep,
	)

	normalizeString("TCP.PrefixOutput", &t.PrefixOutput, def.PrefixOutput, rep)
}

// HTTPConfig defines configuration for HTTP probing and TLS validation.
type HTTPConfig struct {
	Workers       int        `toml:"workers"`
	Host          string     `toml:"host"`
	ServerName    string     `toml:"server_name"`
	Port          int        `toml:"port"`
	Protocol      string     `toml:"protocol"`
	TLSValidation bool       `toml:"tls_validation"`
	MinTLSVersion string     `toml:"min_tls_version"`
	MaxTLSVersion string     `toml:"max_tls_version"`
	Timeout       DurationMS `toml:"timeout"`
	ShuffleIPs    bool       `toml:"shuffle_ips"`
	PrefixOutput  string     `toml:"prefix_output"`
}

// Normalize validates HTTP configuration values and adjusts invalid settings.
func (h *HTTPConfig) Normalize(rep *ValidationReport) {
	def := DefaultHTTPConfig()

	normalizeInt("HTTP.Workers", &h.Workers, 1, 5000, def.Workers, rep)

	normalizeString("HTTP.Host", &h.Host, def.Host, rep)

	normalizeInt("HTTP.Port", &h.Port, 1, 65535, def.Port, rep)

	// Protocol validation
	proto := strings.ToLower(strings.TrimSpace(h.Protocol))
	if proto != "http" && proto != "https" {
		old := h.Protocol
		h.Protocol = def.Protocol
		rep.AddChange("HTTP.Protocol", old, h.Protocol, "invalid → default")
	} else {
		h.Protocol = proto
	}

	normalizeDuration(
		"HTTP.Timeout",
		&h.Timeout,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(60*time.Second),
		def.Timeout,
		rep,
	)

	validTLS := map[string]bool{
		"tls1.0": true,
		"tls1.1": true,
		"tls1.2": true,
		"tls1.3": true,
	}

	min := strings.ToLower(strings.TrimSpace(h.MinTLSVersion))
	if !validTLS[min] {
		old := h.MinTLSVersion
		h.MinTLSVersion = def.MinTLSVersion
		rep.AddChange("HTTP.MinTLSVersion", old, h.MinTLSVersion, "invalid → default")
	} else {
		h.MinTLSVersion = min
	}

	max := strings.ToLower(strings.TrimSpace(h.MaxTLSVersion))
	if !validTLS[max] {
		old := h.MaxTLSVersion
		h.MaxTLSVersion = def.MaxTLSVersion
		rep.AddChange("HTTP.MaxTLSVersion", old, h.MaxTLSVersion, "invalid → default")
	} else {
		h.MaxTLSVersion = max
	}

	normalizeString("HTTP.PrefixOutput", &h.PrefixOutput, def.PrefixOutput, rep)
}

// XrayConfig defines configuration for Xray connectivity testing.
type XrayConfig struct {
	Workers              int              `toml:"workers"`
	ConnectivityTestType ConnectivityTest `toml:"connectivity_test_type"`
	DownloadSpeed        int              `toml:"download_speed"`
	UploadSpeed          int              `toml:"upload_speed"`
	Timeout              DurationMS       `toml:"timeout"`
	ShuffleIPs           bool             `toml:"shuffle_ips"`
	PrefixOutput         string           `toml:"prefix_output"`
	PreScanType          string           `toml:"pre_scan_type"`
}

// Normalize validates Xray configuration values and adjusts invalid settings.
func (x *XrayConfig) Normalize(rep *ValidationReport) {
	def := DefaultXrayConfig()

	normalizeInt("Xray.Workers", &x.Workers, 1, 1000, def.Workers, rep)

	if !x.ConnectivityTestType.IsValid() {
		old := x.ConnectivityTestType
		x.ConnectivityTestType = def.ConnectivityTestType
		rep.AddChange("Xray.ConnectivityTestType", old, x.ConnectivityTestType, "invalid → default")
	}

	normalizeInt("Xray.DownloadSpeed", &x.DownloadSpeed, 0, 10000, def.DownloadSpeed, rep)
	normalizeInt("Xray.UploadSpeed", &x.UploadSpeed, 0, 10000, def.UploadSpeed, rep)

	normalizeDuration(
		"Xray.Timeout",
		&x.Timeout,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(60*time.Second),
		def.Timeout,
		rep,
	)

	preScanType := strings.ToLower(strings.TrimSpace(x.PreScanType))
	switch preScanType {
	case "tcp", "icmp", "none", "http":
		x.PreScanType = preScanType
	default:
		old := x.PreScanType
		x.PreScanType = def.PreScanType
		rep.AddChange("Xray.PreScanType", old, x.PreScanType, "invalid → default")
	}

	normalizeString("Xray.PrefixOutput", &x.PrefixOutput, def.PrefixOutput, rep)
}
