package engine

import (
	"context"
	"sync"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Worker Pool Runtime
// ────────────────────────────────────────────────────────────────────────────────
//

// runWorkerPool launches a fixed‑size pool of goroutines that consume items
// from the input channel and process each one using the provided callback.
//
// Workflow:
//
//  1. Spawns `workers` concurrent goroutines.
//  2. Each goroutine invokes runWorker() until the input channel is closed or
//     the context is cancelled.
//  3. Waits for all workers to finish before returning.
//
// Pausing behavior:
//
//	When a non‑nil PauseController is provided, each worker cooperates
//	with pause/resume signals via pause.Wait(ctx).
//
// This function blocks until all worker goroutines complete.
func runWorkerPool(
	ctx context.Context,
	workers int,
	pause *PauseController,
	input <-chan string,
	process func(string),
) {
	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			runWorker(ctx, pause, input, process)
		})
	}

	wg.Wait()
}

// runWorker is the internal worker loop that processes incoming items from the
// input channel until one of the following occurs:
//
//   - The channel is closed (end of stream)
//   - The context is cancelled
//   - The pause controller terminates the worker
//
// The process callback is invoked once for every received item.
// This function exits gracefully when any termination condition occurs.
func runWorker(
	ctx context.Context,
	pause *PauseController,
	input <-chan string,
	process func(string),
) {
	for {
		// Respect pause/resume control before reading from channel.
		if pause != nil && !pause.Wait(ctx) {
			return
		}

		select {
		case <-ctx.Done():
			// Context canceled → shutdown immediately.
			return

		case item, ok := <-input:
			if !ok {
				// Channel closed → normal termination.
				return
			}
			process(item)
		}
	}
}

// getWorkerCount returns the provided worker count, defaulting to 1
// if the value is zero or negative.
//
// It ensures that the worker pool always has at least one worker.
func getWorkerCount(workers int) int {
	if workers <= 0 {
		return 1
	}
	return workers
}
