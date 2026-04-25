package probe

import (
	"bgscan/internal/core/result"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// TCPProbe performs a TCP reachability test against a target IP address.
//
// It establishes a TCP connection to the specified port and measures
// connection latency. A successful TCP handshake indicates that the
// remote host is reachable on that port.
//
// In addition to connectivity testing, TCPProbe performs two lightweight
// checks:
//
//   - A short read window to detect simple DPI-injected responses
//     (e.g., "blocked", "filter", "deny").
//   - A minimal write operation to verify connection stability.
//
// TCPProbe does not maintain background goroutines or shared state,
// and is safe to use concurrently if each invocation uses its own instance.
type TCPProbe struct {
	// port is the TCP port number (as string) used for dialing.
	port string

	// timeout defines the maximum duration for establishing the TCP connection.
	timeout time.Duration

	// dialer is the configured net.Dialer used for context-aware dialing.
	dialer net.Dialer

	// tries defines how many times Run will attempt a Ping before failing.
	tries uint16
}

// NewTCPProbe returns a configured TCPProbe for the given port and timeout.
//
// Parameters:
//   - port: TCP port number as a string (e.g., "80", "443").
//   - timeout: maximum duration allowed for connection establishment.
//
// The returned value implements the Probe interface.
func NewTCPProbe(port string, timeout time.Duration, tries uint16) Probe {
	return &TCPProbe{
		port:    port,
		tries:   tries,
		timeout: timeout,
		dialer: net.Dialer{
			Timeout: timeout,
		},
	}
}

// Init implements [Probe].
//
// TCPProbe does not allocate persistent resources or spawn background
// goroutines, so Init currently performs no initialization and returns nil.
//
// The method exists to satisfy the common Probe lifecycle.
func (p *TCPProbe) Init(ctx context.Context) error {
	return nil
}

// Run executes the TCP probe against the provided IP address.
//
// The procedure:
//
//  1. Validates the input context for early cancellation.
//  2. Attempts to establish a TCP connection using DialContext.
//  3. Measures connection latency from dial start to completion.
//  4. Performs a short read (120ms deadline) to detect potential
//     DPI-injected responses.
//  5. Performs a minimal write to confirm connection stability.
//
// If the TCP handshake succeeds and no suspicious response is detected,
// Run returns an IPScanResult containing the measured latency.
//
// If the context is canceled, the connection fails, or a DPI-like
// injected response is detected, an error is returned.
func (p *TCPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(ip, p.port)
	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		start := time.Now()

		conn, err := p.dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			lastErr = err
			if isTimeout(err) {
				continue
			}
			return nil, err
		}

		// Brief read window to detect injections or filtering behavior.
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Millisecond))

		buf := make([]byte, 64)
		n, readErr := conn.Read(buf)
		if readErr == nil && n > 0 {
			data := string(buf[:n])

			// Basic heuristic for common filtering/DPI responses.
			if strings.Contains(data, "blocked") ||
				strings.Contains(data, "filter") ||
				strings.Contains(data, "deny") {
				conn.Close()
				return nil, fmt.Errorf("dpi injected response")
			}
		}

		// Minimal write to verify connection stability.
		_ = conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err = conn.Write([]byte{0}); err != nil {
			conn.Close()
			lastErr = err
			if isTimeout(err) {
				continue
			}
			return nil, err
		}

		conn.Close()

		return &result.IPScanResult{
			IP:      ip,
			Latency: time.Since(start),
		}, nil
	}

	return nil, fmt.Errorf("tcp probe failed after %d tries: %w", p.tries, lastErr)
}

// Close implements the Probe interface.
//
// TCPProbe maintains no long-lived resources, open sockets, or background
// goroutines. Close currently performs no action and always returns nil.
func (p *TCPProbe) Close() error {
	return nil
}
