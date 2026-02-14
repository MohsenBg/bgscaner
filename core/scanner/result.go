package scanner

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

/* ================= RESULT ================= */

type ScanResult struct {
	IP      string
	Latency time.Duration
}

// FlushResultsToFile merges with existing file (if any), removes duplicates, sorts, and writes back
func FlushResultsToFile(filename string, results []ScanResult) error {
	// map to store unique IPs
	unique := make(map[string]ScanResult)

	// load existing results from file if it exists
	if _, err := os.Stat(filename); err == nil {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			ip := parts[0]
			latencyStr := parts[1]
			latency, err := time.ParseDuration(latencyStr)
			if err != nil {
				continue
			}
			if existing, ok := unique[ip]; ok {
				if latency < existing.Latency {
					unique[ip] = ScanResult{IP: ip, Latency: latency}
				}
			} else {
				unique[ip] = ScanResult{IP: ip, Latency: latency}
			}
		}
		file.Close()
	}

	// merge current results
	for _, r := range results {
		if existing, ok := unique[r.IP]; ok {
			if r.Latency < existing.Latency {
				unique[r.IP] = r
			}
		} else {
			unique[r.IP] = r
		}
	}

	// convert map back to slice
	finalResults := make([]ScanResult, 0, len(unique))
	for _, r := range unique {
		finalResults = append(finalResults, r)
	}

	// sort by latency
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].Latency < finalResults[j].Latency
	})

	// write back to file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, r := range finalResults {
		fmt.Fprintf(f, "%s %v\n", r.IP, r.Latency)
	}

	return nil
}

func PrintSuccess(res ScanResult) {
	if !SKIP_PROGRESS_BAR_REPLACE.Load() {
		fmt.Print("\033[2A\033[J")
	}

	SKIP_PROGRESS_BAR_REPLACE.Store(true)
	fmt.Printf(
		"%s\n%s -> %v\n%s\n",
		color.CyanString("=========================="),
		color.GreenString(res.IP),
		color.GreenString(res.Latency.String()),
		color.CyanString("=========================="),
	)
}
