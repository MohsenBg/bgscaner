package config

import (
	"strings"
	"time"
)

// Change represents a single configuration adjustment performed during normalization.
type Change struct {
	Field string
	Old   any
	New   any
	Note  string
}

// ValidationReport collects normalization changes and validation errors.
type ValidationReport struct {
	Changes []Change
}

// AddChange records a configuration field modification.
func (r *ValidationReport) AddChange(field string, old, new any, note string) {
	r.Changes = append(r.Changes, Change{
		Field: field,
		Old:   old,
		New:   new,
		Note:  note,
	})
}

// HasErrors reports whether any validation errors were recorded.
func (r *ValidationReport) HasErrors() bool {
	return len(r.Changes) > 0
}

// normalizeInt ensures an integer value is within [min, max].
// If not, it resets the value to def and records the change.
func normalizeInt(field string, v *int, min, max, def int, r *ValidationReport) {
	old := *v

	if *v < min || *v > max {
		*v = def
		r.AddChange(field, old, *v, "out of range → default")
	}
}

// normalizeDuration ensures a DurationMS value is within [min, max].
// A small epsilon is used to tolerate minor rounding differences
// (e.g., from YAML/TOML parsing).
func normalizeDuration(field string, v *DurationMS, min, max, def DurationMS, r *ValidationReport) {
	d := v.Duration()

	const epsilon = time.Millisecond

	if d+epsilon < min.Duration() || d-epsilon > max.Duration() {
		old := *v
		*v = def
		r.AddChange(field, old, *v, "out of range → default")
	}
}

// normalizeString ensures a string is not empty or whitespace-only.
// If empty, it resets the value to def.
func normalizeString(field string, v *string, def string, r *ValidationReport) {
	if strings.TrimSpace(*v) == "" {
		old := *v
		*v = def
		r.AddChange(field, old, *v, "empty → default")
	}
}

// normalizeUint16 ensures a uint16 value is within [min, max].
// If not, it resets the value to def and records the change.
func normalizeUint16(field string, v *uint16, min, max, def uint16, r *ValidationReport) {
	old := *v
	if *v < min || *v > max {
		*v = def
		r.AddChange(field, old, *v, "out of range → default")
	}
}

