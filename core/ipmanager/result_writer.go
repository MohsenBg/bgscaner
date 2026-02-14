package ipmanager

import (
	"bgscan/core/filemanager"
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type ResultWriter struct {
	// How often delta (append-only, unsorted) results are flushed to disk
	deltaFlushInterval time.Duration
	// How often the scheduler attempts to merge (snapshot+merge) into the sorted main file
	mergeFlushInterval time.Duration

	// Protects all delta file write/flush/close/rename operations
	mu sync.Mutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Delta (append-only) file state
	deltaFile   *os.File
	deltaWriter *bufio.Writer
	deltaCSV    *csv.Writer
	deltaPath   string // absolute path

	// Destination main file (sorted by Latency asc, tie IP asc)
	resultPath string

	// Channel for incoming scan results
	input chan ResultIPScan
}

func NewResultWriter(resultPath string, deltaFlushInterval, mergeFlushInterval time.Duration, ctx context.Context) (*ResultWriter, error) {
	if deltaFlushInterval <= 0 {
		deltaFlushInterval = 2 * time.Second
	}
	if mergeFlushInterval <= 0 {
		mergeFlushInterval = 5 * time.Second
	}

	// Create delta file in same directory as resultPath for consistent fsync/rename semantics
	dir := filepath.Dir(resultPath)
	base := filepath.Base(resultPath)
	deltaFile, err := os.CreateTemp(dir, "delta_"+base+".")
	if err != nil {
		return nil, err
	}
	deltaPath, _ := filepath.Abs(deltaFile.Name())

	bw := bufio.NewWriterSize(deltaFile, DefaultBufferSize)
	cw := csv.NewWriter(bw)
	cw.Comma = ','

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	w := &ResultWriter{
		resultPath:         resultPath,
		deltaFlushInterval: deltaFlushInterval,
		mergeFlushInterval: mergeFlushInterval,
		deltaFile:          deltaFile,
		deltaWriter:        bw,
		deltaCSV:           cw,
		deltaPath:          deltaPath,
		ctx:                ctx,
		cancel:             cancel,
		input:              make(chan ResultIPScan, DefaultChanSize),
	}
	return w, nil
}

func (w *ResultWriter) Start() {
	// One goroutine for writing, one for merge scheduling
	w.wg.Add(2)
	go w.writeLoop()
	go w.mergeLoop()
}

// Signal both loops to stop;
func (w *ResultWriter) Stop() error {
	w.cancel()
	w.wg.Wait()
	return nil
}

func (w *ResultWriter) Write(r ResultIPScan) {
	// Fast path: if the writer is already stopped, drop immediately.
	select {
	case <-w.ctx.Done():
		return
	default:
	}

	// Slow path: attempt to enqueue, but abort if shutdown happens while waiting.
	select {
	case w.input <- r:
	case <-w.ctx.Done():
		return
	}
}

func (w *ResultWriter) writeLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.deltaFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case r := <-w.input:
			w.append(r)

		case <-ticker.C:
			w.flushAndSync()

		case <-w.ctx.Done():
			// Context canceled: drain what's already enqueued
			for {
				select {
				case r := <-w.input:
					w.append(r)
				default:
					w.flushAndClose()
					return
				}
			}
		}
	}
}

func (w *ResultWriter) mergeLoop() {
	defer w.wg.Done()

	timer := time.NewTimer(w.mergeFlushInterval)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-timer.C:
			_ = w.mergeOnce()

			if w.ctx.Err() != nil {
				return
			}
			timer.Reset(w.mergeFlushInterval)
		}
	}
}

func (w *ResultWriter) append(r ResultIPScan) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.deltaCSV.Write([]string{r.IP, r.Latency.String()})
}

func (w *ResultWriter) flushAndSync() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.deltaCSV.Flush()
	_ = w.deltaWriter.Flush()
	_ = w.deltaFile.Sync()
}

func (w *ResultWriter) flushAndClose() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.deltaCSV.Flush()
	_ = w.deltaWriter.Flush()
	_ = w.deltaFile.Sync()
	_ = w.deltaFile.Close()
}

// mergeOnce merges either an existing snapshot (oldest) or creates a new snapshot, then merges it.
// On any failure during merge, the snapshot is kept for retry. Only on success it is deleted.
func (w *ResultWriter) mergeOnce() error {
	// 1) If there are pending snapshots (from previous failures/crashes), merge the oldest first.
	if snap, _ := w.findOldestSnapshot(); snap != "" {
		return w.mergeSnapshot(snap)
	}

	// 2) Otherwise, create a new snapshot from the current delta via atomic rename (O(1)).
	snapshotPath, err := w.snapshotDelta()
	if err != nil {
		return err
	}
	if snapshotPath == "" {
		// Nothing to merge (delta empty)
		return nil
	}
	return w.mergeSnapshot(snapshotPath)
}

// snapshotDelta performs the following steps:
//
//  1. Flush buffers and fsync the delta file
//  2. If the delta file is empty, return ""
//  3. Close the delta file (required on Windows)
//  4. Atomically rename delta -> snapshot
//  5. Fsync the parent directory
//  6. Recreate a fresh delta file and writers
func (w *ResultWriter) snapshotDelta() (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush CSV and buffered writer
	w.deltaCSV.Flush()
	if err := w.deltaCSV.Error(); err != nil {
		return "", err
	}

	if err := w.deltaWriter.Flush(); err != nil {
		return "", err
	}
	if err := w.deltaFile.Sync(); err != nil {
		return "", err
	}

	info, err := w.deltaFile.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	// Close before rename (required on Windows)
	if err := w.deltaFile.Close(); err != nil {
		return "", err
	}

	// Unique snapshot name to avoid collisions across crashes
	snapshotPath := w.deltaPath + ".snapshot." +
		strconv.FormatInt(time.Now().UnixNano(), 10)

	// Atomic rename (O(1))
	if err := os.Rename(w.deltaPath, snapshotPath); err != nil {
		// Best-effort recovery to keep writers operational
		_ = w.reopenDeltaUnsafe()
		return "", err
	}

	// Persist rename (best-effort durability)
	if err := syncDir(filepath.Dir(w.deltaPath)); err != nil {
		// TODO: Log
	}

	// Recreate fresh delta at original path
	if err := w.reopenDeltaUnsafe(); err != nil {
		return "", err
	}

	return snapshotPath, nil
}

// reopenDeltaUnsafe recreates the delta file and writers.
// Caller must hold w.mu.
func (w *ResultWriter) reopenDeltaUnsafe() error {
	f, err := os.OpenFile(
		w.deltaPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return err
	}

	w.deltaFile = f
	w.deltaWriter = bufio.NewWriterSize(f, DefaultBufferSize)

	w.deltaCSV = csv.NewWriter(w.deltaWriter)
	w.deltaCSV.Comma = ','

	return nil
}

// findOldestSnapshot returns the lexicographically oldest snapshot file
// matching deltaPath + ".snapshot.*".
func (w *ResultWriter) findOldestSnapshot() (string, error) {
	pattern := w.deltaPath + ".snapshot.*"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", nil
	}

	sort.Strings(matches)
	return matches[0], nil
}

// mergeSnapshot loads+sorts the snapshot and merges it into resultPath via tmp+atomic replace.
// On success deletes the snapshot. On failure keeps snapshot for retry.
func (w *ResultWriter) mergeSnapshot(snapshotPath string) error {
	delta, err := loadAndSortDelta(snapshotPath)
	if err != nil {
		return err
	}

	// Safe to delete empty snapshot
	if len(delta) == 0 {
		_ = os.Remove(snapshotPath)
		return nil
	}

	// Keep snapshot for retry
	if err := mergeStreaming(w.resultPath, delta); err != nil {
		return err
	}

	// Success: remove snapshot
	_ = os.Remove(snapshotPath)
	return nil
}

func loadAndSortDelta(path string) ([]ResultIPScan, error) {
	out := make([]ResultIPScan, 0, 1024)

	err := filemanager.StreamCSV(path, filemanager.CSVConfig{Comma: ','}, func(rec []string) error {
		r, ok := parseResultIP(rec)
		if ok {
			out = append(out, r)
		}
		return nil
	})
	if err != nil {
		// If file is missing (e.g., deleted externally), treat as empty
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Less(out[j])
	})

	return out, nil
}

func mergeStreaming(resultPath string, delta []ResultIPScan) error {
	if len(delta) == 0 {
		return nil
	}

	tmpPath := resultPath + ".tmp"

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(out, DefaultBufferSize)
	cw := csv.NewWriter(bw)
	cw.Comma = ','

	writeDelta := func(i int) error {
		return cw.Write(delta[i].EncodeCSV())
	}

	fail := func(err error) error {
		_ = out.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	// --- merge ---
	if filemanager.CheckFileExists(resultPath) {
		i := 0

		err = filemanager.StreamCSV(
			resultPath,
			filemanager.CSVConfig{Comma: ','},
			func(rec []string) error {
				mainRec, ok := parseResultIP(rec)
				if !ok {
					return nil
				}

				for i < len(delta) && delta[i].Less(mainRec) {
					if err := writeDelta(i); err != nil {
						return err
					}
					i++
				}

				return cw.Write(mainRec.EncodeCSV())
			},
		)
		if err != nil {
			return fail(err)
		}

		// write remaining delta
		for ; i < len(delta); i++ {
			if err := writeDelta(i); err != nil {
				return fail(err)
			}
		}
	} else {
		// result file does not exist → delta becomes result
		for i := range delta {
			if err := writeDelta(i); err != nil {
				return fail(err)
			}
		}
	}

	// --- flush & close ---
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fail(err)
	}
	if err := bw.Flush(); err != nil {
		return fail(err)
	}
	if err := out.Sync(); err != nil {
		return fail(err)
	}
	if err := out.Close(); err != nil {
		return fail(err)
	}

	// --- atomic replace ---
	if err := replaceFile(tmpPath, resultPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	_ = syncDir(filepath.Dir(resultPath))
	return nil

}

type csvFunc func([]string) error

func replaceFile(src, dst string) error {
	// Try direct rename first
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// On Windows (and some FS), rename fails if target exists: delete and retry
	_ = os.Remove(dst)
	return os.Rename(src, dst)
}

// syncDir fsyncs a directory to persist rename/creation on POSIX. On Windows this is a no-op.
func syncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer df.Close()
	return df.Sync()
}
