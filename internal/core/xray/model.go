package xray

// XrayConfig represents the root configuration structure used to
// generate an Xray-core configuration file.
//
// Only the fields required by bgscan are modeled here. The structure
// is intentionally minimal and focuses on configuring local proxy
// inbounds and user-provided outbounds used during scanning.
type XrayConfig struct {
	// Inbounds defines the proxy entry points that accept connections.
	// In bgscan this typically contains a local SOCKS inbound used by
	// probes to route traffic through the Xray instance.
	Inbounds []Inbound `json:"inbounds"`

	// Outbounds defines how traffic should exit Xray.
	// The type is kept flexible because different proxy protocols
	// (VMess, VLESS, Trojan, Shadowsocks, etc.) have different
	// configuration schemas.
	Outbounds []any `json:"outbounds"`
}

// Inbound describes a single inbound proxy listener.
//
// In bgscan this is typically a SOCKS proxy bound to localhost that
// allows scanner probes to send traffic through the running Xray
// instance.
type Inbound struct {
	// Port is the TCP port the inbound listens on.
	Port int `json:"port"`

	// Listen defines the IP address the inbound binds to.
	// For security reasons bgscan restricts this to localhost.
	Listen string `json:"listen"`

	// Tag uniquely identifies the inbound and may be referenced
	// by routing rules inside the Xray configuration.
	Tag string `json:"tag"`

	// Protocol defines the inbound protocol (e.g. "socks", "http").
	Protocol string `json:"protocol"`

	// Settings contains protocol-specific configuration.
	Settings SocksSettings `json:"settings"`

	// Sniffing enables protocol detection for routed traffic.
	Sniffing SniffingSetting `json:"sniffing"`
}

// SocksSettings defines the configuration for a SOCKS inbound.
type SocksSettings struct {
	// Auth defines the authentication method.
	// bgscan uses "noauth" since the proxy is only exposed locally.
	Auth string `json:"auth"`

	// UDP enables UDP support for the SOCKS proxy.
	UDP bool `json:"udp"`

	// IP defines the IP address used for outbound UDP associations.
	IP string `json:"ip"`
}

// SniffingSetting controls protocol sniffing behavior.
//
// When enabled, Xray attempts to detect the underlying protocol
// (such as HTTP or TLS) and adjust routing decisions accordingly.
type SniffingSetting struct {
	// Enabled toggles protocol sniffing.
	Enabled bool `json:"enabled"`

	// DestOverride specifies which protocols should override
	// the destination address after detection.
	DestOverride []string `json:"destOverride"`
}
