package config

import "time"

// WriterConfig defines configuration for the result writer subsystem.
type WriterConfig struct {
	// MergeFlushInterval controls how often buffered results
	// are flushed and merged.
	MergeFlushInterval DurationMS `toml:"merge_flush_interval"`

	// ChanSize defines the size of the internal writer channel.
	ChanSize int `toml:"chan_size"`

	// BatchSize defines how many records are processed per batch.
	BatchSize int `toml:"batch_size"`
}

// Normalize validates and normalizes the WriterConfig fields.
// Out-of-range values are replaced with defaults and recorded
// in the provided ValidationReport.
func (w *WriterConfig) Normalize(rep *ValidationReport) {
	def := DefaultWriterConfig()

	// MergeFlushInterval must be within [100ms, 5m].
	normalizeDuration(
		"Writer.MergeFlushInterval",
		&w.MergeFlushInterval,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(5*time.Minute),
		def.MergeFlushInterval,
		rep,
	)

	// ChanSize must be ≥ 1.
	normalizeInt(
		"Writer.ChanSize",
		&w.ChanSize,
		1,
		1_000_000,
		def.ChanSize,
		rep,
	)

	// BatchSize must be ≥ 1.
	normalizeInt(
		"Writer.BatchSize",
		&w.BatchSize,
		1,
		1_000_000,
		def.BatchSize,
		rep,
	)
}
