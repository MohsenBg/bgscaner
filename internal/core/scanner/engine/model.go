package engine

import (
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"strings"
)

// ScanHooks provides optional lifecycle callbacks for the scanning engine.
// All fields are optional — nil means the hook is disabled.
type ScanHooks struct {
	// OnProgress is called periodically with a scan progress snapshot.
	OnProgress func(Progress)

	// OnSuccess is called for each successfully scanned IP.
	OnSuccess func(result.IPScanResult)

	// OnScanEnd is called once after the entire scan finishes.
	OnScanEnd func()

	// OnError is called when a non-fatal engine error occurs.
	OnError func(error)
}

// callOnError calls OnError if set — avoids nil check at every call site.
func (h ScanHooks) callOnError(err error) {
	if h.OnError != nil {
		h.OnError(err)
	}
}

// callOnSuccess calls OnSuccess if set.
func (h ScanHooks) callOnSuccess(r result.IPScanResult) {
	if h.OnSuccess != nil {
		h.OnSuccess(r)
	}
}

// callOnScanEnd calls OnScanEnd if set.
func (h ScanHooks) callOnScanEnd() {
	if h.OnScanEnd != nil {
		h.OnScanEnd()
	}
}

// ── Chain ────────────────────────────────────────────────────────────────────

// ChainMode defines how multiple scan stages are executed.
type ChainMode string

const (
	// SimpleChain runs stages sequentially.
	// Each stage starts only after the previous one completes.
	SimpleChain ChainMode = "simple"

	// ParallelChain runs all stages concurrently with independent worker pools.
	ParallelChain ChainMode = "parallel"

	// PipelineChain streams results directly from one stage to the next.
	// Successful IPs from stage N feed immediately into stage N+1,
	// without waiting for the entire stage to finish.
	PipelineChain ChainMode = "pipeline"
)

// ChainConfig controls execution strategy for a multi-stage scan.
type ChainConfig struct {
	Mode ChainMode

	// MaxBuffer is the channel buffer size between pipeline stages.
	// Larger values reduce inter-stage blocking at the cost of memory.
	MaxBuffer int

	Stages []ScanConfig

	Pause *PauseController
}

// ScanConfig defines settings for a single scan stage.
type ScanConfig struct {
	Workers int // concurrent workers (0 = use CPU count)
	Rate    int // max requests per second (0 = unlimited)

	Probe  probe.Probe
	Writer *result.Writer
	Hooks  ScanHooks
}

// ParseChainMode converts a string into a valid ChainMode.
// If the input is empty or invalid, SimpleChain is returned as the default.
func ParseChainMode(s string) ChainMode {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case string(SimpleChain):
		return SimpleChain
	case string(ParallelChain):
		return ParallelChain
	case string(PipelineChain):
		return PipelineChain
	default:
		return SimpleChain
	}
}
