package result

import (
	"bgscan/internal/core/filemanager"
	"io"
)

// LoadResultIP streams valid scan results from CSV into a channel.
func LoadResultIP(path string, out chan<- IPScanResult) error {
	return filemanager.StreamCSV(path, csvConfig, func(rec []string) error {
		r, ok := ParseRecord(rec)
		if !ok {
			return nil
		}

		out <- r
		return nil
	})
}

// CountResultIPs counts valid records in streaming mode.
func CountResultIPs(path string) (int64, error) {
	return Count(path)
}

// LoadAll loads the entire result file into memory.
// maxIPs = -1 loads all records.
func LoadAll(path string, maxIPs int64) ([]IPScanResult, error) {
	results := make([]IPScanResult, 0, 1024)
	var c int64

	err := ReadCSV(path, func(r IPScanResult) error {
		if maxIPs >= 0 && c >= maxIPs {
			return io.EOF
		}

		results = append(results, r)
		c++
		return nil
	})

	if err == io.EOF {
		err = nil
	}

	return results, err
}
