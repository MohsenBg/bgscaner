package scanner

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"bgscan/pinger"
)

// ScannerConfig defines all settings for the scanner.
type ScannerConfig struct {
	Threads        int
	ShuffleIPs     bool
	StopAfterFound int
	MaxIPsToTest   int
	Port           int
	Timeout        time.Duration
	FlushInterval  int
	StatusInterval time.Duration
	Verbose        bool
	IPsFile        string
	OutputPrefix   string
	PrintSuccess   bool
	Mode           pinger.PingMode // NEW: allow selecting TCP, ICMP, HTTP
	Host           string          // for HTTP Host header
}

// RunScannerMemorySafe runs the scanner in a memory-efficient, concurrent way
func RunScannerMemorySafe(config ScannerConfig) error {
	// ------------------ shuffle IP file if requested ------------------
	inputFilePath := config.IPsFile
	if config.ShuffleIPs {
		shuffledFile, err := ShuffleFileFullyMemorySafe(config.IPsFile)
		if err != nil {
			return fmt.Errorf("shuffle failed: %v", err)
		}
		inputFilePath = shuffledFile
		defer os.Remove(shuffledFile)
	}

	// ------------------ count total IPs for progress ------------------
	total, err := CountValidIPs(inputFilePath, config.MaxIPsToTest)
	if err != nil {
		return fmt.Errorf("counting IPs failed: %v", err)
	}
	if total == 0 {
		return fmt.Errorf("no valid IPs to scan")
	}

	// ------------------ setup channels & counters ------------------
	ipChan, err := StreamIPs(inputFilePath, config.MaxIPsToTest)
	if err != nil {
		return fmt.Errorf("streaming IPs failed: %v", err)
	}

	resultChan := make(chan ScanResult, config.Threads*2)
	var scanned uint64
	var found uint64

	outputFile := getOutputFileName(config.OutputPrefix)
	doneTicker := StartProgressTracker(&scanned, &found, total, config.StatusInterval)

	// ------------------ launch workers ------------------
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case ip, ok := <-ipChan:
					if !ok {
						return
					}
					atomic.AddUint64(&scanned, 1)

					// ------------------ call unified PingHost ------------------
					okPing, latency := pinger.PingHost(ip, config.Host, config.Port, config.Timeout, config.Mode, config.Verbose)
					if okPing {
						resultChan <- ScanResult{IP: ip, Latency: latency}
					}

				case <-stopCh:
					return
				}
			}
		}()
	}

	// ------------------ close result channel when workers are done ------------------
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// ------------------ consume results ------------------
	var batch []ScanResult
	for res := range resultChan {
		atomic.AddUint64(&found, 1)
		if config.PrintSuccess {
			PrintSuccess(res)
		}

		batch = append(batch, res)
		if len(batch) >= config.FlushInterval {
			FlushResultsToFile(outputFile, batch)
			batch = batch[:0]
		}

		if config.StopAfterFound > 0 && int(found) >= config.StopAfterFound {
			close(stopCh)
		}
	}

	// stop live ticker
	close(doneTicker)

	// flush remaining results
	if len(batch) > 0 {
		FlushResultsToFile(outputFile, batch)
	}

	return nil
}

// getOutputFileName generates a unique output file name
func getOutputFileName(prefix string) string {
	if prefix == "" {
		prefix = "results"
	}
	return fmt.Sprintf("%s_%s.txt", prefix, time.Now().Format("20060102_150405"))
}
