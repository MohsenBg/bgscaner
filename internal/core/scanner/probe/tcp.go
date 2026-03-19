package probe

import (
	"bgscan/internal/core/result"
	"context"
	"net"
	"time"
)

// TCPProbe checks if a TCP port is reachable on a target IP.
// A successful TCP connection indicates the host and port are reachable.
type TCPProbe struct {
	port    string
	timeout time.Duration
	dialer  net.Dialer
}

// NewTCPProbe creates a new TCP probe for the given port.
func NewTCPProbe(port string, timeout time.Duration) Probe {
	return &TCPProbe{
		port:    port,
		timeout: timeout,
		dialer: net.Dialer{
			Timeout: timeout,
		},
	}
}

// Run attempts to establish a TCP connection to the target IP.
func (p *TCPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(ip, p.port)

	start := time.Now()

	conn, err := p.dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	latency := time.Since(start)

	// Some services immediately close or never send data.
	// We attempt a short read to confirm the connection is alive.
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	var buf [1]byte
	_, err = conn.Read(buf[:])

	if err != nil {
		if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
			return nil, err
		}
	}

	return &result.IPScanResult{
		IP:      ip,
		Latency: latency,
	}, nil
}

func (p *TCPProbe) Close() error {
	return nil
}
