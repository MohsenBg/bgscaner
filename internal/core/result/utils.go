package result

import (
	"bgscan/internal/core/filemanager"
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
)

// replaceFile atomically replaces the destination file with the source file.
//
// On Unix-like systems os.Rename provides atomic replacement.
// On Windows, Rename fails if the destination already exists,
// so we fall back to removing the destination first.
func replaceFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Windows fallback: remove destination then retry rename
	_ = os.Remove(dst)
	return os.Rename(src, dst)
}

// syncDir flushes directory metadata to disk.
//
// This ensures file creations/renames are persisted after a crash.
// On some platforms (notably Windows) this may effectively be a no‑op.
func syncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer df.Close()

	return df.Sync()
}

// createDeltaFile creates a temporary delta file used for buffering scan results.
//
// The returned objects include:
//   - *os.File        : underlying file
//   - *bufio.Writer   : buffered writer for performance
//   - *csv.Writer     : CSV encoder used by the result writer
//   - string          : absolute path to the delta file
//
// The file is created inside the result directory to ensure atomic
// rename operations remain on the same filesystem.
func createDeltaFile(resultPath string, bufferSize int) (
	*os.File,
	*bufio.Writer,
	*csv.Writer,
	string,
	error,
) {
	dir := filepath.Dir(resultPath)

	// Ensure result directory exists
	dir, err := filemanager.GetOrCreateBaseDir(dir)
	if err != nil {
		return nil, nil, nil, "", err
	}

	base := filepath.Base(resultPath)

	// Create temporary delta file
	file, err := os.CreateTemp(dir, "delta_"+base+".")
	if err != nil {
		return nil, nil, nil, "", err
	}

	// Resolve absolute path for consistency
	deltaPath, err := filepath.Abs(file.Name())
	if err != nil {
		file.Close()
		return nil, nil, nil, "", err
	}

	// Setup buffered + CSV writers
	bw := bufio.NewWriterSize(file, bufferSize)

	cw := csv.NewWriter(bw)
	cw.Comma = ','

	return file, bw, cw, deltaPath, nil
}
