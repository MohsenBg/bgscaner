package result

import (
	"bufio"
	"context"
	"encoding/csv"
	"os"
	"sync"
	"time"
)

// Writer provides a concurrent-safe result writer for IP scan results.
//
// Architecture overview:
//
//	Scanner Workers
//	      │
//	      ▼
//	  channel buffer
//	      │
//	      ▼
//	  delta file (append-only log)
//	      │
//	      ▼
//	 periodic snapshot
//	      │
//	      ▼
//	 merge into result file
//
// Results are first appended to a temporary delta file to avoid expensive
// rewrites of the final result file. Periodically the delta file is turned
// into a snapshot and merged into the main result file.
//
// This design allows high-throughput writes while keeping disk IO efficient
// and crash-safe.
type Writer struct {
	config Config

	// protects access to file writers
	mu sync.Mutex

	// lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// delta file components
	deltaFile   *os.File
	deltaWriter *bufio.Writer
	deltaCSV    *csv.Writer
	deltaPath   string

	// final result file path
	resultPath string

	// incoming scan results
	input chan IPScanResult
}

// NewWriter initializes a new result writer.
//
// The writer maintains an internal delta file where new scan results
// are appended. Periodically the delta file will be merged into the
// final result file.
func NewWriter(resultPath string, cfg Config, ctx context.Context) (*Writer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg.Normalize()

	deltaFile, bw, cw, deltaPath, err := createDeltaFile(resultPath, cfg.BufferSize)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Writer{
		config:      cfg,
		resultPath:  resultPath,
		deltaFile:   deltaFile,
		deltaWriter: bw,
		deltaCSV:    cw,
		deltaPath:   deltaPath,
		ctx:         ctx,
		cancel:      cancel,
		input:       make(chan IPScanResult, cfg.ChanSize),
	}, nil
}

// Start launches background workers responsible for writing results
// and periodically merging delta snapshots.
func (w *Writer) Start() {
	// clean up snapshots
	w.cleanupSnapshot()
	w.wg.Add(2)

	go w.writeLoop()
	go w.mergeLoop()
}

// Stop gracefully shuts down the writer.
//
// Shutdown procedure:
//   - cancel context
//   - drain remaining results
//   - flush buffers
//   - close files
//   - wait for workers
func (w *Writer) Stop() error {
	w.cancel()
	w.wg.Wait()
	return nil
}

// Write queues a scan result for asynchronous writing.
//
// If the writer is shutting down the result is ignored.
func (w *Writer) Write(r IPScanResult) {
	select {
	case <-w.ctx.Done():
		return
	case w.input <- r:
	}
}

// writeLoop consumes incoming results and writes them to the delta file.
// It also periodically flushes buffers for durability.
func (w *Writer) writeLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.DeltaFlushInterval)
	defer ticker.Stop()

	for {
		select {

		case r := <-w.input:
			w.appendRecord(r)

		case <-ticker.C:
			w.flushBuffers()

		case <-w.ctx.Done():
			w.drainAndClose()
			return
		}
	}
}

// mergeLoop periodically merges delta snapshots into the main result file.
func (w *Writer) mergeLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.MergeFlushInterval)
	defer ticker.Stop()

	for {
		select {

		case <-w.ctx.Done():
			return

		case <-ticker.C:
			_ = w.mergeOnce()
		}
	}
}

// appendRecord writes a single result record into the delta CSV file.
func (w *Writer) appendRecord(r IPScanResult) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_ = w.deltaCSV.Write(r.ToRecord())
}

// flushBuffers flushes all buffered data to disk.
func (w *Writer) flushBuffers() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.deltaCSV.Flush()
	_ = w.deltaWriter.Flush()
	_ = w.deltaFile.Sync()
}

// drainAndClose drains remaining results before shutting down.
func (w *Writer) drainAndClose() {
	for {
		select {
		case r := <-w.input:
			w.appendRecord(r)
		default:
			// convert remaining delta to snapshot
			if snap, err := w.createSnapshot(); err == nil && snap != "" {
				_ = mergeSnapshot(w.resultPath, snap)
			}

			// close delta
			w.closeFile()

			// merge any remaining snapshots
			for {
				snap, _ := findOldestSnapshot(w.deltaPath)
				if snap == "" {
					break
				}
				_ = mergeSnapshot(w.resultPath, snap)
			}
			return
		}
	}
}

// closeFile safely flushes and closes the delta file.
func (w *Writer) closeFile() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.deltaCSV.Flush()
	_ = w.deltaWriter.Flush()
	_ = w.deltaFile.Sync()
	_ = w.deltaFile.Close()

	os.Remove(w.deltaPath)
}

// mergeOnce executes a single merge cycle.
func (w *Writer) mergeOnce() error {

	// if snapshot already exists merge it first
	if snap, _ := findOldestSnapshot(w.deltaPath); snap != "" {
		return mergeSnapshot(w.resultPath, snap)
	}

	// otherwise create a snapshot from current delta file
	snapshotPath, err := w.createSnapshot()
	if err != nil || snapshotPath == "" {
		return err
	}

	return mergeSnapshot(w.resultPath, snapshotPath)
}

func (w *Writer) cleanupSnapshot() {
	for {
		snap, _ := findOldestSnapshot(w.deltaPath)
		if snap == "" {
			break
		}
		os.Remove(snap)
	}
}
