package engine

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/iplist"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ScanHooks defines optional lifecycle callbacks for the scanning engine.
//
// All hooks are optional and can be nil. They allow external components
// (CLI, UI, metrics systems, etc.) to observe scan progress and results
// without coupling them directly to the engine logic.
type ScanHooks struct {

	// OnProgress is called periodically with a snapshot of scan progress.
	OnProgress func(p Progress)

	// OnSuccess is called whenever a scan succeeds.
	OnSuccess func(r result.IPScanResult)

	// OnScanEnd is called once after the scan finishes completely.
	OnScanEnd func()

	// OnError is called when a recoverable engine error occurs.
	OnError func(err error)
}

// RunScan is the main entry point of the scanning engine.
//
// It orchestrates the full scan lifecycle:
//
//  1. Count total IPs to scan
//  2. Spawn worker goroutines
//  3. Stream IPs from the input source
//  4. Execute probes concurrently
//  5. Write results asynchronously
//  6. Report scan progress periodically
//
// The function blocks until the scan finishes or the context is cancelled.
func RunScan(
	ctx context.Context,
	workers int,
	rate int,
	ipFile string,
	writer *result.Writer,
	prb probe.Probe,
	hooks ScanHooks,
	pause *PauseController,
) {

	// Determine total scan size for progress reporting.
	total, err := iplist.CountActiveIPs(ipFile)
	if err != nil {
		if hooks.OnError != nil {
			hooks.OnError(err)
		}
		return
	}

	workers = min(workers, int(total))
	// Channel used to distribute IPs to workers.
	ips := make(chan string, workers*2)

	// Channel used by workers to emit scan results.
	results := make(chan result.IPScanResult)

	// Ensures the writer goroutine flushes all results before returning.
	var writerDone sync.WaitGroup
	writerDone.Add(1)

	// Used to stop the progress reporting goroutine.
	progressDone := make(chan struct{})

	// --- Rate limiter setup
	// A ticker is used to control how often workers are allowed
	// to start a scan operation.
	var rateCh <-chan time.Time

	if rate > 0 {
		interval := time.Second / time.Duration(rate)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		rateCh = ticker.C
	} else {
		// If rate limiting is disabled we create a closed channel.
		// Reading from a closed channel returns immediately,
		// effectively disabling throttling.
		always := make(chan time.Time)
		close(always)

		rateCh = always
	}

	// --- Scan statistics
	var (
		processed int64 // total processed IPs
		succeed   int64 // successful scans
		start     = time.Now()

		// Total time spent paused (nanoseconds).
		pausedTime int64
	)

	// --- Progress reporting goroutine
	//
	// Periodically emits scan progress to the provided hook.
	// This runs independently of worker execution.
	if hooks.OnProgress != nil {

		go func() {

			interval := config.Get().General.StatusInterval.Duration()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {

				case <-progressDone:
					return

				case <-ctx.Done():
					return

				case <-ticker.C:

					// When paused we accumulate pause time instead
					// of reporting progress.
					if pause != nil && pause.IsPaused() {
						atomic.AddInt64(&pausedTime, int64(interval))
						continue
					}

					reportProgress(
						start,
						time.Duration(atomic.LoadInt64(&pausedTime)),
						total,
						&processed,
						&succeed,
						hooks.OnProgress,
					)
				}
			}
		}()
	}

	// --- Result writer goroutine
	//
	// Responsible for persisting scan results asynchronously.
	// This prevents disk I/O from blocking worker goroutines.
	go func() {

		defer writerDone.Done()

		writer.Start()

		for r := range results {
			writer.Write(r)
		}

		writer.Stop()
	}()

	// --- Worker pool
	//
	// Workers consume IPs from the `ips` channel and execute probes.
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()

			runWorker(
				ctx,
				ips,
				results,
				rateCh,
				prb,
				hooks,
				pause,
				&processed,
				&succeed,
			)
		}()
	}

	// --- IP producer
	//
	// Streams IPs from the input file into the worker queue.
	go func() {

		defer close(ips)

		if err := iplist.StreamActiveIPs(ctx, ipFile, ips); err != nil {

			logger.CoreError("StreamActiveIPs: %v", err)

			if hooks.OnError != nil {
				hooks.OnError(err)
			}
		}
	}()

	// Wait for workers to finish processing.
	wg.Wait()

	// Signal writer that no more results will arrive.
	close(results)

	// Wait until all results are flushed to disk.
	writerDone.Wait()

	// Stop progress reporter.
	close(progressDone)

	err = prb.Close()
	if err != nil {
		hooks.OnError(err)
	}

	// Emit final progress snapshot.
	if hooks.OnProgress != nil {
		reportProgress(
			start,
			time.Duration(atomic.LoadInt64(&pausedTime)),
			total,
			&processed,
			&succeed,
			hooks.OnProgress,
		)
	}

	// Notify scan completion.
	if hooks.OnScanEnd != nil {
		hooks.OnScanEnd()
	}
}

// runWorker processes IPs from the queue until the queue is closed
// or the context is cancelled.
func runWorker(
	ctx context.Context,
	ips <-chan string,
	results chan<- result.IPScanResult,
	rateCh <-chan time.Time,
	prb probe.Probe,
	hooks ScanHooks,
	pause *PauseController,
	processed, succeed *int64,
) {

	for {

		// Block if scanning is paused.
		if pause != nil && !pause.Wait(ctx) {
			return
		}

		select {

		case <-ctx.Done():
			return

		case ip, ok := <-ips:

			if !ok {
				return
			}

			// Enforce rate limiting before executing probe.
			select {
			case <-rateCh:
			case <-ctx.Done():
				return
			}

			r, err := prb.Run(ctx, ip)

			// Count every processed IP.
			atomic.AddInt64(processed, 1)

			if err != nil {
				logger.CoreError("scan IP %s failed: %v", ip, err)
				continue
			}

			// Track successful scans.
			atomic.AddInt64(succeed, 1)

			// Notify success hook if present.
			if hooks.OnSuccess != nil {
				hooks.OnSuccess(*r)
			}

			// Send result to writer pipeline.
			select {
			case results <- *r:

			case <-ctx.Done():
				return
			}
		}
	}
}
