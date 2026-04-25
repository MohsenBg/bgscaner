package xray

// getInbound constructs the default inbound SOCKS configuration used
// by bgscan when launching a temporary Xray instance.
//
// The inbound listens on localhost (127.0.0.1) and exposes a SOCKS
// proxy that bgscan probes can route traffic through. Authentication
// is disabled since the proxy is only intended for local use.
//
// Sniffing is enabled with HTTP and TLS destination overrides so that
// Xray can correctly detect protocols even when the target service
// does not explicitly specify them.
func getInbound(port uint16) Inbound {
	return Inbound{
		Port:     port,
		Listen:   "127.0.0.1",
		Tag:      "socks-inbound",
		Protocol: "socks",
		Settings: SocksSettings{
			Auth: "noauth",
			UDP:  false,
			IP:   "127.0.0.1",
		},
		Sniffing: SniffingSetting{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}
}
