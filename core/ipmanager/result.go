package ipmanager

import "time"

// Defaults & tuning constants
const (
	DefaultFallbackLatency = 999 * time.Second
	DefaultChanSize        = 1024
	DefaultBufferSize      = 256 * 1024 // 256 KB
)

// ResultIPScan represents a single scan result for an IP.
type ResultIPScan struct {
	IP      string
	Latency time.Duration
}

// NewResultIPScan creates a ResultIPScan with explicit values.
func NewResultIPScan(ip string, latency time.Duration) ResultIPScan {
	return ResultIPScan{
		IP:      ip,
		Latency: latency,
	}
}

// EncodeCSV returns the CSV representation of the result.
func (r ResultIPScan) EncodeCSV() []string {
	return []string{
		r.IP,
		r.Latency.String(),
	}
}

// Less reports whether r should be ordered before other.
// Ordering: lower latency first, IP as tiebreaker.
func (r ResultIPScan) Less(other ResultIPScan) bool {
	if r.Latency == other.Latency {
		return r.IP < other.IP
	}
	return r.Latency < other.Latency
}

// ResultType represents the scan protocol/type.
type ResultType int

const (
	ResultAll ResultType = iota
	ResultICMP
	ResultTCP
	ResultHTTP
	ResultXRAY
)

// ResultFile describes a stored result file and its metadata.
type ResultFile struct {
	Name        string
	CreatedTime time.Time
	SizeBytes   int64
	IPCount     int64
	Type        ResultType
}
