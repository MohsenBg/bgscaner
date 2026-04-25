package probe

import (
	"bgscan/internal/core/result"
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	// icmpProtocol is the IP protocol number for ICMPv4.
	icmpProtocol = 1

	// maxPacket defines the maximum size of an incoming ICMP packet buffer.
	maxPacket = 4096

	// readTimeout is the per-iteration read deadline used by the reader loop.
	// Short timeouts ensure the goroutine wakes up frequently to check for shutdown.
	readTimeout = 200 * time.Millisecond

	// payload defines the optional ICMP echo payload.
	// Kept empty to minimize packet size for scanning workloads.
	payload = ""
)

// ICMPProbe implements the Probe interface using ICMP echo requests (ping)
// to measure reachability and latency for IPv4 targets.
//
// The probe maintains a single shared ICMP socket and a dedicated reader
// goroutine that demultiplexes echo replies back to waiting Ping callers.
//
// Design highlights:
//
//   - Uses a single reader goroutine per ICMPProbe instance.
//   - Maps (id, sequence) pairs to waiting goroutines via sync.Map.
//   - Supports both raw "ip4:icmp" and "udp4" fallback modes.
//   - Provides a safe shutdown mechanism via done channel and Close().
type ICMPProbe struct {
	// conn is the underlying ICMP socket used for sending and receiving packets.
	conn *icmp.PacketConn

	// mode indicates the socket type used ("raw" or "udp").
	mode string

	// id is the ICMP identifier for echo requests, derived from the process ID.
	id int

	// seq is an atomically incremented sequence number used to construct
	// unique (id, seq) pairs for matching requests and replies.
	seq uint32

	// timeout is the per-Ping timeout used when waiting for a reply.
	timeout time.Duration

	// tries defines how many times Run will attempt a Ping before failing.
	tries uint16

	// waiters holds per-request channels keyed by (id, seq).
	// Each active Ping registers a channel here and waits for a reply signal.
	waiters sync.Map

	// lifecycle
	done      chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// ------------------------------------------------------------
// Constructor
// ------------------------------------------------------------

// NewICMPProbe creates a new ICMPProbe using a shared ICMP socket.
//
// The constructor attempts to open a raw ICMP socket first:
//
//   - "ip4:icmp" bound to 0.0.0.0
//
// If raw sockets are not permitted (e.g. due to OS permissions), it
// falls back to a UDP-based ICMP listener:
//
//   - "udp4" bound to 0.0.0.0
//
// The returned ICMPProbe is ready to be initialized via Init before use.
//
//   - timeout: per-Ping timeout used when waiting for an echo reply.
//   - timesTry: number of Ping attempts performed by Run before giving up.
func NewICMPProbe(timeout time.Duration, timesTry uint16) (Probe, error) {
	conn, mode, err := openICMPSocket()
	if err != nil {
		return nil, err
	}

	return &ICMPProbe{
		conn:    conn,
		mode:    mode,
		id:      os.Getpid() & 0xffff,
		timeout: timeout,
		tries:   timesTry,
		done:    make(chan struct{}),
	}, nil
}

// Init implements [Probe] and starts the background reader goroutine
// on first invocation.
//
// This method is idempotent: multiple calls are safe and will only
// launch a single reader goroutine thanks to startOnce.
//
// The provided context is currently not used directly but is included
// to satisfy the Probe lifecycle contract and allow future extensions.
func (p *ICMPProbe) Init(ctx context.Context) error {
	p.startOnce.Do(func() {
		go p.reader()
	})
	return nil
}

// openICMPSocket attempts to create an ICMP-capable socket suitable for
// sending and receiving echo requests and replies.
//
// It first tries a raw ICMP socket ("ip4:icmp") and, if that fails,
// falls back to "udp4" which is often available without special privileges.
//
// It returns the opened connection, the selected mode ("raw" or "udp"),
// and an error if both attempts fail.
func openICMPSocket() (*icmp.PacketConn, string, error) {
	// Try raw ICMP first.
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err == nil {
		return conn, "raw", nil
	}

	// Fall back to UDP "icmp".
	conn, err = icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return nil, "", err
	}

	return conn, "udp", nil
}

// ------------------------------------------------------------
// Reader (single goroutine)
// ------------------------------------------------------------

// reader is the single background goroutine responsible for consuming
// all incoming ICMP packets from the shared socket.
//
// For each received packet, reader parses the ICMP message and, when
// it encounters a valid Echo Reply, dispatches a wake-up signal to the
// corresponding waiter based on (id, seq).
//
// The loop periodically checks the done channel and terminates promptly
// when Close is called.
func (p *ICMPProbe) reader() {
	buf := make([]byte, maxPacket)

	for {
		select {
		case <-p.done:
			return
		default:
		}

		_ = p.conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, _, err := p.conn.ReadFrom(buf)
		if err != nil {
			if isTimeout(err) {
				continue
			}
			// Non-timeout errors (e.g. closed connection) terminate the reader.
			return
		}

		p.handlePacket(buf[:n])
	}
}

// handlePacket parses a single ICMP packet and, if it is an Echo Reply
// matching an active Ping request, notifies the corresponding waiter channel.
func (p *ICMPProbe) handlePacket(packet []byte) {
	msg, err := icmp.ParseMessage(icmpProtocol, packet)
	if err != nil || msg.Type != ipv4.ICMPTypeEchoReply {
		return
	}

	body, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return
	}

	key := makeKey(body.ID, body.Seq)

	if ch, ok := p.waiters.Load(key); ok {
		// Non-blocking send to avoid deadlocking if the waiter has already
		// given up (e.g. due to timeout or context cancellation).
		select {
		case ch.(chan struct{}) <- struct{}{}:
		default:
		}
	}
}

// makeKey composes a 64-bit key from ICMP identifier and sequence number,
// used to correlate requests and replies.
func makeKey(id, seq int) uint64 {
	return uint64(id)<<32 | uint64(seq)
}

// ------------------------------------------------------------
// Ping
// ------------------------------------------------------------

// Ping sends a single ICMP echo request to the given IP address and
// waits for a corresponding echo reply or timeout.
//
// Behavior:
//
//   - Validates and parses the IP string.
//
//   - Allocates a unique sequence number and registers a waiter channel.
//
//   - Sends an ICMP Echo message using the shared socket.
//
//   - Waits for one of the following events:
//
//   - Context cancellation (returns ctx.Err())
//
//   - Probe shutdown via Close (returns "icmp probe closed")
//
//   - Echo reply received (returns nil)
//
//   - Local timeout expiration (returns "timeout")
//
// The timeout parameter applies only to this Ping invocation and is
// independent of the Probe's default timeout field.
func (p *ICMPProbe) Ping(ctx context.Context, ip string, timeout time.Duration) error {
	dstIP := net.ParseIP(ip)
	if dstIP == nil {
		return errors.New("invalid ip")
	}

	seq := int(atomic.AddUint32(&p.seq, 1))
	key := makeKey(p.id, seq)

	ch := make(chan struct{}, 1)
	p.waiters.Store(key, ch)
	defer p.waiters.Delete(key)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq,
			Data: []byte(payload),
		},
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err = p.conn.WriteTo(data, p.destination(dstIP)); err != nil {
		return err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-p.done:
		return errors.New("icmp probe closed")

	case <-ch:
		return nil

	case <-timer.C:
		return errors.New("timeout")
	}
}

// destination returns the appropriate net.Addr for the current socket mode.
//
//   - In "udp" mode it returns a *net.UDPAddr.
//   - In "raw" mode it returns a *net.IPAddr.
func (p *ICMPProbe) destination(ip net.IP) net.Addr {
	if p.mode == "udp" {
		return &net.UDPAddr{IP: ip}
	}
	return &net.IPAddr{IP: ip}
}

// ------------------------------------------------------------
// Public Run
// ------------------------------------------------------------

// Run implements [Probe] and performs an ICMP-based reachability check
// for the given IP address.
//
// It executes up to p.tries Ping attempts, each using the Probe's
// configured timeout. On each attempt:
//
//   - If the context is canceled, Run returns immediately with ctx.Err().
//   - If Ping succeeds, Run returns a populated IPScanResult including latency.
//   - If Ping fails, Run records the error and continues (until tries exhausted).
//
// If all attempts fail, the last encountered error is returned.
func (p *ICMPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var lastErr error

	for i := 0; i < int(p.tries); i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start := time.Now()

		err := p.Ping(ctx, ip, p.timeout)
		if err != nil {
			lastErr = err
			continue
		}

		return &result.IPScanResult{
			IP:      ip,
			Latency: time.Since(start),
		}, nil
	}

	return nil, lastErr
}

// ------------------------------------------------------------
// Close (safe shutdown)
// ------------------------------------------------------------

// Close terminates the ICMPProbe and releases associated resources.
//
// It performs the following actions atomically and exactly once:
//
//   - Closes the done channel, signaling the reader goroutine to exit.
//   - Closes the underlying ICMP socket.
//
// Subsequent calls to Close are safe and will return the same error value
// as the first call (typically nil).
func (p *ICMPProbe) Close() error {
	var err error

	p.closeOnce.Do(func() {
		close(p.done)
		if p.conn != nil {
			err = p.conn.Close()
		}
	})

	return err
}

// ------------------------------------------------------------
// Utility
// ------------------------------------------------------------

// isTimeout reports whether the provided error represents a timeout
// condition for network operations.
func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
