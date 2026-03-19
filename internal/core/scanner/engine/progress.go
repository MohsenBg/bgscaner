package engine

import (
	"sync/atomic"
	"time"
)

// Progress represents the current execution status of the engine.
type Progress struct {
	Total        int64         // total number of tasks
	Processed    int64         // number of processed tasks
	Succeed      int64         // number of successful tasks
	Percent      float64       // completion percentage (0-100)
	Elapsed      time.Duration // active elapsed time (excluding pauses)
	RatePerSec   float64       // processing rate (items/sec)
	ETA          time.Duration // estimated time remaining
	EstimatedEnd time.Time     // estimated completion timestamp
}

// reportProgress calculates current progress statistics and
// invokes the provided callback with a Progress snapshot.
//
// Thread safety:
// processed and succeed counters must be accessed atomically.
func reportProgress(
	start time.Time,
	paused time.Duration,
	total int64,
	processed *int64,
	succeed *int64,
	cb func(p Progress),
) {
	now := time.Now()

	done := atomic.LoadInt64(processed)
	success := atomic.LoadInt64(succeed)

	elapsed := now.Sub(start) - paused
	if elapsed < 0 {
		elapsed = 0
	}

	// Calculate processing rate
	var rate float64
	if elapsed > 0 {
		rate = float64(done) / elapsed.Seconds()
	}

	// Calculate completion percentage
	var percent float64
	if total > 0 {
		percent = float64(done) / float64(total) * 100
	}

	// Estimate remaining time
	var eta time.Duration
	var estimatedEnd time.Time

	if rate > 0 && done < total {
		remaining := float64(total - done)
		etaSeconds := remaining / rate

		eta = time.Duration(etaSeconds * float64(time.Second))
		estimatedEnd = now.Add(eta)
	}

	cb(Progress{
		Total:        total,
		Processed:    done,
		Succeed:      success,
		Percent:      percent,
		Elapsed:      elapsed,
		RatePerSec:   rate,
		ETA:          eta,
		EstimatedEnd: estimatedEnd,
	})
}
