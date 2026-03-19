package config

import (
	"bgscan/internal/core/filemanager"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////
// Singleton Instance
////////////////////////////////////////////////////////////////

var (
	instance *ScannerConfig
	once     sync.Once
	mu       sync.RWMutex
)

// Get returns the global scanner configuration instance.
func Get() *ScannerConfig {
	once.Do(func() {
		instance = &ScannerConfig{
			General: DefaultGeneralConfig(),
			Writer:  DefaultWriterConfig(),
			ICMP:    DefaultICMPConfig(),
			TCP:     DefaultTCPConfig(),
			HTTP:    DefaultHTTPConfig(),
			Xray:    DefaultXrayConfig(),
		}
	})
	return instance
}

////////////////////////////////////////////////////////////////
// Thread‑Safe Accessors
////////////////////////////////////////////////////////////////

func GetGeneral() *GeneralConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().General
}

func GetWriter() *WriterConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().Writer
}

func GetICMP() *ICMPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().ICMP
}

func GetTCP() *TCPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().TCP
}

func GetHTTP() *HTTPConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().HTTP
}

func GetXray() *XrayConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Get().Xray
}

////////////////////////////////////////////////////////////////
// Thread‑Safe Setters
////////////////////////////////////////////////////////////////

func setGeneral(cfg *GeneralConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().General = cfg
}

func setWriter(cfg *WriterConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().Writer = cfg
}

func setICMP(cfg *ICMPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().ICMP = cfg
}

func setTCP(cfg *TCPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().TCP = cfg
}

func setHTTP(cfg *HTTPConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().HTTP = cfg
}

func setXray(cfg *XrayConfig) {
	mu.Lock()
	defer mu.Unlock()
	Get().Xray = cfg
}

////////////////////////////////////////////////////////////////
// Connectivity Test Enum
////////////////////////////////////////////////////////////////

// ConnectivityTest defines what type of Xray connectivity test to perform.
type ConnectivityTest uint8

const (
	ConnectivityOnly ConnectivityTest = iota
	DownloadSpeedOnly
	UploadSpeedOnly
	Both
)

func (c ConnectivityTest) String() string {
	names := [...]string{
		"Connectivity Only",
		"Download Speed Only",
		"Upload Speed Only",
		"Both",
	}

	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

////////////////////////////////////////////////////////////////
// Protocol Configs
////////////////////////////////////////////////////////////////

type ICMPConfig struct {
	Timeout      DurationMS `toml:"timeout"`
	Workers      int        `toml:"workers"`
	PrefixOutput string     `toml:"prefix_output"`
	ShuffleIPs   bool       `toml:"shuffle_ips"`
}

func DefaultICMPConfig() *ICMPConfig {
	return &ICMPConfig{
		Timeout:      NewDurationMS(2 * time.Second),
		Workers:      200,
		PrefixOutput: "icmp_",
		ShuffleIPs:   true,
	}
}

type TCPConfig struct {
	Port         int        `toml:"port"`
	Timeout      DurationMS `toml:"timeout"`
	Workers      int        `toml:"workers"`
	PrefixOutput string     `toml:"prefix_output"`
	ShuffleIPs   bool       `toml:"shuffle_ips"`
}

func DefaultTCPConfig() *TCPConfig {
	return &TCPConfig{
		Port:         80,
		Timeout:      NewDurationMS(3 * time.Second),
		Workers:      200,
		PrefixOutput: "tcp_",
		ShuffleIPs:   true,
	}
}

type HTTPConfig struct {
	Host          string     `toml:"host"`
	ServerName    string     `toml:"server_name"`
	Port          int        `toml:"port"`
	Timeout       DurationMS `toml:"timeout"`
	Workers       int        `toml:"workers"`
	PrefixOutput  string     `toml:"prefix_output"`
	TLSValidation bool       `toml:"tls_validation"`
	MinTLSVersion string     `toml:"min_tls_version"`
	MaxTLSVersion string     `toml:"max_tls_version"`
	Protocol      string     `toml:"protocol"`
	ShuffleIPs    bool       `toml:"shuffle_ips"`
}

func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host:          "example.com",
		ServerName:    "",
		Port:          443,
		Timeout:       NewDurationMS(4 * time.Second),
		Workers:       50,
		PrefixOutput:  "http_",
		Protocol:      "https",
		TLSValidation: true,
		MinTLSVersion: "tls1.1",
		MaxTLSVersion: "tls1.3",
		ShuffleIPs:    true,
	}
}

type XrayConfig struct {
	Timeout              DurationMS       `toml:"timeout"`
	Workers              int              `toml:"workers"`
	ShuffleIPs           bool             `toml:"shuffle_ips"`
	ConnectivityTestType ConnectivityTest `toml:"connectivity_test_type"`
	DownloadSpeed        int              `toml:"download_speed"`
	UploadSpeed          int              `toml:"upload_speed"`
	PrefixOutput         string           `toml:"prefix_output"`
}

func DefaultXrayConfig() *XrayConfig {
	return &XrayConfig{
		Timeout:              NewDurationMS(6 * time.Second),
		Workers:              32,
		ShuffleIPs:           true,
		ConnectivityTestType: ConnectivityOnly,
		DownloadSpeed:        100,
		UploadSpeed:          50,
		PrefixOutput:         "xray_",
	}
}

////////////////////////////////////////////////////////////////
// General + Writer Config
////////////////////////////////////////////////////////////////

type GeneralConfig struct {
	StatusInterval DurationMS `toml:"status_interval"`
	StopAfterFound int        `toml:"stop_after_found"`
	MaxIPsToTest   int        `toml:"max_ips_to_test"`
	Verbose        bool       `toml:"verbose"`
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		StatusInterval: NewDurationMS(1 * time.Second),
		StopAfterFound: 0,
		MaxIPsToTest:   0,
		Verbose:        false,
	}
}

type WriterConfig struct {
	DeltaFlushInterval DurationMS `toml:"delta_flush_interval"`
	MergeFlushInterval DurationMS `toml:"merge_flush_interval"`
	ChanSize           int        `toml:"chan_size"`
	BufferSize         int        `toml:"buffer_size"`
}

func DefaultWriterConfig() *WriterConfig {
	return &WriterConfig{
		DeltaFlushInterval: NewDurationMS(2 * time.Second),
		MergeFlushInterval: NewDurationMS(5 * time.Second),
		ChanSize:           1024,
		BufferSize:         4096,
	}
}

////////////////////////////////////////////////////////////////
// Root Config
////////////////////////////////////////////////////////////////

type ScannerConfig struct {
	General *GeneralConfig
	Writer  *WriterConfig
	ICMP    *ICMPConfig
	TCP     *TCPConfig
	HTTP    *HTTPConfig
	Xray    *XrayConfig
}

////////////////////////////////////////////////////////////////
// Paths
////////////////////////////////////////////////////////////////

const (
	settingsDir = "settings"

	icmpFile    = "icmp_settings.toml"
	tcpFile     = "tcp_settings.toml"
	httpFile    = "http_settings.toml"
	xrayFile    = "xray_settings.toml"
	generalFile = "general_settings.toml"
	writerFile  = "writer_settings.toml"
)

func configPath(filename string) (string, error) {
	base, err := filemanager.GetCurrentPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, settingsDir, filename), nil
}

////////////////////////////////////////////////////////////////
// Generic Load / Save Helpers
////////////////////////////////////////////////////////////////

func loadConfig[T any](filename string, cfg *T, def *T, set func(*T)) error {
	path, err := configPath(filename)
	if err != nil {
		return err
	}

	if err := filemanager.GetTOMLFileOrDefault(path, cfg, def); err != nil {
		return fmt.Errorf("load config %s: %w", filename, err)
	}

	set(cfg)
	return nil
}

func saveConfig[T any](filename string, cfg *T, set func(*T)) error {
	path, err := configPath(filename)
	if err != nil {
		return err
	}

	if err := filemanager.WriteTOMLFile(path, cfg); err != nil {
		return fmt.Errorf("save config %s: %w", filename, err)
	}

	set(cfg)
	return nil
}

////////////////////////////////////////////////////////////////
// Public Load / Save
////////////////////////////////////////////////////////////////

func LoadGeneralConfig() error {
	cfg := &GeneralConfig{}
	return loadConfig(generalFile, cfg, DefaultGeneralConfig(), setGeneral)
}

func LoadWriterConfig() error {
	cfg := &WriterConfig{}
	return loadConfig(writerFile, cfg, DefaultWriterConfig(), setWriter)
}

func LoadICMPConfig() error {
	cfg := &ICMPConfig{}
	return loadConfig(icmpFile, cfg, DefaultICMPConfig(), setICMP)
}

func LoadTCPConfig() error {
	cfg := &TCPConfig{}
	return loadConfig(tcpFile, cfg, DefaultTCPConfig(), setTCP)
}

func LoadHTTPConfig() error {
	cfg := &HTTPConfig{}
	return loadConfig(httpFile, cfg, DefaultHTTPConfig(), setHTTP)
}

func LoadXrayConfig() error {
	cfg := &XrayConfig{}
	return loadConfig(xrayFile, cfg, DefaultXrayConfig(), setXray)
}

func SaveWriterConfig(cfg *WriterConfig) error {
	return saveConfig(writerFile, cfg, setWriter)
}

func SaveGeneralConfig(cfg *GeneralConfig) error {
	return saveConfig(generalFile, cfg, setGeneral)
}

func SaveICMPConfig(cfg *ICMPConfig) error {
	return saveConfig(icmpFile, cfg, setICMP)
}

func SaveTCPConfig(cfg *TCPConfig) error {
	return saveConfig(tcpFile, cfg, setTCP)
}

func SaveHTTPConfig(cfg *HTTPConfig) error {
	return saveConfig(httpFile, cfg, setHTTP)
}

func SaveXrayConfig(cfg *XrayConfig) error {
	return saveConfig(xrayFile, cfg, setXray)
}

////////////////////////////////////////////////////////////////
// Initialization
////////////////////////////////////////////////////////////////

// Init loads all scanner configuration files.
func Init() error {
	loaders := []func() error{
		LoadGeneralConfig,
		LoadICMPConfig,
		LoadTCPConfig,
		LoadHTTPConfig,
		LoadXrayConfig,
		LoadWriterConfig,
	}

	for _, load := range loaders {
		if err := load(); err != nil {
			return err
		}
	}

	return nil
}
