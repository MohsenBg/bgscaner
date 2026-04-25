package probe

import (
	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"context"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
)

// DnsRequest defines how a DNS resolver should be tested.
//
// It controls:
//
//   - Target domain and optional random subdomain behavior
//   - DNS record types to query (A, AAAA, TXT, …)
//   - Transport mode (UDP/TCP/DoT/DoH, depending on dns.Transport)
//   - EDNS0 buffer size
//   - Resolver honesty / DPI detection attempts
//   - Retry behavior and timeouts for both DPI and normal probes
type DnsRequest struct {
	// Domain is the base domain to query during normal probing.
	// It may be prefixed with a random label when RandomSubdomain is enabled.
	Domain string

	// Port is the resolver port (typically 53 for UDP/TCP DNS,
	// or protocol-specific ports for encrypted transports).
	Port uint16

	// RandomSubdomain, when true, causes each probe to generate a unique
	// random subdomain under Domain in order to avoid resolver cache effects.
	RandomSubdomain bool

	// DpiCheck enables resolver honesty/DPI detection using a guaranteed
	// NXDOMAIN domain prior to normal record queries.
	DpiCheck bool

	// DpiTimeout is the per-request timeout for DPI verification queries.
	DpiTimeout time.Duration

	// DpiTries is the maximum number of DPI verification attempts before
	// giving up and treating the resolver as unresponsive or unreliable.
	DpiTries int

	// Edns0Size controls the EDNS0 UDP buffer size advertised in queries.
	Edns0Size uint16

	// CheckTypes is the ordered list of DNS record types to test (by name),
	// e.g. []string{"A", "AAAA"}. The first acceptable response terminates
	// the probe successfully.
	CheckTypes []string

	// AcceptedRcodes defines which DNS response codes are considered
	// successful for normal probing. Responses with other rcodes are treated
	// as failures for that record type.
	AcceptedRcodes []uint16

	// Timeout is the per-query timeout for normal (non-DPI) resolver tests.
	Timeout time.Duration

	// Transport is the underlying transport mechanism used to contact the
	// resolver (UDP, TCP, DoT, DoH, etc.), as defined by dns.Transport.
	Transport dns.Transport

	// Tries is the maximum number of retries per record type during normal
	// probing before considering that type failed.
	Tries int
}

// ResolverProbe performs recursive DNS validation against a single resolver
// IP address.
//
// It optionally runs a DPI/hijacking honesty check using a guaranteed-invalid
// domain (under the .invalid TLD) and then executes normal DNS record tests
// according to the DnsRequest configuration.
type ResolverProbe struct {
	request *DnsRequest
}

// NewResolverProbe constructs a ResolverProbe and normalizes the retry
// counters to safe minimum values.
//
// Specifically:
//
//   - If Tries <= 0, it is forced to 1.
//   - If DpiTries <= 0, it is forced to 1.
//
// The returned value implements the Probe interface.
func NewResolverProbe(req *DnsRequest) Probe {
	if req.Tries <= 0 {
		req.Tries = 1
	}
	if req.DpiTries <= 0 {
		req.DpiTries = 1
	}
	return &ResolverProbe{request: req}
}

// Init implements [Probe] and currently performs no initialization.
//
// ResolverProbe does not maintain background goroutines or shared state;
// the method exists to satisfy the common Probe lifecycle.
func (r *ResolverProbe) Init(ctx context.Context) error {
	return nil
}

// Close implements [Probe] and releases any associated resources.
//
// ResolverProbe has no persistent resources, so Close is a no-op and
// always returns nil.
func (r *ResolverProbe) Close() error {
	return nil
}

// Run executes the full resolver validation sequence against the given IP.
//
// The flow is:
//
//  1. Validate the input context (early exit on cancellation).
//  2. If DpiCheck is enabled, run verifyResolverHonesty to detect
//     hijacking/DPI based on a guaranteed-invalid domain.
//  3. Execute normal probing (executeNormalProbe) using the configured
//     record types and AcceptedRcodes.
//
// On success, an IPScanResult with the measured latency is returned.
// On failure, an error describing the failure mode is returned.
func (r *ResolverProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if r.request.DpiCheck {
		if err := r.verifyResolverHonesty(ctx, ip); err != nil {
			return nil, err
		}
	}

	return r.executeNormalProbe(ctx, ip)
}

// verifyResolverHonesty checks whether a resolver improperly returns a
// success rcode (0) for a guaranteed-invalid domain, which indicates
// potential DPI, hijacking, or other dishonest behavior.
//
// The function:
//
//   - Generates a random label under the ".invalid" TLD.
//   - Sends a series of A queries to the resolver.
//   - Treats any rcode == 0 as DPI/hijacking evidence (error).
//   - Treats any non-zero rcode as honest behavior.
//   - Retries up to DpiTries times before failing with an aggregated error.
func (r *ResolverProbe) verifyResolverHonesty(ctx context.Context, ip string) error {
	fakeDomain := generateRandomString(16) + ".invalid"

	timeout := r.request.DpiTimeout
	if timeout == 0 {
		timeout = 500 * time.Millisecond
	}

	query := dns.DNSQuery{
		Resolver:         ip,
		Port:             r.request.Port,
		Domain:           fakeDomain,
		RecordType:       dns.TypeA,
		Transport:        r.request.Transport,
		EDNSBufSize:      r.request.Edns0Size,
		RecursionDesired: true,
		Timeout:          timeout,
	}

	var lastErr error

	for i := 0; i < r.request.DpiTries; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		resp, err := query.Run()
		if err != nil {
			lastErr = err
			continue
		}

		// rcode 0 = resolver claims success → likely hijacking/DPI.
		if resp.Rcode == 0 {
			return fmt.Errorf("dpi detected: resolver returned rcode 0 for %s", fakeDomain)
		}

		// Any non-zero rcode is considered honest.
		return nil
	}

	return fmt.Errorf("dpi verification failed after %d tries: %w", r.request.DpiTries, lastErr)
}

// executeNormalProbe runs DNS queries against the configured record types
// and returns the first acceptable result as defined by AcceptedRcodes.
//
// For each record type in CheckTypes:
//
//   - It issues up to Tries queries (with context-aware cancellation).
//   - Records the latency of the first successful response.
//   - If the response rcode is in AcceptedRcodes, it returns success.
//   - If the response rcode is not accepted, it stops retrying this type
//     and proceeds to the next type.
//
// If no record type produces an accepted response, an error is returned.
func (r *ResolverProbe) executeNormalProbe(ctx context.Context, ip string) (*result.IPScanResult, error) {
	query := dns.DNSQuery{
		Resolver:         ip,
		Port:             r.request.Port,
		Transport:        r.request.Transport,
		EDNSBufSize:      r.request.Edns0Size,
		RecursionDesired: true,
		Timeout:          r.request.Timeout,
	}

	target := r.request.Domain
	if r.request.RandomSubdomain {
		target = generateRandomString(10) + "." + target
	}
	query.Domain = target

	resultObj := &result.IPScanResult{IP: ip}

	for _, typeStr := range r.request.CheckTypes {
		query.RecordType = parseRecordType(typeStr)

		var lastErr error

		for i := 0; i < r.request.Tries; i++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			start := time.Now()
			resp, err := query.Run()
			if err != nil {
				lastErr = err
				continue
			}

			resultObj.Latency = time.Since(start)

			if r.isRcodeAccepted(uint16(resp.Rcode)) {
				return resultObj, nil
			}

			// Got a response but with unacceptable rcode; stop retrying this type.
			break
		}

		_ = lastErr // reserved for future logging or metrics
	}

	return nil, fmt.Errorf("no accepted response for %s", target)
}

// isRcodeAccepted reports whether the given DNS response code is considered
// successful based on the probe configuration.
func (r *ResolverProbe) isRcodeAccepted(code uint16) bool {
	return slices.Contains(r.request.AcceptedRcodes, code)
}

// parseRecordType converts a string representation of a DNS record type
// into the corresponding dns.RecordType. Unknown values default to A.
func parseRecordType(s string) dns.RecordType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "TXT":
		return dns.TypeTXT
	case "NS":
		return dns.TypeNS
	case "CNAME":
		return dns.TypeCNAME
	case "MX":
		return dns.TypeMX
	default:
		return dns.TypeA
	}
}

// generateRandomString returns a random alphanumeric string of length n.
// It is used for generating unique subdomains and fake DPI check domains.
//
// Note: this uses math/rand and is intended for non-cryptographic purposes.
func generateRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
