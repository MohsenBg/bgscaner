package probe

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/netutil"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

//
// HTTP PROBE
//

// HTTPProbe implements the Probe interface for performing HTTP/HTTPS requests.
//
// The probe is designed for network scanning scenarios where:
//
//   - The TCP connection must be made directly to a specific IP address
//   - The original hostname must be preserved for HTTP Host headers and TLS SNI
//
// This separation allows accurate virtual‑host and TLS scanning while bypassing DNS.
type HTTPProbe struct {
	req HTTPRequest
}

// HTTPRequest represents a fully resolved, normalized HTTP request configuration.
//
// It is derived from user configuration and contains all information required
// to perform a probe against a single IP address.
type HTTPRequest struct {
	// URL is the full request URL including scheme, host, and port.
	// Example: "https://example.com:443"
	URL string

	// Host is the normalized hostname extracted from configuration.
	// It is primarily used for SNI and hostname validation.
	Host string

	// SNI is the Server Name Indication value used during the TLS handshake.
	// If empty, the TLS stack may fall back to Host.
	SNI string

	// UseTLS indicates whether HTTPS (true) or plain HTTP (false) is used.
	UseTLS bool

	// SkipTLSVerify disables TLS certificate verification.
	// This is useful for scanning but should not be used in production clients.
	SkipTLSVerify bool

	// Timeout defines the total time allowed for the request,
	// including dialing, TLS handshake, and response headers.
	Timeout time.Duration

	// MinTLSVersion and MaxTLSVersion constrain the allowed TLS protocol versions.
	MinTLSVersion uint16
	MaxTLSVersion uint16
}

// NewHTTPProbe creates a new HTTPProbe instance using a resolved HTTPRequest.
func NewHTTPProbe(req HTTPRequest) Probe {
	return &HTTPProbe{req: req}
}

// Run executes the HTTP probe against a specific IP address.
//
// The function:
//
//   - Respects context cancellation
//   - Establishes a TCP connection directly to the IP
//   - Preserves hostname semantics for HTTP and TLS
//   - Measures request latency
func (p *HTTPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {

	// Abort early if the context has already been canceled.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client := p.buildHTTPClient(ip)

	// Use HEAD to minimize bandwidth while still validating service availability.
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &result.IPScanResult{
		IP:      ip,
		Latency: time.Since(start),
	}, nil
}

//
// HTTP CLIENT CONSTRUCTION
//

// buildHTTPClient creates a customized http.Client for this probe.
//
// The client:
//
//   - Forces connections to the provided IP address
//   - Preserves original host and SNI information
//   - Disables connection reuse to simplify lifecycle management
//   - Applies strict timeouts suitable for scanning workloads
func (p *HTTPProbe) buildHTTPClient(ip string) *http.Client {

	dialer := &net.Dialer{
		Timeout: p.req.Timeout,
	}

	transport := &http.Transport{
		// DialContext overrides the destination address to force the IP,
		// while still letting http.Client believe it is connecting to the host.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			target := net.JoinHostPort(ip, port)
			return dialer.DialContext(ctx, network, target)
		},

		// DisableKeepAlives ensures each request uses a fresh connection.
		// This avoids cross‑request state leakage during scanning.
		DisableKeepAlives: true,

		// Timeouts for header reading and TLS handshakes.
		ResponseHeaderTimeout: p.req.Timeout,
		TLSHandshakeTimeout:   p.req.Timeout,
	}

	// Apply TLS configuration when HTTPS is enabled.
	if p.req.UseTLS {
		transport.TLSClientConfig = &tls.Config{
			ServerName:         p.req.SNI,
			InsecureSkipVerify: p.req.SkipTLSVerify,
			MinVersion:         p.req.MinTLSVersion,
			MaxVersion:         p.req.MaxTLSVersion,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   p.req.Timeout,
	}
}

//
// CONFIG NORMALIZATION
//

// NewHTTPRequestFromConfig converts a raw HTTPConfig into a fully resolved HTTPRequest.
//
// This function performs:
//
//   - Protocol normalization (HTTP vs HTTPS)
//   - Hostname validation and normalization (IDN support)
//   - Default port resolution
//   - TLS version parsing and validation
//   - SNI determination
func NewHTTPRequestFromConfig(cfg config.HTTPConfig) (*HTTPRequest, error) {

	scheme := netutil.ProtocolToScheme(cfg.Protocol)
	isHTTPS := netutil.IsHTTPS(scheme)

	host, err := netutil.ExtractTLSServerName(cfg.Host)
	if err != nil {
		return nil, err
	}

	urlHost, err := netutil.NormalizeHostWithSuffix(cfg.Host)
	if err != nil {
		return nil, err
	}

	port := resolvePort(cfg.Port, isHTTPS)
	url := fmt.Sprintf("%s%s:%d", scheme, urlHost, port)

	sni, err := resolveSNI(cfg.ServerName, isHTTPS)
	if err != nil {
		return nil, err
	}

	minTLS, maxTLS, err := resolveTLSVersions(cfg)
	if err != nil {
		return nil, err
	}

	return &HTTPRequest{
		URL:           url,
		Host:          host,
		SNI:           sni,
		UseTLS:        isHTTPS,
		SkipTLSVerify: !cfg.TLSValidation,
		Timeout:       cfg.Timeout.Duration(),
		MinTLSVersion: minTLS,
		MaxTLSVersion: maxTLS,
	}, nil
}

//
// HELPER FUNCTIONS
//

// resolvePort returns the explicitly configured port,
// or the protocol default (443 for HTTPS, 80 for HTTP).
func resolvePort(port int, isHTTPS bool) uint16 {

	if port > 0 {
		return uint16(port)
	}

	if isHTTPS {
		return 443
	}

	return 80
}

// resolveSNI determines the Server Name Indication value for TLS.
//
// If TLS is disabled or no server name is provided,
// an empty string is returned.
func resolveSNI(serverName string, isHTTPS bool) (string, error) {

	if !isHTTPS || serverName == "" {
		return "", nil
	}

	return netutil.ExtractTLSServerName(serverName)
}

// resolveTLSVersions parses and validates TLS version constraints.
//
// It ensures that the minimum TLS version is not greater
// than the maximum allowed version.
func resolveTLSVersions(cfg config.HTTPConfig) (uint16, uint16, error) {

	minTLS, err := netutil.ParseTLSVersion(cfg.MinTLSVersion)
	if err != nil {
		return 0, 0, err
	}

	maxTLS, err := netutil.ParseTLSVersion(cfg.MaxTLSVersion)
	if err != nil {
		return 0, 0, err
	}

	if minTLS > maxTLS {
		return 0, 0, fmt.Errorf("min TLS version cannot be greater than max TLS version")
	}

	return minTLS, maxTLS, nil
}

func (p *HTTPProbe) Close() error {
	return nil
}
