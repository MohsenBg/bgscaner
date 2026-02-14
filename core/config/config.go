package config

import (
	"bgscan/core/errx"
	"bgscan/core/filemanager"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// ═══════════════════════════════════════════════════════════
// Global Singleton Instance
// ═══════════════════════════════════════════════════════════

var (
	instance *ScannerConfig
	once     sync.Once
	mu       sync.RWMutex
)

// Get returns the global config instance (initializes if needed)
func Get() *ScannerConfig {
	once.Do(func() {
		instance = &ScannerConfig{
			GeneralConfig: DefaultGeneralConfig(),
			ICMPConfig:    DefaultICMPConfig(),
			TCPConfig:     DefaultTCPConfig(),
			HTTPConfig:    DefaultHTTPConfig(),
			XrayConfig:    DefaultXrayConfig(),
		}
	})
	return instance
}

// ═══════════════════════════════════════════════════════════
// Thread-Safe Getters
// ═══════════════════════════════════════════════════════════

func GetGeneral() *GeneralConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().GeneralConfig
}

func GetICMP() *ICMPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().ICMPConfig
}

func GetTCP() *TCPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().TCPConfig
}

func GetHTTP() *HTTPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().HTTPConfig
}

func GetXray() *XrayConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().XrayConfig
}

// ═══════════════════════════════════════════════════════════
// Thread-Safe Setters
// ═══════════════════════════════════════════════════════════

func SetGeneral(cfg *GeneralConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().GeneralConfig = cfg
}

func SetICMP(cfg *ICMPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().ICMPConfig = cfg
}

func SetTCP(cfg *TCPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().TCPConfig = cfg
}

func SetHTTP(cfg *HTTPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().HTTPConfig = cfg
}

func SetXray(cfg *XrayConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().XrayConfig = cfg
}

// ═══════════════════════════════════════════════════════════
// Enums
// ═══════════════════════════════════════════════════════════

type ConnectivityTest int

const (
	ConnectivityOnly ConnectivityTest = iota
	DownloadSpeedOnly
	UploadSpeedOnly
	Both
)

func (c ConnectivityTest) String() string {
	return [...]string{
		"Connectivity Only",
		"Download Speed Only",
		"Upload Speed Only",
		"Both",
	}[c]
}

// ───────────────────────────────────────────────────────────

type ScanType int

const (
	ScanICMP ScanType = iota
	ScanTCP
	ScanHTTP
	ScanXrayConfig
)

func (s ScanType) String() string {
	return [...]string{"ICMP", "TCP", "HTTP", "Xray Config"}[s]
}

// ═══════════════════════════════════════════════════════════
// Protocol-Specific Configs
// ═══════════════════════════════════════════════════════════

type ICMPConfig struct {
	Timeout      time.Duration `json:"timeout"`
	Workers      int           `json:"workers"`
	PrefixOutput string        `json:"prefix_output"`
	ShuffleIPs   bool          `json:"shuffle_ips"`
}

func DefaultICMPConfig() *ICMPConfig {
	return &ICMPConfig{
		Timeout:      5 * time.Second,
		Workers:      100,
		PrefixOutput: "icmp_results",
		ShuffleIPs:   true,
	}
}

// ───────────────────────────────────────────────────────────

type TCPConfig struct {
	Port         int           `json:"port"`
	Timeout      time.Duration `json:"timeout"`
	Workers      int           `json:"workers"`
	PrefixOutput string        `json:"prefix_output"`
	ShuffleIPs   bool          `json:"shuffle_ips"`
}

func DefaultTCPConfig() *TCPConfig {
	return &TCPConfig{
		Port:         80,
		Timeout:      5 * time.Second,
		Workers:      100,
		PrefixOutput: "tcp_results",
		ShuffleIPs:   true,
	}
}

// ───────────────────────────────────────────────────────────

type HTTPConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Timeout      time.Duration `json:"timeout"`
	Workers      int           `json:"workers"`
	PrefixOutput string        `json:"prefix_output"`
	ShuffleIPs   bool          `json:"shuffle_ips"`
}

func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host:         "example.com",
		Port:         443,
		Timeout:      10 * time.Second,
		Workers:      50,
		PrefixOutput: "http_results",
		ShuffleIPs:   true,
	}
}

// ───────────────────────────────────────────────────────────

type XrayConfig struct {
	Timeout              time.Duration    `json:"timeout"`
	Workers              int              `json:"workers"`
	ConnectivityTestType ConnectivityTest `json:"connectivity_test_type"`
	DownloadSpeed        int              `json:"download_speed"` // Mbps
	UploadSpeed          int              `json:"upload_speed"`   // Mbps
}

func DefaultXrayConfig() *XrayConfig {
	return &XrayConfig{
		Timeout:              6 * time.Second,
		Workers:              10,
		ConnectivityTestType: ConnectivityOnly,
		DownloadSpeed:        10,
		UploadSpeed:          5,
	}
}

// ═══════════════════════════════════════════════════════════
// General Config
// ═══════════════════════════════════════════════════════════

type GeneralConfig struct {
	StatusInterval time.Duration `json:"status_interval"`
	FlushInterval  int           `json:"flush_interval"`
	StopAfterFound int           `json:"stop_after_found"`
	MaxIPsToTest   int           `json:"max_ips_to_test"`
	Verbose        bool          `json:"verbose"`
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		StatusInterval: 5 * time.Second,
		FlushInterval:  100,
		StopAfterFound: 0,
		MaxIPsToTest:   0,
		Verbose:        false,
	}
}

// ═══════════════════════════════════════════════════════════
// Main Scanner Config
// ═══════════════════════════════════════════════════════════

type ScannerConfig struct {
	Type          ScanType       `json:"type"`
	IPFile        string         `json:"ip_file"`
	GeneralConfig *GeneralConfig `json:"general_config"`
	ICMPConfig    *ICMPConfig    `json:"icmp_config"`
	TCPConfig     *TCPConfig     `json:"tcp_config"`
	HTTPConfig    *HTTPConfig    `json:"http_config"`
	XrayConfig    *XrayConfig    `json:"xray_config"`
}

// ═══════════════════════════════════════════════════════════
// Config File Paths
// ═══════════════════════════════════════════════════════════

const (
	SettingsDir         = "settings"
	ICMPSettingsFile    = "icmp_settings.json"
	TCPSettingsFile     = "tcp_settings.json"
	HTTPSettingsFile    = "http_settings.json"
	XraySettingsFile    = "xray_settings.json"
	GeneralSettingsFile = "general_settings.json"
)

func GetConfigPath(filename string) (string, error) {
	baseDir, err := filemanager.GetCurrentPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, SettingsDir, filename), nil
}

// ═══════════════════════════════════════════════════════════
// Config Loading (Updates Global Instance)
// ═══════════════════════════════════════════════════════════

func LoadICMPConfig() error {
	path, err := GetConfigPath(ICMPSettingsFile)
	if err != nil {
		return err
	}

	cfg := &ICMPConfig{}
	if err := filemanager.GetJSONFileOrDefault(path, cfg, DefaultICMPConfig()); err != nil {
		return fmt.Errorf("failed to load ICMP config: %w", err)
	}

	SetICMP(cfg)
	return nil
}

func LoadTCPConfig() error {
	path, err := GetConfigPath(TCPSettingsFile)
	if err != nil {
		return err
	}

	cfg := &TCPConfig{}
	if err := filemanager.GetJSONFileOrDefault(path, cfg, DefaultTCPConfig()); err != nil {
		return fmt.Errorf("failed to load TCP config: %w", err)
	}

	SetTCP(cfg)
	return nil
}

func LoadHTTPConfig() error {
	path, err := GetConfigPath(HTTPSettingsFile)
	if err != nil {
		return err
	}

	cfg := &HTTPConfig{}
	if err := filemanager.GetJSONFileOrDefault(path, cfg, DefaultHTTPConfig()); err != nil {
		return fmt.Errorf("failed to load HTTP config: %w", err)
	}

	SetHTTP(cfg)
	return nil
}

func LoadXrayConfig() error {
	path, err := GetConfigPath(XraySettingsFile)
	if err != nil {
		return err
	}

	cfg := &XrayConfig{}
	if err := filemanager.GetJSONFileOrDefault(path, cfg, DefaultXrayConfig()); err != nil {
		return fmt.Errorf("failed to load Xray config: %w", err)
	}

	SetXray(cfg)
	return nil
}

func LoadGeneralConfig() error {
	path, err := GetConfigPath(GeneralSettingsFile)
	if err != nil {
		return err
	}

	cfg := &GeneralConfig{}
	if err := filemanager.GetJSONFileOrDefault(path, cfg, DefaultGeneralConfig()); err != nil {
		return fmt.Errorf("failed to load general config: %w", err)
	}

	SetGeneral(cfg)
	return nil
}

// ═══════════════════════════════════════════════════════════
// Config Saving
// ═══════════════════════════════════════════════════════════

func SaveICMPConfig(cfg *ICMPConfig) error {
	path, err := GetConfigPath(ICMPSettingsFile)
	if err != nil {
		return err
	}

	if err := filemanager.WriteJSONFile(path, cfg); err != nil {
		return fmt.Errorf("failed to save ICMP config: %w", err)
	}

	SetICMP(cfg)
	return nil
}

func SaveTCPConfig(cfg *TCPConfig) error {
	path, err := GetConfigPath(TCPSettingsFile)
	if err != nil {
		return err
	}

	if err := filemanager.WriteJSONFile(path, cfg); err != nil {
		return fmt.Errorf("failed to save TCP config: %w", err)
	}

	SetTCP(cfg)
	return nil
}

func SaveHTTPConfig(cfg *HTTPConfig) error {
	path, err := GetConfigPath(HTTPSettingsFile)
	if err != nil {
		return err
	}

	if err := filemanager.WriteJSONFile(path, cfg); err != nil {
		return fmt.Errorf("failed to save HTTP config: %w", err)
	}

	SetHTTP(cfg)
	return nil
}

func SaveXrayConfig(cfg *XrayConfig) error {
	path, err := GetConfigPath(XraySettingsFile)
	if err != nil {
		return err
	}

	if err := filemanager.WriteJSONFile(path, cfg); err != nil {
		return fmt.Errorf("failed to save Xray config: %w", err)
	}

	SetXray(cfg)
	return nil
}

func SaveGeneralConfig(cfg *GeneralConfig) error {
	path, err := GetConfigPath(GeneralSettingsFile)
	if err != nil {
		return err
	}

	if err := filemanager.WriteJSONFile(path, cfg); err != nil {
		return fmt.Errorf("failed to save general config: %w", err)
	}

	SetGeneral(cfg)
	return nil
}

// ═══════════════════════════════════════════════════════════
// Initialization
// ═══════════════════════════════════════════════════════════

func Init() error {
	if err := LoadGeneralConfig(); err != nil {
		return fmt.Errorf("failed to init general config: %w", err)
	}

	if err := LoadICMPConfig(); err != nil {
		return fmt.Errorf("failed to init ICMP config: %w", err)
	}

	if err := LoadTCPConfig(); err != nil {
		return fmt.Errorf("failed to init TCP config: %w", err)
	}

	if err := LoadHTTPConfig(); err != nil {
		return fmt.Errorf("failed to init HTTP config: %w", err)
	}

	if err := LoadXrayConfig(); err != nil {
		return fmt.Errorf("failed to init Xray config: %w", err)
	}

	return nil
}

// ═══════════════════════════════════════════════════════════
// Validation
// ═══════════════════════════════════════════════════════════

func Validate() error {
	cfg := Get()

	if cfg.IPFile == "" {
		return errx.ErrMissingIPFile
	}

	switch cfg.Type {
	case ScanICMP:
		if cfg.ICMPConfig == nil {
			return errx.ErrMissingICMPConfig
		}
		if cfg.ICMPConfig.Workers < 1 {
			return errx.ErrInvalidWorkerCount
		}

	case ScanTCP:
		if cfg.TCPConfig == nil {
			return errx.ErrMissingTCPConfig
		}
		if cfg.TCPConfig.Port < 1 || cfg.TCPConfig.Port > 65535 {
			return errx.ErrInvalidPort
		}
		if cfg.TCPConfig.Workers < 1 {
			return errx.ErrInvalidWorkerCount
		}

	case ScanHTTP:
		if cfg.HTTPConfig == nil {
			return errx.ErrMissingHTTPConfig
		}
		if cfg.HTTPConfig.Host == "" {
			return errx.ErrMissingHost
		}
		if cfg.HTTPConfig.Port < 1 || cfg.HTTPConfig.Port > 65535 {
			return errx.ErrInvalidPort
		}
		if cfg.HTTPConfig.Workers < 1 {
			return errx.ErrInvalidWorkerCount
		}

	case ScanXrayConfig:
		if cfg.XrayConfig == nil {
			return errx.ErrMissingXrayConfig
		}
		if cfg.XrayConfig.Workers < 1 {
			return errx.ErrInvalidWorkerCount
		}

	default:
		return errx.ErrInvalidScanType
	}

	return nil
}

// GetActiveConfig returns the config for the current scan type
func GetActiveConfig() any {
	cfg := Get()
	switch cfg.Type {
	case ScanICMP:
		return GetICMP()
	case ScanTCP:
		return GetTCP()
	case ScanHTTP:
		return GetHTTP()
	case ScanXrayConfig:
		return GetXray()
	default:
		return nil
	}
}
