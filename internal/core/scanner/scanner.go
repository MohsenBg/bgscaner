// Package scanner provides high-level scanning orchestration built on top
// of the scanning engine and probe implementations.
//
// A Scanner coordinates:
//   - loading configuration
//   - preparing input IP lists
//   - initializing probes
//   - running the scan engine
//   - writing results
//
// Different scanner implementations exist for different probe types
// such as ICMP, TCP, HTTP, and Xray-based scans.
package scanner

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/iplist"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/probe"
	"context"
	"fmt"
	"os"
	"time"
)

// ScanMode represents the type of scan performed by a Scanner.
type ScanMode int

const (
	// ICMP_SCAN performs ping-based reachability checks.
	ICMP_SCAN ScanMode = iota

	// TCP_SCAN attempts TCP connection probes.
	TCP_SCAN

	// HTTP_SCAN performs HTTP requests against targets.
	HTTP_SCAN

	// XRAY_SCAN runs an Xray proxy to test connectivity and bandwidth.
	XRAY_SCAN
)

// ScanHooks is re-exported from engine to simplify usage for callers.
// Hooks allow external components to observe scan progress.
type ScanHooks = engine.ScanHooks

// Scanner defines the public interface implemented by all scan types.
//
// Implementations differ only in probe behavior and configuration
// but share the same lifecycle.
type Scanner interface {

	// Mode returns the scan type.
	Mode() ScanMode

	// Input returns the path to the IP input file used for scanning.
	Input() string

	// Pause temporarily halts scan progress.
	Pause()

	// IsPaused reports whether the scan is currently paused.
	IsPaused() bool

	// Resume continues a paused scan.
	Resume()

	// Close cancels the scan and releases resources.
	Close()

	// PreProcess performs preparation steps such as IP shuffling.
	PreProcess() error

	// Scan executes the scan using the configured probe.
	Scan(hooks ScanHooks)
}

// baseScanner contains shared state and logic used by all scanner types.
// Concrete scanners embed this struct and only provide probe/config logic.
type baseScanner struct {
	ctx    context.Context
	cancel context.CancelFunc

	pause  *engine.PauseController
	writer *result.Writer
	probe  probe.Probe

	input        string // path to IP file (may be shuffled)
	shuffledFile string // temp file to delete when closing

	mode ScanMode
	rate int

	cfgGeneral *config.GeneralConfig
	cfgWriter  *config.WriterConfig
}

func (b *baseScanner) Input() string  { return b.input }
func (b *baseScanner) Mode() ScanMode { return b.mode }

func (b *baseScanner) Pause()         { b.pause.Pause() }
func (b *baseScanner) Resume()        { b.pause.Resume() }
func (b *baseScanner) IsPaused() bool { return b.pause.IsPaused() }

// Close cancels the scanner context and cleans up temporary resources.
func (b *baseScanner) Close() {
	if b.cancel != nil {
		b.cancel()
	}

	if b.shuffledFile != "" {
		_ = os.Remove(b.shuffledFile)
		b.shuffledFile = ""
	}
}

// run invokes the scanning engine with the provided worker count.
func (b *baseScanner) run(workers int, hooks ScanHooks) {
	engine.RunScan(
		b.ctx,
		workers,
		b.rate,
		b.input,
		b.writer,
		b.probe,
		hooks,
		b.pause,
	)
}

// calcRate estimates a safe request rate based on probe duration.
//
// Formula:
//
//	cyclesPerSecond = 1s / minProbeTime
//	rate = cyclesPerSecond * workers
//
// This helps avoid overloading the system with excessive requests.
func calcRate(workers int, minProbeTime time.Duration) int {
	cyclesPerSecond := int(time.Second / minProbeTime)
	return cyclesPerSecond * workers
}

//
// ICMP Scanner
//

// ICMPScanner performs reachability checks using ICMP probes.
type ICMPScanner struct {
	baseScanner
	cfgICMP *config.ICMPConfig
}

func (s *ICMPScanner) Scan(hooks ScanHooks) {
	s.run(s.cfgICMP.Workers, hooks)
}

// PreProcess prepares the input list before scanning.
func (s *ICMPScanner) PreProcess() error {
	ipFile, shuffled, err := preprocessShuffle(s.ctx, s.input, s.cfgICMP.ShuffleIPs)
	if err != nil {
		s.cancel()
		return fmt.Errorf("shuffle IPs: %w", err)
	}

	s.baseScanner.input = ipFile
	s.baseScanner.shuffledFile = shuffled
	return nil
}

// NewICMPScanner constructs a scanner configured for ICMP probing.
func NewICMPScanner(ctx context.Context, input string) (Scanner, error) {

	cfgGeneral, cfgICMP, cfgWriter, err := loadConfigs(
		config.GetGeneral,
		config.GetICMP,
		config.GetWriter,
	)
	if err != nil {
		return nil, err
	}

	file, err := result.BuildResultFilePath(result.ICMPResultDir, cfgICMP.PrefixOutput)
	if err != nil {
		return nil, fmt.Errorf("build result path: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)

	writer, err := result.NewWriter(file, writerToResultCfg(cfgWriter), scanCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create writer: %w", err)
	}

	icmpProbe, err := probe.NewICMPProbe(cfgICMP.Timeout.Duration())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create ICMP probe: %w", err)
	}

	const minProbeTime = 25 * time.Millisecond

	return &ICMPScanner{
		baseScanner: baseScanner{
			ctx:          scanCtx,
			cancel:       cancel,
			input:        input,
			shuffledFile: "",
			cfgGeneral:   cfgGeneral,
			cfgWriter:    cfgWriter,
			writer:       writer,
			mode:         ICMP_SCAN,
			probe:        icmpProbe,
			rate:         calcRate(cfgICMP.Workers, minProbeTime),
			pause:        engine.NewPauseController(),
		},
		cfgICMP: cfgICMP,
	}, nil
}

//
// TCP Scanner
//

// TCPScanner performs TCP connection checks against a specific port.
type TCPScanner struct {
	baseScanner
	cfgTCP *config.TCPConfig
}

func (s *TCPScanner) Scan(hooks ScanHooks) {
	s.run(s.cfgTCP.Workers, hooks)
}

func (s *TCPScanner) PreProcess() error {
	ipFile, shuffled, err := preprocessShuffle(s.ctx, s.input, s.cfgTCP.ShuffleIPs)
	if err != nil {
		s.cancel()
		return fmt.Errorf("shuffle IPs: %w", err)
	}

	s.baseScanner.input = ipFile
	s.baseScanner.shuffledFile = shuffled
	return nil
}

// NewTCPScanner constructs a TCP scanning instance.
func NewTCPScanner(ctx context.Context, input string) (Scanner, error) {

	cfgGeneral, cfgTCP, cfgWriter, err := loadConfigs(
		config.GetGeneral,
		config.GetTCP,
		config.GetWriter,
	)
	if err != nil {
		return nil, err
	}

	file, err := result.BuildResultFilePath(result.TCPResultDir, cfgTCP.PrefixOutput)
	if err != nil {
		return nil, fmt.Errorf("build result path: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)

	writer, err := result.NewWriter(file, writerToResultCfg(cfgWriter), scanCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create writer: %w", err)
	}

	tcpProbe := probe.NewTCPProbe(
		fmt.Sprint(cfgTCP.Port),
		cfgTCP.Timeout.Duration(),
	)

	const minProbeTime = 50 * time.Millisecond

	return &TCPScanner{
		baseScanner: baseScanner{
			ctx:          scanCtx,
			cancel:       cancel,
			input:        input,
			shuffledFile: "",
			cfgGeneral:   cfgGeneral,
			cfgWriter:    cfgWriter,
			writer:       writer,
			mode:         TCP_SCAN,
			probe:        tcpProbe,
			rate:         calcRate(cfgTCP.Workers, minProbeTime),
			pause:        engine.NewPauseController(),
		},
		cfgTCP: cfgTCP,
	}, nil
}

//
// HTTP Scanner
//

// HTTPScanner performs HTTP requests against targets and records responses.
type HTTPScanner struct {
	baseScanner
	cfgHTTP *config.HTTPConfig
}

func (s *HTTPScanner) Scan(hooks ScanHooks) {
	s.run(s.cfgHTTP.Workers, hooks)
}

func (s *HTTPScanner) PreProcess() error {
	ipFile, shuffled, err := preprocessShuffle(s.ctx, s.input, s.cfgHTTP.ShuffleIPs)
	if err != nil {
		s.cancel()
		return fmt.Errorf("shuffle IPs: %w", err)
	}

	s.baseScanner.input = ipFile
	s.baseScanner.shuffledFile = shuffled
	return nil
}

// NewHTTPScanner constructs a scanner configured for HTTP probing.
func NewHTTPScanner(ctx context.Context, input string) (Scanner, error) {

	cfgGeneral, cfgHTTP, cfgWriter, err := loadConfigs(
		config.GetGeneral,
		config.GetHTTP,
		config.GetWriter,
	)
	if err != nil {
		return nil, err
	}

	file, err := result.BuildResultFilePath(result.HTTPResultDir, cfgHTTP.PrefixOutput)
	if err != nil {
		return nil, fmt.Errorf("build result path: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)

	writer, err := result.NewWriter(file, writerToResultCfg(cfgWriter), scanCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create writer: %w", err)
	}

	reqCfg, err := probe.NewHTTPRequestFromConfig(*cfgHTTP)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create HTTP request config: %w", err)
	}

	httpProbe := probe.NewHTTPProbe(*reqCfg)

	const minProbeTime = 80 * time.Millisecond

	return &HTTPScanner{
		baseScanner: baseScanner{
			ctx:          scanCtx,
			cancel:       cancel,
			input:        input,
			shuffledFile: "",
			cfgGeneral:   cfgGeneral,
			cfgWriter:    cfgWriter,
			writer:       writer,
			mode:         HTTP_SCAN,
			probe:        httpProbe,
			rate:         calcRate(cfgHTTP.Workers, minProbeTime),
			pause:        engine.NewPauseController(),
		},
		cfgHTTP: cfgHTTP,
	}, nil
}

//
// Xray Scanner
//

// XrayScanner uses an external Xray process to test proxy connectivity
// and optionally measure bandwidth.
type XrayScanner struct {
	baseScanner
	cfgXray *config.XrayConfig
}

func (s *XrayScanner) Scan(hooks ScanHooks) {
	s.run(s.cfgXray.Workers, hooks)
}

func (s *XrayScanner) PreProcess() error {
	ipFile, shuffled, err := preprocessShuffle(s.ctx, s.input, s.cfgXray.ShuffleIPs)
	if err != nil {
		s.cancel()
		return fmt.Errorf("shuffle IPs: %w", err)
	}

	s.baseScanner.input = ipFile
	s.baseScanner.shuffledFile = shuffled
	return nil
}

// NewXrayScanner constructs a scanner that executes Xray-based probes.
func NewXrayScanner(ctx context.Context, input string, template string) (Scanner, error) {

	cfgGeneral, cfgXray, cfgWriter, err := loadConfigs(
		config.GetGeneral,
		config.GetXray,
		config.GetWriter,
	)
	if err != nil {
		return nil, err
	}

	file, err := result.BuildResultFilePath(result.XRAYResultDir, cfgXray.PrefixOutput)
	if err != nil {
		return nil, fmt.Errorf("build result path: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)

	writer, err := result.NewWriter(file, writerToResultCfg(cfgWriter), scanCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create writer: %w", err)
	}

	xrayProbe, err := probe.NewXrayProbe(cfgXray, template)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create Xray probe: %w", err)
	}

	const minProbeTime = 200 * time.Millisecond

	return &XrayScanner{
		baseScanner: baseScanner{
			ctx:          scanCtx,
			cancel:       cancel,
			input:        input,
			shuffledFile: input,
			cfgGeneral:   cfgGeneral,
			cfgWriter:    cfgWriter,
			writer:       writer,
			mode:         XRAY_SCAN,
			probe:        xrayProbe,
			rate:         calcRate(cfgXray.Workers, minProbeTime),
			pause:        engine.NewPauseController(),
		},
		cfgXray: cfgXray,
	}, nil
}

//
// Helpers
//

// shuffleIfNeeded optionally shuffles an IP list file.
//
// If shuffling is enabled, the function creates a temporary shuffled file
// using a memory-safe algorithm suitable for very large IP lists.
func shuffleIfNeeded(ctx context.Context, path string, shouldShuffle bool) (filePath, tempFile string, err error) {
	if !shouldShuffle {
		return path, "", nil
	}

	shuffled, err := iplist.ShuffleFileFullyMemorySafe(ctx, path)
	if err != nil {
		return "", "", err
	}

	return shuffled, shuffled, nil
}

// writerToResultCfg converts WriterConfig into result.Config.
func writerToResultCfg(cfg *config.WriterConfig) result.Config {
	return result.Config{
		DeltaFlushInterval: cfg.DeltaFlushInterval.Duration(),
		MergeFlushInterval: cfg.MergeFlushInterval.Duration(),
		ChanSize:           cfg.ChanSize,
		BufferSize:         cfg.BufferSize,
	}
}

// preprocessShuffle prepares the input file before scanning.
func preprocessShuffle(ctx context.Context, input string, shouldShuffle bool) (string, string, error) {
	ipFile, shuffled, err := shuffleIfNeeded(ctx, input, shouldShuffle)
	if err != nil {
		return "", "", err
	}
	return ipFile, shuffled, nil
}

// loadConfigs retrieves and validates three configuration objects.
//
// It uses generics so each scanner constructor can load its own config types
// while sharing the same validation logic.
func loadConfigs[A, B, C any](
	getA func() *A,
	getB func() *B,
	getC func() *C,
) (*A, *B, *C, error) {

	a := getA()
	b := getB()
	c := getC()

	if a == nil {
		return nil, nil, nil, fmt.Errorf("config %T is nil", a)
	}
	if b == nil {
		return nil, nil, nil, fmt.Errorf("config %T is nil", b)
	}
	if c == nil {
		return nil, nil, nil, fmt.Errorf("config %T is nil", c)
	}

	return a, b, c, nil
}
