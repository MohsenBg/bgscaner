package config

import (
	"math"
	"strings"
	"time"
)

// DNSConfig represents the top‑level DNS configuration, combining resolver,
// DNSTT, and SlipStream settings.
type DNSConfig struct {
	Resolver   *ResolverConfig   `toml:"resolver"`
	DNSTT      *DNSTTConfig      `toml:"dnstt"`
	SlipStream *SlipStreamConfig `toml:"slip_stream"`
}

// Normalize validates all DNS-related configurations and adjusts values
// to their allowed ranges. The provided ValidationReport is updated with
// any corrections made.
func (d *DNSConfig) Normalize(rep *ValidationReport) *ValidationReport {
	if d.Resolver != nil {
		d.Resolver.Normalize(rep)
	}
	if d.DNSTT != nil {
		d.DNSTT.Normalize(rep)
	}
	if d.SlipStream != nil {
		d.SlipStream.Normalize(rep)
	}
	return rep
}

///////////////////////////////////////////////////////////////////////////////
// Resolver
///////////////////////////////////////////////////////////////////////////////

// ResolverConfig defines settings for traditional DNS resolvers.
type ResolverConfig struct {
	Workers         int        `toml:"workers"`
	Protocol        string     `toml:"protocol"`
	Domain          string     `toml:"domain"`
	Port            uint16     `toml:"port"`
	CheckTypes      []string   `toml:"check_types"`
	EDNSBufSize     uint16     `toml:"ends_buffer_size"`
	Timeout         DurationMS `toml:"timeout"`
	Tries           int        `toml:"tries"`
	ShuffleIPs      bool       `toml:"shuffle_ips"`
	RandomSubdomain bool       `toml:"random_subdomain"`
	AcceptedRCodes  []string   `toml:"accepted_rcodes"`
	CheckDPI        bool       `toml:"check_dpi"`
	DPITimeout      DurationMS `toml:"dpi_timeout"`
	DPITries        int        `toml:"dpi_tries"`
	PrefixOutput    string     `toml:"prefix_output"`
}

// Normalize validates resolver settings and adjusts invalid values to defaults.
// All corrections are written into the ValidationReport.
func (r *ResolverConfig) Normalize(rep *ValidationReport) {
	def := DefaultDNSConfig().Resolver

	normalizeInt("DNS.Resolver.Workers", &r.Workers, 1, 2500, def.Workers, rep)

	proto := strings.ToLower(strings.TrimSpace(r.Protocol))
	switch proto {
	case "udp", "tcp", "dot", "doh":
		r.Protocol = proto
	default:
		old := r.Protocol
		r.Protocol = def.Protocol
		rep.AddChange("DNS.Resolver.Protocol", old, r.Protocol, "invalid → default")
	}

	normalizeString("DNS.Resolver.Domain", &r.Domain, def.Domain, rep)
	normalizeUint16("DNS.Resolver.Port", &r.Port, 1, math.MaxUint16, def.Port, rep)

	if len(r.CheckTypes) == 0 {
		old := r.CheckTypes
		r.CheckTypes = def.CheckTypes
		rep.AddChange("DNS.Resolver.CheckTypes", old, r.CheckTypes, "empty → default")
	}

	normalizeDuration("DNS.Resolver.Timeout", &r.Timeout,
		NewDurationMS(100*time.Millisecond), NewDurationMS(30*time.Second), def.Timeout, rep)

	normalizeInt("DNS.Resolver.Tries", &r.Tries, 1, 10, def.Tries, rep)
	normalizeInt("DNS.Resolver.DPITries", &r.DPITries, 1, 10, def.DPITries, rep)

	normalizeDuration("DNS.Resolver.DPITimeout", &r.DPITimeout,
		NewDurationMS(100*time.Millisecond), NewDurationMS(10*time.Second), def.DPITimeout, rep)

	normalizeString("DNS.Resolver.PrefixOutput", &r.PrefixOutput, def.PrefixOutput, rep)
}

///////////////////////////////////////////////////////////////////////////////
// DNSTT
///////////////////////////////////////////////////////////////////////////////

// DNSTTConfig defines configuration for DNSTT (DNS Tunnel Transport) scanning.
type DNSTTConfig struct {
	Enabled      bool       `toml:"enabled"`
	Workers      int        `toml:"workers"`
	Domain       string     `toml:"domain"`
	PublicKey    string     `toml:"public_key"`
	Timeout      DurationMS `toml:"timeout"`
	PrefixOutput string     `toml:"prefix_output"`
}

// Normalize validates DNSTT settings when the feature is enabled.
func (d *DNSTTConfig) Normalize(rep *ValidationReport) {
	if !d.Enabled {
		return
	}

	def := DefaultDNSConfig().DNSTT

	normalizeInt("DNS.DNSTT.Workers", &d.Workers, 1, 500, def.Workers, rep)
	normalizeString("DNS.DNSTT.Domain", &d.Domain, def.Domain, rep)
	normalizeString("DNS.DNSTT.PublicKey", &d.PublicKey, def.PublicKey, rep)

	normalizeDuration("DNS.DNSTT.Timeout", &d.Timeout,
		NewDurationMS(100*time.Millisecond), NewDurationMS(60*time.Second), def.Timeout, rep)

	normalizeString("DNS.DNSTT.PrefixOutput", &d.PrefixOutput, def.PrefixOutput, rep)
}

///////////////////////////////////////////////////////////////////////////////
// SlipStream
///////////////////////////////////////////////////////////////////////////////

// SlipStreamConfig defines configuration for SlipStream-based DNS scanning.
type SlipStreamConfig struct {
	Enabled      bool       `toml:"enabled"`
	Workers      int        `toml:"workers"`
	Domain       string     `toml:"domain"`
	CertPath     string     `toml:"cert_path"`
	Timeout      DurationMS `toml:"timeout"`
	PrefixOutput string     `toml:"prefix_output"`
}

// Normalize validates SlipStream settings when the feature is enabled.
func (s *SlipStreamConfig) Normalize(rep *ValidationReport) {
	if !s.Enabled {
		return
	}

	def := DefaultDNSConfig().SlipStream

	normalizeInt("DNS.SlipStream.Workers", &s.Workers, 1, 500, def.Workers, rep)
	normalizeString("DNS.SlipStream.Domain", &s.Domain, def.Domain, rep)

	// empty cert is valid too
	//normalizeString("DNS.SlipStream.CertPath", &s.CertPath, def.CertPath, rep)

	normalizeDuration("DNS.SlipStream.Timeout", &s.Timeout,
		NewDurationMS(100*time.Millisecond), NewDurationMS(60*time.Second), def.Timeout, rep)

	normalizeString("DNS.SlipStream.PrefixOutput", &s.PrefixOutput, def.PrefixOutput, rep)
}
