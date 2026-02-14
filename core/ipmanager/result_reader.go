package ipmanager

import (
	"net"
	"os"
	"strings"
	"time"

	"bgscan/core/filemanager"
)

// Result directories by scan type.
const (
	ICMPResultDir = "result/icmp/"
	TCPResultDir  = "result/tcp/"
	HTTPResultDir = "result/http/"
	XRAYResultDir = "result/xray/"
)

// shared CSV configuration for result files
var resultCSVConfig = filemanager.CSVConfig{
	Comma:           ',',
	HasHeader:       false,
	FieldsPerRecord: -1,
	LazyQuotes:      true,
}

// LoadResultIP streams valid scan results from a CSV file into out.
// Invalid rows are skipped silently.
func LoadResultIP(path string, out chan<- ResultIPScan) error {
	return filemanager.StreamCSV(path, resultCSVConfig, func(rec []string) error {
		r, ok := parseResultIP(rec)
		if !ok {
			return nil
		}
		out <- r
		return nil
	})
}

// CountResultIPs counts the number of valid IP scan results in a CSV file.
// The file is processed in streaming mode with O(1) memory usage.
func CountResultIPs(path string) (int64, error) {
	var count int64
	err := filemanager.StreamCSV(path, resultCSVConfig, func(rec []string) error {
		if _, ok := parseResultIP(rec); ok {
			count++
		}
		return nil
	})

	return count, err
}

// parseResultIP parses a CSV record into a ResultIPScan.
// Expected format: ip, latency
func parseResultIP(rec []string) (ResultIPScan, bool) {
	if len(rec) == 0 {
		return ResultIPScan{}, false
	}

	ip := net.ParseIP(strings.TrimSpace(rec[0]))
	if ip == nil {
		return ResultIPScan{}, false
	}

	latency := DefaultFallbackLatency
	if len(rec) >= 2 {
		if d, err := time.ParseDuration(strings.TrimSpace(rec[1])); err == nil {
			latency = d
		}
	}

	return NewResultIPScan(ip.String(), latency), true
}

// ListResultFiles lists result files by type.
// It does NOT count IPs for performance reasons.
func ListResultFiles(searchType ResultType) ([]ResultFile, error) {
	dirs := resolveResultDirs(searchType)
	if len(dirs) == 0 {
		return nil, nil
	}

	var out []ResultFile

	for _, d := range dirs {
		entries, err := os.ReadDir(d.dir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			info, err := e.Info()
			if err != nil {
				continue
			}

			out = append(out, ResultFile{
				Name:        e.Name(),
				SizeBytes:   info.Size(),
				CreatedTime: info.ModTime(),
				Type:        d.rType,
				IPCount:     -1,
			})
		}
	}

	return out, nil
}

type resultDir struct {
	dir   string
	rType ResultType
}

func resolveResultDirs(searchType ResultType) []resultDir {
	all := []resultDir{
		{ICMPResultDir, ResultICMP},
		{TCPResultDir, ResultTCP},
		{HTTPResultDir, ResultHTTP},
		{XRAYResultDir, ResultXRAY},
	}

	if searchType == ResultAll {
		return all
	}

	for _, d := range all {
		if d.rType == searchType {
			return []resultDir{d}
		}
	}

	return nil
}
