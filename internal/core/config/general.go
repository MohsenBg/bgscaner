package config

import (
	"strings"
	"time"
)

// GeneralConfig defines global scanner behavior and execution settings.
type GeneralConfig struct {
	StatusInterval        DurationMS `toml:"status_interval"`
	StopAfterFound        int        `toml:"stop_after_found"`
	MaxIPsToTest          int        `toml:"max_ips_to_test"`
	ChainMode             string     `toml:"chain_mode"`
	ChannelBufferMultiple int        `toml:"channel_buffer_multiple"`
	Verbose               bool       `toml:"verbose"`
}

// Normalize validates general configuration values and adjusts them to
// allowed ranges when necessary. All corrections are recorded in the
// provided ValidationReport.
func (g *GeneralConfig) Normalize(rep *ValidationReport) {
	def := DefaultGeneralConfig()

	// StatusInterval must be between 100ms and 1 minute.
	normalizeDuration(
		"General.StatusInterval",
		&g.StatusInterval,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(time.Minute),
		def.StatusInterval,
		rep,
	)

	// StopAfterFound must be non-negative.
	if g.StopAfterFound < 0 {
		old := g.StopAfterFound
		g.StopAfterFound = def.StopAfterFound
		rep.AddChange("General.StopAfterFound", old, g.StopAfterFound, "negative → default")
	}

	// MaxIPsToTest must be non-negative.
	if g.MaxIPsToTest < 0 {
		old := g.MaxIPsToTest
		g.MaxIPsToTest = def.MaxIPsToTest
		rep.AddChange("General.MaxIPsToTest", old, g.MaxIPsToTest, "negative → default")
	}

	// ChannelBufferMultiple must be within a safe range.
	normalizeInt(
		"General.ChannelBufferMultiple",
		&g.ChannelBufferMultiple,
		1,
		1000,
		def.ChannelBufferMultiple,
		rep,
	)

	// Validate ChainMode.
	mode := strings.ToLower(strings.TrimSpace(g.ChainMode))
	switch mode {
	case "simple", "advanced", "parallel":
		g.ChainMode = mode
	default:
		old := g.ChainMode
		g.ChainMode = def.ChainMode
		rep.AddChange(
			"General.ChainMode",
			old,
			g.ChainMode,
			"invalid → default (allowed: simple, advanced, parallel)",
		)
	}
}
