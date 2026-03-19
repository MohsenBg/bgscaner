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
	icmpProtocol = 1
	maxPacket    = 1500
	readTimeout  = 2 * time.Second
	payload      = "bgscan"
)

// ICMPProbe implements an ICMP echo probe using a single shared socket.
//
// The probe maintains a background reader goroutine that receives ICMP
// replies and dispatches them to waiting requests using a waiter map.
// Each outgoing echo request is identified by (ID, Seq).
type ICMPProbe struct {
	conn *icmp.PacketConn
	mode string

	id      int
	seq     uint32
	timeout time.Duration

	waiters sync.Map // map[key]chan struct{}
}

// NewICMPProbe initializes an ICMP probe.
//
// The function attempts to open a raw ICMP socket first. If raw sockets
// are not permitted (e.g., non‑root environments), it falls back to UDP
// mode which works on most systems.
func NewICMPProbe(timeout time.Duration) (Probe, error) {

	conn, mode, err := openICMPSocket()
	if err != nil {
		return nil, err
	}

	p := &ICMPProbe{
		conn:    conn,
		mode:    mode,
		id:      os.Getpid() & 0xffff,
		timeout: timeout,
	}

	go p.reader()

	return p, nil
}

// openICMPSocket attempts to open a raw ICMP socket and falls back to UDP.
func openICMPSocket() (*icmp.PacketConn, string, error) {

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err == nil {
		return conn, "raw", nil
	}

	conn, err = icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return nil, "", err
	}

	return conn, "udp", nil
}

// makeKey generates a unique key from ICMP id and sequence.
func makeKey(id, seq int) uint64 {
	return uint64(id)<<32 | uint64(seq)
}

// reader continuously reads ICMP packets and dispatches replies
// to waiting goroutines.
func (p *ICMPProbe) reader() {

	buf := make([]byte, maxPacket)

	for {
		_ = p.conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, _, err := p.conn.ReadFrom(buf)
		if err != nil {
			if isTimeout(err) {
				continue
			}
			return
		}

		p.handlePacket(buf[:n])
	}
}

// handlePacket processes a single ICMP packet.
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
		select {
		case ch.(chan struct{}) <- struct{}{}:
		default:
		}
	}
}

// Ping sends an ICMP echo request and waits for a reply.
func (p *ICMPProbe) Ping(ip string, timeout time.Duration) error {

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

	if _, err := p.conn.WriteTo(data, p.destination(dstIP)); err != nil {
		return err
	}

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// destination returns the correct destination address
// depending on probe mode.
func (p *ICMPProbe) destination(ip net.IP) net.Addr {

	if p.mode == "udp" {
		return &net.UDPAddr{IP: ip}
	}

	return &net.IPAddr{IP: ip}
}

// Run executes the probe and returns a scan result.
func (p *ICMPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start := time.Now()

	if err := p.Ping(ip, p.timeout); err != nil {
		return nil, err
	}

	return &result.IPScanResult{
		IP:      ip,
		Latency: time.Since(start),
	}, nil
}

// isTimeout checks if the error is a network timeout.
func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}

func (p *ICMPProbe) Close() error {
	if p.conn == nil {
		return nil
	}
	// Close the socket. This will cause reader()'s ReadFrom to return
	err := p.conn.Close()
	return err
}
