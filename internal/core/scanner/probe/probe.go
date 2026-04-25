package probe

import (
	"bgscan/internal/core/result"
	"context"
)

// Probe defines the interface for all active scanning primitives.
type Probe interface {

	// Init performs probe initialization.
	//
	// This method is called once during scanner startup before any
	// Run calls are executed. It allows probes to allocate resources
	// such as sockets, background goroutines, caches, or protocol
	// state.
	//
	// Implementations that require no initialization may return nil.
	Init(ctx context.Context) error

	// Run executes a probe against the provided IP address.
	//
	// The method should:
	//   • honor ctx for cancellation
	//   • return a populated IPScanResult on success
	//   • return an error if the probe fails or times out
	Run(ctx context.Context, ip string) (*result.IPScanResult, error)

	// Close releases probe‑specific resources such as sockets,
	// goroutines, or file descriptors. It is called once at the
	// end of the scanner lifecycle.
	Close() error
}
