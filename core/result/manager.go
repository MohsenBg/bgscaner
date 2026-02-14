package result

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ResultWriter struct {
	deltaFlushInterval time.Duration
	mergeFlushInterval time.Duration

	mu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	deltaFile   *os.File
	deltaWriter *bufio.Writer
	deltaCSV    *csv.Writer
	deltaPath   string

	resultPath string

	input chan ResultIPScan
}

func NewResultWriter(
	resultPath string,
	deltaFlushInterval,
	mergeFlushInterval time.Duration,
	ctx context.Context,
) (*ResultWriter, error) {

	if deltaFlushInterval <= 0 {
		deltaFlushInterval = 2 * time.Second
	}
	if mergeFlushInterval <= 0 {
		mergeFlushInterval = 5 * time.Second
	}

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

	return &ResultWriter{
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
	}, nil
}

func (w *ResultWriter) Start() {
	w.wg.Add(2)
	go w.writeLoop()
	go w.mergeLoop()
}

func (w *ResultWriter) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *ResultWriter) Write(r ResultIPScan) {
	select {
	case <-w.ctx.Done():
		return
	default:
	}

	select {
	case w.input <- r:
	case <-w.ctx.Done():
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
	defer timer.Stop()

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
	_ = w.deltaCSV.Write(r.EncodeCSV())
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

