package probe

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/xray"
	"context"
	"net"
	"os"
	"sync"
	"time"
)

// defaultXrayWarmup defines the delay after starting the Xray process
// before running connectivity tests. This gives Xray enough time to
// initialize its inbound listener and routing pipeline.
const defaultXrayWarmup = 500 * time.Millisecond

// XrayProbe implements the Probe interface using an external Xray process
// to verify connectivity and optionally measure bandwidth.
//
// Each probe execution:
//
//  1. Reserves a temporary local TCP port.
//  2. Generates a runtime Xray configuration targeting the given IP.
//  3. Launches an Xray process using that configuration.
//  4. Waits briefly for initialization (warm‑up).
//  5. Measures latency and optionally download/upload performance.
//  6. Terminates the Xray process and cleans up temporary resources.
type XrayProbe struct {
	templateName  string
	timeout       time.Duration
	testMode      config.ConnectivityTest
	downloadBytes int64
	uploadBytes   int64
	warmup        time.Duration
}

// NewXrayProbe creates a new XrayProbe using the provided Xray configuration
// and template name. The template defines the outbound protocol and base
// configuration used for generating the runtime Xray configuration.
func NewXrayProbe(cfg *config.XrayConfig, templateName string) (Probe, error) {
	if _, err := xray.GetTemplateByName(templateName); err != nil {
		return nil, err
	}

	return &XrayProbe{
		templateName:  templateName,
		timeout:       cfg.Timeout.Duration(),
		testMode:      cfg.ConnectivityTestType,
		downloadBytes: int64(cfg.DownloadSpeed) * 1024,
		uploadBytes:   int64(cfg.UploadSpeed) * 1024,
		warmup:        defaultXrayWarmup,
	}, nil
}

// Run executes the probe against the specified IP address.
//
// The method dynamically generates an Xray configuration, starts an Xray
// process, and performs the configured connectivity and performance tests.
// It respects the provided context and will terminate early if cancelled.
func (p *XrayProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {

	// Abort early if the context has already been cancelled.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Reserve a local TCP port for the Xray inbound proxy.
	port, err := reserveFreePort()
	if err != nil {
		return nil, err
	}
	defer releasePort(port)

	// Generate a temporary Xray configuration for the target IP.
	configPath, err := xray.GenerateConfig(p.templateName, ip, port)
	if err != nil {
		return nil, err
	}
	defer os.Remove(configPath)

	// Validate the generated configuration before starting Xray.
	if err := xray.ValidateConfig(configPath); err != nil {
		return nil, err
	}

	// Start the Xray process.
	proc, err := xray.StartXray(ctx, configPath)
	if err != nil {
		return nil, err
	}
	defer proc.Kill()

	// Wait for Xray to finish initializing.
	if err := p.waitWarmup(ctx); err != nil {
		return nil, err
	}

	// Measure latency through the proxy.
	latency, err := xray.MeasureLatency(p.timeout, port)
	if err != nil {
		return nil, err
	}

	res := &result.IPScanResult{
		IP:      ip,
		Latency: latency,
	}

	// Execute additional tests depending on the configured test mode.
	switch p.testMode {

	case config.ConnectivityOnly:
		return res, nil

	case config.DownloadSpeedOnly:
		res.Download, err = xray.MeasureDownloadDuration(
			p.timeout, p.downloadBytes, port,
		)
		if err != nil {
			return nil, err
		}

	case config.UploadSpeedOnly:
		res.Upload, err = xray.MeasureUploadDuration(
			p.timeout, p.uploadBytes, port,
		)
		if err != nil {
			return nil, err
		}

	case config.Both:
		res.Download, err = xray.MeasureDownloadDuration(
			p.timeout, p.downloadBytes, port,
		)
		if err != nil {
			return nil, err
		}

		res.Upload, err = xray.MeasureUploadDuration(
			p.timeout, p.uploadBytes, port,
		)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// waitWarmup pauses execution for the configured warm‑up duration.
// This allows the Xray process to initialize its network listeners
// before tests begin. The wait respects context cancellation.
func (p *XrayProbe) waitWarmup(ctx context.Context) error {
	if p.warmup <= 0 {
		return nil
	}

	timer := time.NewTimer(p.warmup)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// portMu protects access to the portUsed registry.
var (
	portMu   sync.Mutex
	portUsed = make(map[int]struct{})
)

// reserveFreePort finds and reserves a free local TCP port.
//
// The function first asks the OS for an available port by binding to
// "127.0.0.1:0". The selected port is then recorded in an in‑process
// registry to avoid collisions between concurrent probes.
//
// This does not prevent external processes from claiming the port,
// but it eliminates conflicts between workers within the scanner.
func reserveFreePort() (int, error) {
	for {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return 0, err
		}

		port := l.Addr().(*net.TCPAddr).Port
		l.Close()

		portMu.Lock()
		if _, exists := portUsed[port]; !exists {
			portUsed[port] = struct{}{}
			portMu.Unlock()
			return port, nil
		}
		portMu.Unlock()
	}
}

// releasePort removes a previously reserved port from the registry,
// making it available for future probes.
func releasePort(port int) {
	portMu.Lock()
	delete(portUsed, port)
	portMu.Unlock()
}

func (p *XrayProbe) Close() error {
	return nil
}
