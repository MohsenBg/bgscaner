package result

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// createSnapshot converts the current delta file into a snapshot file.
//
// The operation is designed to be crash‑safe:
//
// 1. Flush CSV and buffered writers
// 2. Sync the file to disk
// 3. Atomically rename the delta file to a snapshot
// 4. Reopen a fresh delta file for continued writes
//
// If the delta file is empty, no snapshot is created and an empty
// path is returned.
func (w *Writer) createSnapshot() (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush CSV writer
	w.deltaCSV.Flush()
	if err := w.deltaCSV.Error(); err != nil {
		return "", err
	}

	// Flush buffered writer
	if err := w.deltaWriter.Flush(); err != nil {
		return "", err
	}

	// Ensure file data is persisted
	if err := w.deltaFile.Sync(); err != nil {
		return "", err
	}

	info, err := w.deltaFile.Stat()
	if err != nil {
		return "", err
	}

	// Nothing to snapshot
	if info.Size() == 0 {
		return "", nil
	}

	// File must be closed before rename on Windows
	if err := w.deltaFile.Close(); err != nil {
		return "", err
	}

	// Generate unique snapshot filename
	snapshotPath := w.deltaPath + ".snapshot." +
		strconv.FormatInt(time.Now().UnixNano(), 10)

	// Atomic rename of delta → snapshot
	if err := os.Rename(w.deltaPath, snapshotPath); err != nil {
		// Attempt recovery by reopening the delta file
		_ = w.reopenDeltaUnsafe()
		return "", err
	}

	// Persist directory metadata for crash safety
	_ = syncDir(filepath.Dir(w.deltaPath))

	// Recreate delta file so writing can continue
	if err := w.reopenDeltaUnsafe(); err != nil {
		return "", err
	}

	return snapshotPath, nil
}

// reopenDeltaUnsafe recreates the delta file and associated writers.
//
// Caller MUST already hold w.mu.
func (w *Writer) reopenDeltaUnsafe() error {
	f, err := os.OpenFile(
		w.deltaPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return err
	}

	w.deltaFile = f
	w.deltaWriter = bufio.NewWriterSize(f, w.config.BufferSize)

	w.deltaCSV = csv.NewWriter(w.deltaWriter)
	w.deltaCSV.Comma = ','

	return nil
}

// findOldestSnapshot returns the oldest snapshot associated
// with the provided delta file.
//
// Snapshot files follow the pattern:
//
//	<delta>.snapshot.<timestamp>
//
// If no snapshot exists, an empty string is returned.
func findOldestSnapshot(deltaPath string) (string, error) {
	pattern := deltaPath + ".snapshot.*"

	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", err
	}

	sort.Strings(matches)

	return matches[0], nil
}

// loadSnapshot loads all records from a snapshot file into memory
// and returns them sorted.
//
// Sorting is required because the merge algorithm expects ordered input.
func loadSnapshot(path string) ([]IPScanResult, error) {
	results := make([]IPScanResult, 0, 1024)

	err := ReadCSV(path, func(r IPScanResult) error {
		results = append(results, r)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // treat missing snapshot as empty
		}
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Less(results[j])
	})

	return results, nil
}
