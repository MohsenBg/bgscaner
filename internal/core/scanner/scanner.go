package scanner

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/dns"
	"bgscan/internal/core/iplist"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"context"
	"fmt"
	"time"
)

//
// ────────────────────────────────────────────────────────────────
// Scan Modes
// ────────────────────────────────────────────────────────────────
//

// ScanMode represents the logical type of a scanning stage.
//
// Each mode determines the implementation of Probe, configuration
// source, and how results are persisted.
type ScanMode string

const (
	ICMP_SCAN       ScanMode = "ICMP"
	TCP_SCAN        ScanMode = "TCP"
	HTTP_SCAN       ScanMode = "HTTP"
	XRAY_SCAN       ScanMode = "Xray"
	RESOLVE_SCAN    ScanMode = "Resolve"
	DNSTT_SCAN      ScanMode = "DNSTT"
	SLIPSTREAM_SCAN ScanMode = "Slipstream"
)

//
// ────────────────────────────────────────────────────────────────
// Preprocessors
// ────────────────────────────────────────────────────────────────
//

// InputPreprocessor defines a processing step that transforms the raw
// input file path into a new file path. Typical examples include:
//
//   - Shuffling IP list (randomization)
//   - Deduplication
//   - Filtering
//
// Preprocessors are executed before any scanning stage is run.
type InputPreprocessor interface {
	Process(ctx context.Context, input string) (string, error)
}

// ShufflePreprocessor randomly shuffles all IP entries in the input
// file while remaining memory‑safe for large lists.
type ShufflePreprocessor struct{}

// Process applies full‑file shuffling using a memory‑safe algorithm.
func (p *ShufflePreprocessor) Process(ctx context.Context, input string) (string, error) {
	return shuffleFile(ctx, input)
}

//
// ────────────────────────────────────────────────────────────────
// StageConfig
// ────────────────────────────────────────────────────────────────
//

// StageConfig represents an executable scanning stage.
//
// Each stage is configured with:
//
//   - Mode:    high‑level scan type
//   - Workers: concurrency degree
//   - Probe:   network probing implementation
//   - Writer:  result Writer for this stage
//   - Rate:    global rate limit for stage
//   - Hooks:   extended lifecycle callbacks
//
// Stages can be executed individually or chained (RunChain).
type StageConfig struct {
	Mode    ScanMode
	Workers int
	Probe   probe.Probe
	Writer  *result.Writer
	Rate    int
	Hooks   engine.ScanHooks
}

// AddHooks attaches additional scan hooks to the stage.
func (s *StageConfig) AddHooks(h engine.ScanHooks) *StageConfig {
	s.Hooks = h
	return s
}

//
// ────────────────────────────────────────────────────────────────
// Scanner
// ────────────────────────────────────────────────────────────────
//

// Scanner orchestrates multi‑stage IP scanning operations.
//
// Responsibilities:
//
//   - Manages lifecycle context for all scan stages
//   - Runs input preprocessors
//   - Provides DI for PortManager
//   - Executes stages (single or chained)
//   - Supports pause/resume functionality
//
// Scanner does NOT run Probe.Init() automatically—initialization is
// left to the Engine layer or Stage builders.
type Scanner struct {
	ctx           context.Context
	cancel        context.CancelFunc
	pause         *engine.PauseController
	input         string
	pm            *portmgr.PortManager
	preprocessors []InputPreprocessor
	stages        []StageConfig
}

// NewScanner creates a new Scanner instance with an isolated context
// and an internal PortManager.
//
// The returned Scanner owns the context and must be closed via Close()
// to cancel all internal goroutines.
func NewScanner(ctx context.Context, input string) *Scanner {
	scanCtx, cancel := context.WithCancel(ctx)

	var poolSize uint16 = 3000
	pm, _ := portmgr.NewPortManager(portmgr.RandomBasePort(poolSize), poolSize)

	return &Scanner{
		ctx:           scanCtx,
		cancel:        cancel,
		pause:         engine.NewPauseController(),
		pm:            pm,
		input:         input,
		preprocessors: make([]InputPreprocessor, 0),
		stages:        make([]StageConfig, 0),
	}
}

// GetStages returns all configured scanning stages.
func (s *Scanner) GetStages() []StageConfig {
	return s.stages
}

// AddPreprocessor appends an input preprocessing step.
func (s *Scanner) AddPreprocessor(p InputPreprocessor) {
	s.preprocessors = append(s.preprocessors, p)
}

// prepareInput applies all preprocessors in sequence.
func (s *Scanner) prepareInput() (string, error) {
	current := s.input
	for _, p := range s.preprocessors {
		next, err := p.Process(s.ctx, current)
		if err != nil {
			return "", fmt.Errorf("preprocessor %T failed: %w", p, err)
		}
		current = next
	}
	return current, nil
}

//
// ────────────────────────────────────────────────────────────────
// Stage Storage
// ────────────────────────────────────────────────────────────────
//

// AddStage registers a new StageConfig.
func (s *Scanner) AddStage(stage StageConfig) {
	s.stages = append(s.stages, stage)
}

// Run executes all configured stages.
//
// Behavior:
//
//   - No stage → panic
//   - 1 stage → RunSingle
//   - N stages → RunChain
//
// PortManager.Start() is invoked automatically.
func (s *Scanner) Run() {
	if len(s.stages) == 0 {
		panic("no stages added to Scanner")
	}

	s.pause.Start()
	if len(s.stages) == 1 {
		stg := s.stages[0]
		s.runSingle(stg, stg.Hooks)
		return
	}
	s.runChain(s.stages)
	s.pm.Close()
	s.pause.Stop()
}

//
// ────────────────────────────────────────────────────────────────
// RunSingle / RunChain
// ────────────────────────────────────────────────────────────────
//

// RunSingle executes a single scan stage from the prepared input file.
func (s *Scanner) runSingle(stage StageConfig, hooks engine.ScanHooks) {
	input, err := s.prepareInput()
	if err != nil {
		if hooks.OnError != nil {
			hooks.OnError(err)
		}
		return
	}

	engine.RunScan(s.ctx, input, config.GetGeneral().MaxIPsToTest, engine.ScanConfig{
		Workers: stage.Workers,
		Probe:   stage.Probe,
		Writer:  stage.Writer,
		Rate:    stage.Rate,
		Hooks:   hooks,
	}, s.pause)
}

// RunChain executes multiple stages sequentially in a chain.
//
// The ChainConfig determines how IPs flow between stages and how
// intermediate results are handled.
func (s *Scanner) runChain(stages []StageConfig) {
	input, err := s.prepareInput()
	if err != nil {
		if stages[0].Hooks.OnError != nil {
			stages[0].Hooks.OnError(err)
		}
		return
	}

	engineStages := make([]engine.ScanConfig, len(stages))
	for i, stage := range stages {
		engineStages[i] = engine.ScanConfig{
			Workers: stage.Workers,
			Probe:   stage.Probe,
			Writer:  stage.Writer,
			Rate:    stage.Rate,
			Hooks:   stage.Hooks,
		}
	}

	engine.RunScanWithChain(s.ctx, input, config.GetGeneral().MaxIPsToTest, &engine.ChainConfig{
		Mode:   engine.ParseChainMode(config.GetGeneral().ChainMode),
		Stages: engineStages,
		Pause:  s.pause,
	})
}

//
// ────────────────────────────────────────────────────────────────
// Lifecycle Helpers
// ────────────────────────────────────────────────────────────────
//

// Pause suspends all Engine‑controlled scan workers.
func (s *Scanner) Pause() { s.pause.Pause() }

// Resume resumes worker activity after Pause().
func (s *Scanner) Resume() { s.pause.Resume() }

// IsPaused reports whether the scanner is currently paused.
func (s *Scanner) IsPaused() bool { return s.pause.IsPaused() }

// PausedDuration returns the total amount of time scanning has been paused.
func (s *Scanner) PausedDuration() time.Duration { return s.pause.PausedDuration() }

// Close cancels the Scanner context and terminates internal goroutines.
// PortManager is cleaned up automatically inside Run().
func (s *Scanner) Close() { s.cancel() }

//
// ────────────────────────────────────────────────────────────────
// Stage Builders
// ────────────────────────────────────────────────────────────────
//

// BuildICMPStage constructs a StageConfig for ICMP probing.
//
// The stage includes:
//   - ICMPProbe
//   - result.Writer
//   - Workers / Rate settings
func (s *Scanner) BuildICMPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetICMP()

	file, err := result.BuildResultFilePath(result.ICMPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewICMPProbe(cfg.Timeout.Duration(), cfg.Tries)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    ICMP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 25*time.Millisecond),
	}, nil
}

// BuildTCPStage constructs a TCP port‑checking stage.
func (s *Scanner) BuildTCPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetTCP()

	file, err := result.BuildResultFilePath(result.TCPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb := probe.NewTCPProbe(fmt.Sprint(cfg.Port), cfg.Timeout.Duration(), cfg.Tries)

	return StageConfig{
		Mode:    TCP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 50*time.Millisecond),
	}, nil
}

// BuildHTTPStage constructs an HTTP probing stage.
func (s *Scanner) BuildHTTPStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetHTTP()

	file, err := result.BuildResultFilePath(result.HTTPResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	reqCfg, err := probe.NewHTTPRequestFromConfig(*cfg)
	if err != nil {
		return StageConfig{}, err
	}

	prb := probe.NewHTTPProbe(*reqCfg)

	return StageConfig{
		Mode:    HTTP_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 80*time.Millisecond),
	}, nil
}

// BuildXrayStage constructs an Xray connectivity test stage.
func (s *Scanner) BuildXrayStage(ctx context.Context, template string) (StageConfig, error) {
	cfg := config.GetXray()

	file, err := result.BuildResultFilePath(result.XRAYResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewXrayProbe(cfg, template, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    XRAY_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 200*time.Millisecond),
	}, nil
}

// BuildResolveStage constructs a DNS resolver testing stage.
func (s *Scanner) BuildResolveStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().Resolver

	file, err := result.BuildResultFilePath(result.ResolveResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, r := range cfg.AcceptedRCodes {
		rcodes = append(rcodes, uint16(dns.ParseDNSRcode(r)))
	}

	prb := probe.NewResolverProbe(&probe.DnsRequest{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		RandomSubdomain: cfg.RandomSubdomain,
		DpiCheck:        cfg.CheckDPI,
		DpiTimeout:      cfg.DPITimeout.Duration(),
		DpiTries:        cfg.DPITries,
		Edns0Size:       cfg.EDNSBufSize,
		CheckTypes:      cfg.CheckTypes,
		AcceptedRcodes:  rcodes,
		Timeout:         cfg.Timeout.Duration(),
		Transport:       dns.ParseTransport(cfg.Protocol),
		Tries:           cfg.Tries,
	})

	return StageConfig{
		Mode:    RESOLVE_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, 500*time.Millisecond),
	}, nil
}

// BuildDNSTTStage constructs a DNSTT tunneling test stage.
func (s *Scanner) BuildDNSTTStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().DNSTT
	transport := config.GetDNS().Resolver.Protocol
	port := config.GetDNS().Resolver.Port

	file, err := result.BuildResultFilePath(result.DNSTTResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewDNSTTProbe(probe.DNSTTConfig{
		Domain:    cfg.Domain,
		PubKey:    cfg.PublicKey,
		Transport: dns.ParseTransport(transport),
		DNSPort:   port,
		Timeout:   cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    DNSTT_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, time.Second),
	}, nil
}

// BuildSlipStreamStage constructs a SlipStream DPI‑evading resolver stage.
func (s *Scanner) BuildSlipStreamStage(ctx context.Context) (StageConfig, error) {
	cfg := config.GetDNS().SlipStream
	port := config.GetDNS().Resolver.Port

	file, err := result.BuildResultFilePath(result.SlipStreamResultDir, cfg.PrefixOutput)
	if err != nil {
		return StageConfig{}, err
	}
	writer, err := result.NewWriter(file, writerConfig(), ctx)
	if err != nil {
		return StageConfig{}, err
	}

	prb, err := probe.NewSlipstreamProbe(cfg.Workers, probe.SlipstreamConfig{
		Domain:   cfg.Domain,
		CertPath: cfg.CertPath,
		DNSPort:  port,
		Timeout:  cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, err
	}

	return StageConfig{
		Mode:    SLIPSTREAM_SCAN,
		Workers: cfg.Workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(cfg.Workers, time.Second),
	}, nil
}

//
// ────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────
//

// calcRate computes a global per‑stage maximum event rate given
// worker count and minimum expected probe execution time.
func calcRate(workers int, minProbeTime time.Duration) int {
	return int(time.Second/minProbeTime) * workers
}

// writerConfig returns a result.Writer configuration sourced
// from global writer settings.
func writerConfig() result.Config {
	cfg := config.GetWriter()
	return result.Config{
		MergeFlushInterval: cfg.MergeFlushInterval.Duration(),
		ChanSize:           cfg.ChanSize,
		BatchSize:          cfg.BatchSize,
	}
}

// shuffleFile applies memory‑safe shuffling for IP lists.
func shuffleFile(ctx context.Context, path string) (string, error) {
	shuffled, err := iplist.ShuffleFileFullyMemorySafe(ctx, path)
	if err != nil {
		return "", fmt.Errorf("shuffle failed for %q: %w", path, err)
	}
	return shuffled, nil
}
