package result

import (
	"net"
	"strings"
	"time"
)

func ParseCSV(rec []string) (ResultIPScan, bool) {
	if len(rec) == 0 {
		return ResultIPScan{}, false
	}

	ip := net.ParseIP(strings.TrimSpace(rec[0]))
	if ip == nil {
		return ResultIPScan{}, false
	}

	lat := DefaultFallbackLatency
	if len(rec) >= 2 {
		if d, err := time.ParseDuration(strings.TrimSpace(rec[1])); err == nil {
			lat = d
		}
	}

	return ResultIPScan{IP: ip.String(), Latency: lat}, true
}

func (s ResultIPScan) EncodeCSV() []string {
	return []string{s.IP, s.Latency.String()}
}
