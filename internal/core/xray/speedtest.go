package xray

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// cloudflareTraceURL is used for latency measurements.
	// The endpoint returns a small response and is globally
	// distributed, making it useful for quick RTT estimation.
	cloudflareTraceURL = "https://speed.cloudflare.com/cdn-cgi/trace"

	// cloudflareDownURL is used for controlled download tests.
	// The %d placeholder specifies the number of bytes to download.
	cloudflareDownURL = "https://speed.cloudflare.com/__down?bytes=%d"

	// cloudflareUpURL is used for upload speed tests.
	cloudflareUpURL = "https://speed.cloudflare.com/__up"
)

// newHTTPClientWithSocks5 creates an HTTP client that routes all
// requests through a local SOCKS5 proxy.
//
// The proxy is expected to be running on 127.0.0.1:port, typically
// provided by a locally running Xray instance started by the scanner.
//
// The returned client uses the provided timeout for the entire request
// lifecycle, including connection establishment and response reading.
func newHTTPClientWithSocks5(timeout time.Duration, port int) (*http.Client, error) {
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %w", err)
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}, nil
}

// MeasureLatency measures round-trip latency through the given SOCKS5 proxy.
//
// The function performs an HTTP GET request to Cloudflare's trace endpoint
// via the proxy and returns the elapsed time between request initiation
// and response receipt.
//
// This measurement includes:
//
//   - TCP connection establishment
//   - TLS handshake
//   - proxy routing overhead
//   - server response latency
//
// It provides a practical estimate of end-to-end latency for traffic
// routed through the Xray proxy.
func MeasureLatency(timeout time.Duration, proxyPort int) (time.Duration, error) {
	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	resp, err := client.Get(cloudflareTraceURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return time.Since(start), nil
}

// MeasureDownloadDuration measures how long it takes to download
// a specified number of bytes through the SOCKS5 proxy.
//
// The test downloads data from Cloudflare's controlled speed endpoint.
// The response body is fully consumed to ensure the entire payload
// is transferred before the duration is calculated.
func MeasureDownloadDuration(timeout time.Duration, bytesSize int64, proxyPort int) (time.Duration, error) {
	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf(cloudflareDownURL, bytesSize)

	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// MeasureUploadDuration measures how long it takes to upload
// a specified number of bytes through the SOCKS5 proxy.
//
// The function generates an in-memory payload and sends it
// to Cloudflare's upload endpoint. The returned duration
// represents the total time required to transmit the data
// and receive the server response.
func MeasureUploadDuration(timeout time.Duration, bytesSize int64, proxyPort int) (time.Duration, error) {
	client, err := newHTTPClientWithSocks5(timeout, proxyPort)
	if err != nil {
		return 0, err
	}

	data := bytes.Repeat([]byte("A"), int(bytesSize))

	start := time.Now()
	resp, err := client.Post(cloudflareUpURL, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return time.Since(start), nil
}
