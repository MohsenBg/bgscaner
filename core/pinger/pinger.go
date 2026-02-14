package pinger

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// TcpPing attempts to establish a TCP connection to the given IP and port.
// Returns whether the host is reachable and the latency.
// If verbose is true, it prints connection errors.
func TcpPing(ip, port string, timeout time.Duration, verbose bool) (bool, time.Duration) {
	address := net.JoinHostPort(ip, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	latency := time.Since(start)

	if conn != nil {
		conn.Close()
	}

	if err != nil && verbose {
		fmt.Printf("// TCP ping failed %s:%s -> %v\n", ip, port, err)
	}

	return err == nil, latency
}

// IcmpPing sends an ICMP Echo request to the given IP.
// Returns whether the host is reachable and the latency.
func IcmpPing(ip string, timeout time.Duration, verbose bool) (bool, time.Duration) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		if verbose {
			fmt.Println("// ICMP listen error:", err)
		}
		return false, 0
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("ping"),
		},
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		if verbose {
			fmt.Println("// ICMP marshal error:", err)
		}
		return false, 0
	}

	dst := &net.IPAddr{IP: net.ParseIP(ip)}

	start := time.Now()
	if _, err := conn.WriteTo(data, dst); err != nil {
		if verbose {
			fmt.Println("// ICMP write error:", err)
		}
		return false, 0
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 1500)
	n, _, err := conn.ReadFrom(buf)
	latency := time.Since(start)

	if err != nil {
		if verbose {
			fmt.Println("// ICMP read error:", err)
		}
		return false, latency
	}

	rm, err := icmp.ParseMessage(1, buf[:n])
	if err != nil {
		if verbose {
			fmt.Println("// ICMP parse error:", err)
		}
		return false, latency
	}

	if rm.Type == ipv4.ICMPTypeEchoReply {
		return true, latency
	}

	if verbose {
		fmt.Println("// ICMP unexpected reply:", rm.Type)
	}

	return false, latency
}

// HttpPing sends an HTTP GET request to the given IP with a custom Host header.
// Returns true if the response status code is 2xx or 3xx.
func HttpPing(ip, host string, timeout time.Duration, verbose bool) bool {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         host, // SNI
			},
		},
	}

	url := fmt.Sprintf("https://%s/", ip)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if verbose {
			fmt.Println("// HTTP request creation failed:", err)
		}
		return false
	}

	// ✅ HTTP layer
	req.Host = host
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
	)

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("// HTTP request failed:", err)
		}
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// PingMode defines the type of ping to perform.
type PingMode int

const (
	PingICMP PingMode = iota
	PingTCP
	PingHTTP
)

// PingHost performs a ping check based on the selected mode.
func PingHost(ip, host string, port int, timeout time.Duration, mode PingMode, verbose bool) (bool, time.Duration) {
	switch mode {
	case PingICMP:
		return IcmpPing(ip, timeout, verbose)
	case PingTCP:
		return TcpPing(ip, fmt.Sprintf("%d", port), timeout, verbose)
	case PingHTTP:
		start := time.Now()
		ok := HttpPing(ip, host, timeout, verbose)
		return ok, time.Since(start)
	default:
		return false, 0
	}
}

func CanUseICMP() bool {
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
