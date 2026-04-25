package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type pauseEvent uint8

const (
	evPause pauseEvent = iota
	evResume
	evStop
)

type PauseController struct {
	paused atomic.Bool

	mu     sync.RWMutex
	resume chan struct{}

	pausedAt   time.Time
	totalPause atomic.Int64 // nanoseconds

	events chan pauseEvent
	done   chan struct{}
	once   sync.Once
}

func NewPauseController() *PauseController {
	return &PauseController{
		resume: make(chan struct{}),
		events: make(chan pauseEvent, 4),
		done:   make(chan struct{}),
	}
}

// -------------------------------
// Start loop
// -------------------------------

func (p *PauseController) Start() {
	go p.loop()
}

func (p *PauseController) loop() {
	for {
		select {
		case <-p.done:
			return

		case ev := <-p.events:
			switch ev {

			case evPause:
				p.handlePause()

			case evResume:
				p.handleResume()

			case evStop:
				p.handleResume()
				close(p.done)
				return
			}
		}
	}
}

// -------------------------------
// Event handlers
// -------------------------------

func (p *PauseController) handlePause() {
	if !p.paused.CompareAndSwap(false, true) {
		return
	}

	p.mu.Lock()
	p.pausedAt = time.Now()
	p.resume = make(chan struct{})
	p.mu.Unlock()
}

func (p *PauseController) handleResume() {
	if !p.paused.CompareAndSwap(true, false) {
		return
	}

	p.mu.Lock()
	if !p.pausedAt.IsZero() {
		elapsed := time.Since(p.pausedAt).Nanoseconds()
		p.totalPause.Add(elapsed)
		p.pausedAt = time.Time{}
	}
	close(p.resume)
	p.mu.Unlock()
}

// -------------------------------
// Public API
// -------------------------------

func (p *PauseController) Pause() {
	select {
	case p.events <- evPause:
	case <-p.done:
	}
}

func (p *PauseController) Resume() {
	select {
	case p.events <- evResume:
	case <-p.done:
	}
}

func (p *PauseController) Stop() {
	p.once.Do(func() {
		p.events <- evStop
	})
}

func (p *PauseController) IsPaused() bool {
	return p.paused.Load()
}

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

func (p *PauseController) PausedDuration() time.Duration {
	total := p.totalPause.Load()

	if p.paused.Load() {
		p.mu.RLock()
		pausedAt := p.pausedAt
		p.mu.RUnlock()

		if !pausedAt.IsZero() {
			total += time.Since(pausedAt).Nanoseconds()
		}
	}

	return time.Duration(total)
}
