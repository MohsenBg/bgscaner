package probe

import (
	"bgscan/internal/core/result"
	"context"
)

// Probe represents a scanning primitive capable of probing a single IP
// address and returning a scan result.
//
// Implementations may perform different types of checks such as:
//   - ICMP ping
//   - TCP port probe
//   - HTTP request
//   - XRAY request
//
// The Run method must respect the provided context for cancellation and
// deadlines. If the context is canceled, the probe should terminate early
// and return ctx.Err().
type Probe interface {
	Run(ctx context.Context, ip string) (*result.IPScanResult, error)
	Close() error
}
