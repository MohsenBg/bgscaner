package result

import "time"

// IPScanResult represents a single IP scan outcome.
// Latency, Download, and Upload are optional metrics that
// may be zero when not applicable for a given scan type.
type IPScanResult struct {
	IP       string        // target IP address
	Latency  time.Duration // round-trip or connection latency
	Download time.Duration // optional download time
	Upload   time.Duration // optional upload time
}

// Config controls the background result writer behavior
// and file flush frequencies.
type Config struct {
	// Interval at which newly found IPs are flushed to delta file.
	DeltaFlushInterval time.Duration

	// Interval at which delta file is merged into the main file.
	MergeFlushInterval time.Duration

	// Channel capacity for passing discovered results to writer goroutines.
	ChanSize int

	// File write buffer size in bytes.
	BufferSize int
}

// Default performance and safety values.
const (
	FallbackLatency       = 999 * time.Second // if latency cannot be measured
	DefaultChanSize       = 1024
	DefaultBufferSize     = 256 * 1024 // 256 KB
	MinDeltaFlushInterval = 2 * time.Second
	MinMergeFlushInterval = 5 * time.Second
)

// DefaultConfig returns a configuration populated with sane defaults.
func DefaultConfig() Config {
	return Config{
		DeltaFlushInterval: MinDeltaFlushInterval,
		MergeFlushInterval: MinMergeFlushInterval,
		ChanSize:           DefaultChanSize,
		BufferSize:         DefaultBufferSize,
	}
}

// Normalize ensures the configuration satisfies safe minimum values.
func (c *Config) Normalize() {
	if c.DeltaFlushInterval < MinDeltaFlushInterval {
		c.DeltaFlushInterval = MinDeltaFlushInterval
	}
	if c.MergeFlushInterval < MinMergeFlushInterval {
		c.MergeFlushInterval = MinMergeFlushInterval
	}
	if c.ChanSize <= 0 {
		c.ChanSize = DefaultChanSize
	}
	if c.BufferSize <= 0 {
		c.BufferSize = DefaultBufferSize
	}
}

// ResultType enumerates the supported scanner result categories.
type ResultType int

const (
	ResultAll ResultType = iota
	ResultICMP
	ResultTCP
	ResultHTTP
	ResultXRAY
)

// String returns the string label for the result type.
// Returns "unknown" for out-of-range values.
func (t ResultType) String() string {
	types := [...]string{"all", "icmp", "tcp", "http", "xray"}
	if t < 0 || int(t) >= len(types) {
		return "unknown"
	}
	return types[t]
}

// ResultFile describes metadata for a stored result file.
type ResultFile struct {
	Name        string     // file name (without extension)
	SizeBytes   int64      // file size in bytes
	CreatedTime time.Time  // last modification time
	Type        ResultType // classification of scanner type
	IPCount     int64      // total IP entries (-1 if not yet counted)
	Path        string     // absolute file path
}
