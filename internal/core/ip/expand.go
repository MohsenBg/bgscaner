package ip

import (
	"context"
	"fmt"
	"net"
)

// StreamCIDR sequentially streams all IP addresses of the given CIDR
// as strings into the provided channel. It stops if the context is canceled.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	ch := make(chan string)
//	go func() {
//	    defer close(ch)
//	    StreamCIDR(ctx, "192.168.1.0/30", ch)
//	}()
//
//	for ip := range ch { fmt.Println(ip) }
//
// NOTE: For IPv6 prefixes broader than /64, the range may be enormous.
//
//	Callers should apply RangeLimits before streaming to avoid exhaustion.
func StreamCIDR(ctx context.Context, cidr string, out chan<- string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("parse CIDR %q: %w", cidr, err)
	}

	start := masked(ip, network.Mask)
	for curr := start; network.Contains(curr); curr = increment(curr) {
		select {
		case out <- curr.String():
			// keep streaming
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// increment returns the next IP address by incrementing the given IP.
// It supports both IPv4 and IPv6 transparently.
func increment(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

// masked applies a subnet mask to an IP and returns a new masked copy.
func masked(ip net.IP, mask net.IPMask) net.IP {
	m := ip.Mask(mask)
	out := make(net.IP, len(m))
	copy(out, m)
	return out
}
