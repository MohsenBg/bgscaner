package dns

import (
	"bgscan/internal/core/process"
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"
)

// SlipstreamClientPaths returns the candidate filesystem paths where the
// slipstream-client binary may be located.
func SlipstreamClientPaths() []string {
	return []string{
		"assets/slipstream-client",
		"assets/dns/slipstream-client",
		"slipstream-client",
	}
}

// SlipstreamClient wraps the external slipstream-client binary and manages
// the lifecycle of a DNS tunneling session.
type SlipstreamClient struct {
	bin      string
	domain   string
	certPath string
	dnsPort  uint16
	process  *process.Process
}

// NewSlipstreamClient initializes a Slipstream client wrapper.
//
// The function resolves the slipstream-client binary automatically using
// FindSlipstreamClient and prepares the runtime configuration.
func NewSlipstreamClient(domain string, dnsPort uint16, certPath string) (*SlipstreamClient, error) {
	path, err := FindSlipstreamClient()
	if err != nil {
		return nil, err
	}

	return &SlipstreamClient{
		bin:      path,
		domain:   domain,
		certPath: certPath,
		dnsPort:  dnsPort,
	}, nil
}

// FindSlipstreamClient searches for the slipstream-client binary in common
// project paths and returns the resolved executable path.
func FindSlipstreamClient() (string, error) {
	return process.FindBinaryInPaths("slipstream-client", SlipstreamClientPaths())
}

// RunTunnel starts a Slipstream DNS tunnel.
//
// The command executed resembles:
//
//	slipstream-client -d <domain> -r <resolver>:<port> -l <listenPort> [--cert <cert>]
//
// The process is managed via the internal process package and is stored
// in the client instance for later shutdown.
func (s *SlipstreamClient) RunTunnel(ctx context.Context, ip string, listenPort uint16) (*process.Process, error) {
	args := []string{
		"-d", s.domain,
		"-r", net.JoinHostPort(ip, fmt.Sprint(s.dnsPort)),
		"-l", fmt.Sprintf("%d", listenPort),
	}

	if s.certPath != "" {
		args = append(args, "--cert", s.certPath)
	}

	proc, err := process.Start(ctx, s.bin, args...)
	if err != nil {
		return nil, err
	}

	s.process = proc
	return proc, nil
}

// StopTunnel gracefully terminates the running Slipstream tunnel process.
//
// If no process is currently active, an error is returned.
func (s *SlipstreamClient) StopTunnel(ctx context.Context) error {
	if s.process == nil {
		return fmt.Errorf("slipstream-client process not running")
	}

	return s.process.StopGracefully(2 * time.Second)
}

// VerifySlipstreamClient performs a basic health check of the
// slipstream-client binary by executing the "--help" command.
//
// This ensures the binary exists and can start successfully.
func VerifySlipstreamClient() error {
	path, err := FindSlipstreamClient()
	if err != nil {
		return fmt.Errorf("slipstream-client not found: %w", err)
	}

	cmd := exec.Command(path, "--help")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("slipstream-client failed to start: %w", err)
	}

	return nil
}
