package result

import (
	"bgscan/internal/core/filemanager"
	"net"
	"strings"
	"time"
)

// csvConfig defines CSV parsing/dumping options.
var csvConfig = filemanager.CSVConfig{Comma: ','}

// ParseRecord converts a CSV record to an IPScanResult.
// It returns false if the record is invalid or cannot be parsed.
func ParseRecord(rec []string) (IPScanResult, bool) {
	if len(rec) < 1 {
		return IPScanResult{}, false
	}

	ip := net.ParseIP(strings.TrimSpace(rec[0]))
	if ip == nil {
		return IPScanResult{}, false
	}

	result := IPScanResult{
		IP:       ip.String(),
		Latency:  FallbackLatency,
		Download: 0,
		Upload:   0,
	}

	if len(rec) >= 2 {
		if d, err := time.ParseDuration(rec[1]); err == nil {
			result.Latency = d
		}
	}

	if len(rec) >= 3 {
		if d, err := time.ParseDuration(rec[2]); err == nil {
			result.Download = d
		}
	}

	if len(rec) >= 4 {
		if d, err := time.ParseDuration(rec[3]); err == nil {
			result.Upload = d
		}
	}

	return result, true
}

// ToRecord converts an IPScanResult into a CSV record.
func (r IPScanResult) ToRecord() []string {
	return []string{
		r.IP,
		r.Latency.String(),
		r.Download.String(),
		r.Upload.String(),
	}
}

// ReadCSV streams a CSV file and applies the provided function to each valid result.
func ReadCSV(path string, fn func(IPScanResult) error) error {
	return filemanager.StreamCSV(path, csvConfig, func(rec []string) error {
		result, ok := ParseRecord(rec)
		if !ok {
			return nil // skip invalid records
		}
		return fn(result)
	})
}

// StreamWriteResults streams results to a CSV file.
// It receives a function that itself accepts a function to write each result.
func StreamWriteResults(path string, fn func(func(IPScanResult) error) error) error {
	return filemanager.StreamWriteCSV(path, csvConfig, func(write func([]string) error) error {
		return fn(func(r IPScanResult) error {
			return write(r.ToRecord())
		})
	})
}
