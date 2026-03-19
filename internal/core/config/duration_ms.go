package config

import "time"

// DurationMS represents a duration stored as milliseconds.
// It is primarily used for TOML configuration values where
// durations are expressed as integer milliseconds.
type DurationMS int64

// NewDurationMS converts a standard time.Duration to DurationMS.
func NewDurationMS(d time.Duration) DurationMS {
	return DurationMS(d.Milliseconds())
}

// Duration converts DurationMS back to a time.Duration.
func (d DurationMS) Duration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// SetDuration updates the value using a standard time.Duration.
func (d *DurationMS) SetDuration(v time.Duration) {
	*d = DurationMS(v.Milliseconds())
}

// String returns a human readable representation of the duration.
func (d DurationMS) String() string {
	return d.Duration().String()
}
