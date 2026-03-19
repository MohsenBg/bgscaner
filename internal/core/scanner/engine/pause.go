package engine

import (
	"context"
	"sync"
	"sync/atomic"
)

// PauseController allows pausing and resuming concurrent work.
// Goroutines can call Wait() to block while the controller is paused.
type PauseController struct {
	paused atomic.Bool

	mu     sync.RWMutex
	resume chan struct{}
}

// NewPauseController creates a new PauseController.
func NewPauseController() *PauseController {
	return &PauseController{
		resume: make(chan struct{}),
	}
}

// Pause transitions the controller into a paused state.
// If already paused, this operation is a no-op.
func (p *PauseController) Pause() {
	if !p.paused.CompareAndSwap(false, true) {
		return
	}

	p.mu.Lock()
	p.resume = make(chan struct{})
	p.mu.Unlock()
}

// Resume transitions the controller back to the running state.
// All goroutines blocked in Wait() will be released.
func (p *PauseController) Resume() {
	if !p.paused.CompareAndSwap(true, false) {
		return
	}

	p.mu.Lock()
	close(p.resume)
	p.mu.Unlock()
}

// IsPaused returns true if the controller is currently paused.
func (p *PauseController) IsPaused() bool {
	return p.paused.Load()
}

// Wait blocks while the controller is paused.
//
// Returns:
//
//	true  -> resumed normally
//	false -> context cancelled
func (p *PauseController) Wait(ctx context.Context) bool {
	if !p.paused.Load() {
		return true
	}

	p.mu.RLock()
	ch := p.resume
	p.mu.RUnlock()

	select {
	case <-ch:
		return true
	case <-ctx.Done():
		return false
	}
}
