package config

import "time"

// ConnectivityTest represents the type of connectivity test
// performed by the Xray subsystem.
type ConnectivityTest uint8

const (
	// ConnectivityOnly performs only a connectivity check.
	ConnectivityOnly ConnectivityTest = iota

	// DownloadSpeedOnly measures download speed only.
	DownloadSpeedOnly

	// UploadSpeedOnly measures upload speed only.
	UploadSpeedOnly

	// Both performs connectivity and speed tests.
	Both
)

// String returns a human-readable representation of the test type.
func (c ConnectivityTest) String() string {
	names := [...]string{
		"Connectivity Only",
		"Download Speed Only",
		"Upload Speed Only",
		"Both",
	}

	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// IsValid reports whether the ConnectivityTest value is within the defined range.
func (c ConnectivityTest) IsValid() bool {
	return c <= Both
}

// ScannerConfig aggregates configuration for all scanner subsystems.
type ScannerConfig struct {
	General *GeneralConfig
	Writer  *WriterConfig
	ICMP    *ICMPConfig
	TCP     *TCPConfig
	HTTP    *HTTPConfig
	Xray    *XrayConfig
	DNS     *DNSConfig
}

// DurationMS represents a duration stored as milliseconds.
// It is mainly used for configuration values where durations
// are expressed as integer milliseconds (e.g., in TOML files).
type DurationMS int64

// NewDurationMS converts a time.Duration to DurationMS.
func NewDurationMS(d time.Duration) DurationMS {
	return DurationMS(d.Milliseconds())
}

// Duration converts DurationMS to a standard time.Duration.
func (d DurationMS) Duration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// SetDuration updates the value using a standard time.Duration.
func (d *DurationMS) SetDuration(v time.Duration) {
	*d = DurationMS(v.Milliseconds())
}

// String returns a human-readable representation of the duration.
func (d DurationMS) String() string {
	return d.Duration().String()
}
