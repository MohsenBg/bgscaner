package engine

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/iplist"
	"bgscan/internal/logger"
	"context"
	"sync/atomic"
	"time"
)

type stageExecutor struct {
	ctx    context.Context
	stage  ScanConfig
	pause  *PauseController
	rateCh <-chan time.Time

	start     time.Time
	total     atomic.Uint64
	processed atomic.Uint64
	succeed   atomic.Uint64

	progressDone chan struct{}
}

func newStageExecutor(ctx context.Context, stage ScanConfig, pause *PauseController, total uint64) *stageExecutor {
	exec := &stageExecutor{
		ctx:    ctx,
		stage:  stage,
		pause:  pause,
		rateCh: makeRateCh(stage.Rate),
	}

	exec.total.Store(total)
	exec.stage.Writer.Start()
	exec.startProgressReporter()
	stage.Probe.Init(ctx)

	return exec
}

func runStageProgressReporter(
	ctx context.Context,
	done <-chan struct{},
	start time.Time,
	pause *PauseController,
	total *atomic.Uint64,
	processed *atomic.Uint64,
	succeed *atomic.Uint64,
	onProgress func(Progress),
) {
	ticker := time.NewTicker(config.GetGeneral().StatusInterval.Duration())
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if pause != nil && pause.IsPaused() {
				continue
			}

			t := total.Load()
			p := processed.Load()
			s := succeed.Load()
			reportProgress(start, pause.PausedDuration(), t, &p, &s, onProgress)
		}
	}
}

func (e *stageExecutor) startProgressReporter() {
	logger.DebugInfo("check nil")
	if e.stage.Hooks.OnProgress == nil {
		return
	}

	e.progressDone = make(chan struct{})
	e.start = time.Now()
	go runStageProgressReporter(
		e.ctx,
		e.progressDone,
		e.start,
		e.pause,
		&e.total,
		&e.processed,
		&e.succeed,
		e.stage.Hooks.OnProgress,
	)
}

func streamStageFromFile(
	ctx context.Context,
	input string,
	maxIP int,
	stage ScanConfig,
	output chan string,
	exec *stageExecutor,
	next *stageExecutor,
	pause *PauseController,
) {
	workers := getWorkerCount(stage.Workers)
	inputCh := make(chan string, workers*2)

	streamDone := make(chan error, 1)
	go func() {
		defer close(inputCh)
		streamDone <- iplist.StreamActiveIPs(ctx, input, maxIP, inputCh)
	}()

	runWorkerPool(ctx, workers, pause, inputCh, func(ip string) {
		if exec.processIP(ip) && output != nil {
			select {
			case output <- ip:
				if next != nil {
					next.total.Add(1)
				}
			case <-ctx.Done():
			}
		}
	})

	if err := <-streamDone; err != nil && err != context.Canceled {
		logger.CoreError("StreamActiveIPs: %v", err)
		stage.Hooks.callOnError(err)
	}
}

func streamStageFromChannel(
	ctx context.Context,
	input chan string,
	stage ScanConfig,
	output chan string,
	exec *stageExecutor,
	next *stageExecutor,
	pause *PauseController,
) {
	workers := getWorkerCount(stage.Workers)

	runWorkerPool(ctx, workers, pause, input, func(ip string) {
		if exec.processIP(ip) && output != nil {
			select {
			case output <- ip:
				if next != nil {
					next.total.Add(1) // ← atomic add
				}
			case <-ctx.Done():
			}
		}
	})
}

// cleanup finalizes stage execution and releases resources.
//
// It performs the following shutdown sequence:
//
//  1. Stops the result writer
//  2. Terminates the progress reporter
//  3. Closes the probe implementation
//  4. Triggers the OnScanEnd hook
//
// cleanup must always be called after a stageExecutor finishes execution.
func (e *stageExecutor) cleanup() {
	e.stage.Writer.Stop()
	e.stopProgressReporter()

	// send last report
	total := e.total.Load()
	processed := e.processed.Load()
	succeed := e.succeed.Load()
	if e.stage.Hooks.OnProgress != nil {
		reportProgress(
			e.start,
			e.pause.PausedDuration(),
			total,
			&processed,
			&succeed,
			e.stage.Hooks.OnProgress,
		)
	}

	if err := e.stage.Probe.Close(); err != nil {
		e.stage.Hooks.callOnError(err)
	}
	e.stage.Hooks.callOnScanEnd()
}

// stopProgressReporter signals the progress reporting goroutine to stop.
func (e *stageExecutor) stopProgressReporter() {
	if e.progressDone != nil {
		close(e.progressDone)
	}
}

// processIP executes the configured probe against a single IP address.
//
// Workflow:
//
//  1. Wait for the rate limiter (if enabled)
//  2. Execute the probe.Run() method
//  3. Update processed/success counters
//  4. Trigger success hooks
//  5. Write the result using the stage writer
//
// Returns true if the probe succeeded and produced a valid result.
func (e *stageExecutor) processIP(ip string) (success bool) {
	// Rate limiting
	select {
	case <-e.rateCh:
	case <-e.ctx.Done():
		return false
	}

	r, err := e.stage.Probe.Run(e.ctx, ip)
	e.processed.Add(1)

	if err != nil {
		logger.CoreError("%s", err.Error())
		return false
	}

	e.succeed.Add(1)
	e.stage.Hooks.callOnSuccess(*r)
	e.stage.Writer.Write(*r)

	return true
}
