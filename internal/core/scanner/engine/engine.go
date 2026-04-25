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

//
// ────────────────────────────────────────────────────────────────────────────────
//  runScan (Main Stage Orchestrator)
// ────────────────────────────────────────────────────────────────────────────────
//

// RunScan orchestrates the full lifecycle of a **single scan stage**.
//
// It performs the following operations:
//
//  1. Count total active IPs.
//  2. Create worker pool sized by IP count and cfg.Workers.
//  3. Initialize producer → worker pool → result writer pipeline.
//  4. Apply per-stage rate limiting.
//  5. Connect progress reporter with pause/resume support.
//  6. Ensure graceful teardown and final progress snapshot.
//
// This function blocks until the entire stage completes or the context is cancelled.
func RunScan(
	ctx context.Context,
	input string,
	maxIp int,
	cfg ScanConfig,
	pause *PauseController,
) {
	total, err := iplist.CountActiveIPs(input)
	if err != nil {
		cfg.Hooks.callOnError(err)
		return
	}

	// Worker count must not exceed number of IPs.
	workers := min(cfg.Workers, int(total))

	ips := make(chan string, workers*2)
	results := make(chan result.IPScanResult)

	var writerDone sync.WaitGroup
	writerDone.Add(1)

	progressDone := make(chan struct{})

	rateCh := makeRateCh(cfg.Rate)

	// Atomic counters used by progress reporter.
	var (
		processed uint64
		succeed   uint64
		start     = time.Now()
	)

	//
	// ─── Progress Reporter ───────────────────────────────────────────────
	//

	if cfg.Hooks.OnProgress != nil {
		go runProgressReporter(
			ctx,
			progressDone,
			pause,
			start,
			total,
			&processed,
			&succeed,
			cfg.Hooks.OnProgress,
		)
	}

	//
	// ─── Writer Goroutine ───────────────────────────────────────────────
	//
	// Consumes results from workers and writes them using stage.Writer.
	go func() {
		defer writerDone.Done()
		cfg.Writer.Start()
		for r := range results {
			cfg.Writer.Write(r)
		}
		cfg.Writer.Stop()
	}()

	//
	// ─── Worker Pool ─────────────────────────────────────────────────────
	//
	var wg sync.WaitGroup
	cfg.Probe.Init(ctx)
	for range workers {
		wg.Go(func() {
			runScanWorker(ctx, ips, results, rateCh, cfg.Probe, cfg.Hooks, pause, &processed, &succeed)
		})
	}

	//
	// ─── IP Producer ─────────────────────────────────────────────────────
	//
	// Streams IPs from the input file asynchronously.
	go func() {
		defer close(ips)
		if err := iplist.StreamActiveIPs(ctx, input, maxIp, ips); err != nil {
			logger.CoreError("StreamActiveIPs: %v", err)
			cfg.Hooks.callOnError(err)
		}
	}()

	//
	// ─── Shutdown Sequence ───────────────────────────────────────────────
	//

	wg.Wait()                 // workers done
	cfg.Hooks.callOnScanEnd() // call scan end hook
	close(results)            // stop sending new results
	writerDone.Wait()         // wait for writer
	close(progressDone)       // stop progress reporter

	if err := cfg.Probe.Close(); err != nil {
		cfg.Hooks.callOnError(err)
	}

	// Final progress update
	if cfg.Hooks.OnProgress != nil {
		reportProgress(
			start,
			pause.PausedDuration(),
			total, &processed, &succeed,
			cfg.Hooks.OnProgress,
		)
	}

	cfg.Hooks.callOnScanEnd()
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Worker
// ────────────────────────────────────────────────────────────────────────────────
//

// runScanWorker processes IP addresses until:
//   - the input channel is closed,
//   - the context is cancelled,
//   - or the PauseController returns false.
//
// Responsibilities:
//   - Wait for pause/resume cycles.
//   - Receive IP from queue.
//   - Apply per-IP rate limiting.
//   - Execute probe.Run().
//   - Update atomic counters.
//   - Push successful results to results channel.
func runScanWorker(
	ctx context.Context,
	ips <-chan string,
	results chan<- result.IPScanResult,
	rateCh <-chan time.Time,
	prb probe.Probe,
	hooks ScanHooks,
	pause *PauseController,
	processed *uint64,
	succeed *uint64,
) {
	for {
		// Handle pause/resume flow.
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

			// Rate limiting (per-IP)
			select {
			case <-rateCh:
			case <-ctx.Done():
				return
			}

			// Probe execution
			r, err := prb.Run(ctx, ip)
			atomic.AddUint64(processed, 1)

			if err != nil {
				logger.CoreError("probe failed for %s: %v", ip, err)
				continue
			}

			atomic.AddUint64(succeed, 1)
			hooks.callOnSuccess(*r)

			// Non-blocking unless context canceled
			select {
			case results <- *r:
			case <-ctx.Done():
				return
			}
		}
	}
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Rate Limiter
// ────────────────────────────────────────────────────────────────────────────────
//

// makeRateCh returns a ticker channel that enforces a per-second rate limit.
//
//	rate > 0  → new Ticker(1/rate)
//	rate == 0 → closed channel (no rate limiting)
func makeRateCh(rate int) <-chan time.Time {
	if rate > 0 {
		return time.NewTicker(time.Second / time.Duration(rate)).C
	}
	ch := make(chan time.Time)
	close(ch)
	return ch
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Progress Reporter
// ────────────────────────────────────────────────────────────────────────────────
//

// runProgressReporter emits periodic progress updates until:
//   - progressDone channel is closed,
//   - or context cancels.
//
// It also tracks paused durations using PauseController.
// The reporting interval is configured in General.StatusInterval.
func runProgressReporter(
	ctx context.Context,
	progressDone <-chan struct{},
	pause *PauseController,
	start time.Time,
	total uint64,
	processed *uint64,
	succeed *uint64,
	onProgress func(Progress),
) {
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
			if pause != nil && pause.IsPaused() {
				continue
			}
			reportProgress(
				start,
				pause.PausedDuration(),
				total, processed, succeed,
				onProgress,
			)
		}
	}
}
