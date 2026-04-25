package result

import (
	"bgscan/internal/core/filemanager"
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// mergeResults merges a batch of scan results into the main result file.
//
// Guarantees:
//   - Sorted output
//   - Duplicate replacement (new records override existing ones)
//   - Streaming merge with constant memory usage
//   - Atomic file replacement
//   - Crash‑safe writes (fsync before rename)
//
// The merge is implemented as a classic merge‑sort merge phase between the
// existing result file and the new batch of records.
func mergeResults(resultPath string, ips []IPScanResult) error {
	if len(ips) == 0 {
		return nil
	}

	// Ensure delta results are sorted before merging.
	sort.Slice(ips, func(i, j int) bool { return ips[i].Less(ips[j]) })

	tmpPath := resultPath + ".tmp"

	dir := filepath.Dir(tmpPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(out, DefaultBatchSize)
	cw := csv.NewWriter(bw)

	cleanup := func(e error) error {
		_ = out.Close()
		_ = os.Remove(tmpPath)
		return e
	}

	if filemanager.CheckFileExists(resultPath) {
		if err := mergeWithExisting(resultPath, ips, cw); err != nil {
			return cleanup(err)
		}
	} else {
		if err := writeIPs(ips, cw); err != nil {
			return cleanup(err)
		}
	}

	if err := finalizeFile(cw, bw, out); err != nil {
		return cleanup(err)
	}

	if err := replaceFile(tmpPath, resultPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return syncDir(filepath.Dir(resultPath))
}

// mergeWithExisting merges delta records with an already existing sorted
// result file.
//
// The implementation performs a streaming merge similar to the merge phase
// of merge‑sort. Only one record from the existing file is kept in memory at
// any time, allowing the function to process extremely large result files
// with minimal memory usage.
//
// If a duplicate record exists, the delta record replaces the original.
func mergeWithExisting(resultPath string, delta []IPScanResult, cw *csv.Writer) error {
	i := 0

	err := ReadCSV(resultPath, func(mainRec IPScanResult) error {

		// Write all delta entries that should appear before the current record.
		for i < len(delta) && delta[i].Less(mainRec) {
			if err := cw.Write(delta[i].ToRecord()); err != nil {
				return err
			}
			i++
		}

		// Replace existing record if a duplicate appears in delta.
		if i < len(delta) && delta[i].Equal(mainRec) {
			if err := cw.Write(delta[i].ToRecord()); err != nil {
				return err
			}
			i++
			return nil
		}

		// Otherwise preserve the existing record.
		return cw.Write(mainRec.ToRecord())
	})

	if err != nil {
		return err
	}

	// Write any remaining delta entries.
	for ; i < len(delta); i++ {
		if err := cw.Write(delta[i].ToRecord()); err != nil {
			return err
		}
	}

	return nil
}

// writeIPs writes a full slice of scan results to the CSV writer.
func writeIPs(ips []IPScanResult, cw *csv.Writer) error {
	for i := range ips {
		if err := cw.Write(ips[i].ToRecord()); err != nil {
			return err
		}
	}
	return nil
}

// finalizeFile flushes all buffered data and synchronizes the file to disk
// before closing it. This ensures the temporary file is fully persisted prior
// to the atomic rename step.
func finalizeFile(cw *csv.Writer, bw *bufio.Writer, out *os.File) error {
	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	return out.Close()
}
