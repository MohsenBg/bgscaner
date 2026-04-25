// Package dns provides helpers and abstractions for performing DNS queries
// within scanning and network analysis tools.
//
// It defines supported DNS transports, common record types, and conversion
// utilities for interacting with github.com/miekg/dns wire‑format constants.
package dns

import (
	"strings"

	"github.com/miekg/dns"
)

// Transport represents the protocol used to perform DNS queries.
//
// Each transport provides different characteristics in terms of speed,
// compatibility, privacy, and firewall traversal.
type Transport string

// Supported DNS transport modes.
const (
	// UDP performs classic DNS queries over UDP (port 53).
	// Fast, lightweight, and universally supported.
	UDP Transport = "UDP"

	// TCP performs DNS queries over TCP (port 53).
	// Used for large responses or when UDP is truncated or unreliable.
	TCP Transport = "TCP"

	// DOT performs DNS‑over‑TLS (RFC 7858) on port 853.
	// Provides encryption between client and resolver.
	DOT Transport = "DOT"

	// DOH represents DNS‑over‑HTTPS (RFC 8484).
	// Currently not implemented by this scanner because DoH requires
	// domain‑based resolvers, while the scanner primarily targets resolvers by IP.
	DOH Transport = "DOH"
)

// RecordType represents a DNS record type used in queries.
//
// These record types are the core focus of scanning workflows, revealing
// infrastructure layout, service configuration, and domain metadata.
type RecordType string

// Common DNS record types.
const (
	// TypeA resolves a domain to an IPv4 address.
	TypeA RecordType = "A"

	// TypeAAAA resolves a domain to an IPv6 address.
	TypeAAAA RecordType = "AAAA"

	// TypeCNAME defines a canonical alias to another domain.
	TypeCNAME RecordType = "CNAME"

	// TypeNS identifies authoritative nameservers for a domain.
	TypeNS RecordType = "NS"

	// TypeMX specifies mail exchangers for a domain.
	TypeMX RecordType = "MX"

	// TypeTXT stores free‑form text data (SPF, DKIM, verification, etc.).
	TypeTXT RecordType = "TXT"
)

// ParseTransport converts a string into a Transport value.
// Input is trimmed and case‑insensitive.
//
// Behavior:
//   - Recognized values: UDP, TCP, DOT, DOH
//   - DOH is mapped to DOT due to lack of direct support
//   - Unknown or empty values fall back to UDP (fastest + widely supported)
func ParseTransport(s string) Transport {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UDP":
		return UDP
	case "TCP":
		return TCP
	case "DOT":
		return DOT
	case "DOH":
		return DOT
	default:
		return UDP
	}
}

// toMiekgDNS converts a RecordType into the corresponding
// github.com/miekg/dns constant. Unknown record types return dns.TypeNone.
func toMiekgDNS(record RecordType) uint16 {
	switch record {
	case TypeA:
		return dns.TypeA
	case TypeAAAA:
		return dns.TypeAAAA
	case TypeCNAME:
		return dns.TypeCNAME
	case TypeNS:
		return dns.TypeNS
	case TypeMX:
		return dns.TypeMX
	case TypeTXT:
		return dns.TypeTXT
	default:
		return dns.TypeNone
	}
}

// ParseDNSRcode converts a textual DNS RCODE name into the corresponding
// miekg/dns numeric constant. Input is case‑insensitive.
//
// Unknown codes map to dns.RcodeServerFailure.
func ParseDNSRcode(rCode string) int {
	switch strings.ToLower(strings.TrimSpace(rCode)) {

	case "noerror", "success":
		return dns.RcodeSuccess

	case "formerr", "formaterror":
		return dns.RcodeFormatError

	case "servfail", "serverfailure":
		return dns.RcodeServerFailure

	case "nxdomain", "nameerror":
		return dns.RcodeNameError

	case "notimp", "notimplemented":
		return dns.RcodeNotImplemented

	case "refused":
		return dns.RcodeRefused

	default:
		return dns.RcodeServerFailure
	}
}
