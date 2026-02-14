package errx

import "errors"

var (
	ErrMissingIPFile      = errors.New("IP file path is required")
	ErrMissingICMPConfig  = errors.New("ICMP config is required for ICMP scan")
	ErrMissingTCPConfig   = errors.New("TCP config is required for TCP scan")
	ErrMissingHTTPConfig  = errors.New("HTTP config is required for HTTP scan")
	ErrMissingXrayConfig  = errors.New("Xray config is required for Xray scan")
	ErrMissingHost        = errors.New("host is required for HTTP scan")
	ErrInvalidPort        = errors.New("port must be between 1 and 65535")
	ErrInvalidWorkerCount = errors.New("worker count must be at least 1")
	ErrInvalidScanType    = errors.New("invalid scan type")
)
