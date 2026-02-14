package result

import "time"

type ResultIPScan struct {
	IP      string
	Latency time.Duration
}

const DefaultFallbackLatency = 999 * time.Second
