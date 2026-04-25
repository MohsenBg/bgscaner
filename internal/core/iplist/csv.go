package iplist

import (
	"context"
	"io"
	"strconv"
	"strings"

	"bgscan/internal/core/filemanager"
	"bgscan/internal/core/ip"
)

// DefaultCSVConfig defines the canonical format for .csv IP list files.
// Each row follows:  <ip-or-cidr>,<enable-flag>
//
// • Columns:
//
//   - col[0]: IP address or CIDR (e.g. "192.168.1.1" or "10.0.0.0/24")
//
//   - col[1]: Enable flag (1 = active, 0 = disabled, optional)
//
//   - Example:
//     192.168.1.1,1
//     10.0.0.0/24,0
var DefaultCSVConfig = filemanager.CSVConfig{
	Comma:            ',',
	HasHeader:        false,
	FieldsPerRecord:  -1, // allow variable number of columns
	LazyQuotes:       true,
	TrimLeadingSpace: true,
}

// ParseRecord converts a CSV row into an IPList entry.
// Returns (entry,false) if the row is syntactically invalid.
//
// It accepts both single IP and CIDR notation.
// If Enable flag is omitted, defaults to “1”.
func ParseRecord(rec []string) (IPList, bool) {
	if len(rec) == 0 {
		return IPList{}, false
	}

	raw := strings.TrimSpace(rec[0])
	ipStr, ok := ip.NormalizeIPOrCIDR(raw)
	if !ok {
		return IPList{}, false
	}

	enable := 1
	if len(rec) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(rec[1])); err == nil {
			enable = v
		}
	}

	return New(ipStr, enable), true
}

// ReadCSV streams IPList entries from a CSV file at the given path.
//
// It uses a memory‑efficient streaming reader and calls fn for each
// successfully parsed row.  Invalid lines are silently skipped.
func ReadCSV(path string, fn func(IPList) error) error {
	return filemanager.StreamCSV(path, DefaultCSVConfig, func(rec []string) error {
		entry, ok := ParseRecord(rec)
		if !ok {
			return nil // skip invalid rows
		}
		return fn(entry)
	})
}

// WriteCSV writes a stream of IPList entries to a CSV file.
//
// The provided callback receives a writer function that writes a single row.
// Example:
//
//	WriteCSV("out.csv", func(write func(IPList) error) error {
//	    for _, row := range items {
//	        if err := write(row); err != nil {
//	            return err
//	        }
//	    }
//	    return nil
//	})
func WriteCSV(path string, fn func(func(IPList) error) error) error {
	return filemanager.StreamWriteCSV(path, DefaultCSVConfig, func(write func([]string) error) error {
		return fn(func(item IPList) error {
			return write(item.EncodeCSV())
		})
	})
}

// StreamActiveIPs reads a CSV file and sends active (enabled) IPs to the out channel.
// CIDRs are expanded into individual IPs.
//
// The stream stops cleanly if the context is canceled.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	out := make(chan string)
//	go func() {
//	    defer close(out)
//	    StreamActiveIPs(ctx, "targets.csv",0, out)
//	}()
//
//	for ip := range out { fmt.Println(ip) }
func StreamActiveIPs(ctx context.Context, path string, limit int, out chan<- string) error {
	count := 0

	return ReadCSV(path, func(row IPList) error {
		if !row.Enable {
			return nil
		}

		if limit > 0 && count >= limit {
			return io.EOF // stop CSV iteration
		}

		if row.IsCIDR() {
			return ip.StreamCIDR(ctx, row.IP, limit-count, out)
		}

		select {
		case out <- row.IP:
			count++
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
