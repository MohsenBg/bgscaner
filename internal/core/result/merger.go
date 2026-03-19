package result

import (
	"bgscan/internal/core/filemanager"
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
)

// mergeSnapshot merges a snapshot file into the main result file.
//
// The snapshot represents a batch of results written earlier by the writer.
// After a successful merge the snapshot file is deleted.
//
// If the merge fails the snapshot is intentionally kept so the operation
// can be retried later.
func mergeSnapshot(resultPath, snapshotPath string) error {
	delta, err := loadSnapshot(snapshotPath)
	if err != nil {
		return err
	}

	// Nothing to merge
	if len(delta) == 0 {
		if err := os.Remove(snapshotPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	// Perform merge
	if err := mergeResults(resultPath, delta); err != nil {
		return err // keep snapshot for retry
	}

	// Remove snapshot after successful merge
	if err := os.Remove(snapshotPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// mergeResults performs a streaming merge of delta records into the result file.
//
// The algorithm guarantees:
//   - sorted output
//   - duplicate replacement
//   - atomic file replacement
//   - crash-safe writes
func mergeResults(resultPath string, delta []IPScanResult) error {
	if len(delta) == 0 {
		return nil
	}

	// Ensure delta records are sorted before merging
	sort.Slice(delta, func(i, j int) bool {
		return delta[i].Less(delta[j])
	})

	tmpPath := resultPath + ".tmp"

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(out, DefaultBufferSize)
	cw := csv.NewWriter(bw)
	cw.Comma = ','

	cleanup := func(err error) error {
		_ = out.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	// Merge logic
	if filemanager.CheckFileExists(resultPath) {
		if err := mergeWithExisting(resultPath, delta, cw); err != nil {
			return cleanup(err)
		}
	} else {
		if err := writeDelta(delta, cw); err != nil {
			return cleanup(err)
		}
	}

	// Finalize writes
	if err := finalizeFile(cw, bw, out); err != nil {
		return cleanup(err)
	}

	// Atomic replace
	if err := replaceFile(tmpPath, resultPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	// Ensure directory metadata is persisted
	return syncDir(filepath.Dir(resultPath))
}

// mergeWithExisting merges delta records with an existing result file.
//
// The function performs a streaming merge similar to the merge phase
// of merge-sort, ensuring minimal memory usage.
func mergeWithExisting(resultPath string, delta []IPScanResult, cw *csv.Writer) error {
	i := 0

	err := ReadCSV(resultPath, func(mainRec IPScanResult) error {

		// Write delta entries that come before the current main record
		for i < len(delta) && delta[i].Less(mainRec) {
			if err := cw.Write(delta[i].ToRecord()); err != nil {
				return err
			}
			i++
		}

		// Replace duplicate entries
		if i < len(delta) && delta[i].Equal(mainRec) {
			if err := cw.Write(delta[i].ToRecord()); err != nil {
				return err
			}
			i++
			return nil
		}

		return cw.Write(mainRec.ToRecord())
	})

	if err != nil {
		return err
	}

	// Write remaining delta records
	for ; i < len(delta); i++ {
		if err := cw.Write(delta[i].ToRecord()); err != nil {
			return err
		}
	}

	return nil
}

// writeDelta writes delta records directly when the result file
// does not yet exist.
func writeDelta(delta []IPScanResult, cw *csv.Writer) error {
	for i := range delta {
		if err := cw.Write(delta[i].ToRecord()); err != nil {
			return err
		}
	}
	return nil
}

// finalizeFile flushes buffers and ensures the data is fully synced
// to disk before closing the file.
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
