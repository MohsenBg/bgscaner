package xray

import (
	"bgscan/internal/core/filemanager"
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

const fallbackXrayPath = "assets/xray/xray"

// XrayProcess represents a running Xray instance.
//
// It wraps an exec.Cmd and provides helper methods for controlling
// the process lifecycle.
type XrayProcess struct {
	cmd *exec.Cmd
}

// findXrayBinary attempts to locate the Xray executable.
//
// Resolution order:
//
//  1. System PATH using exec.LookPath("xray")
//  2. Fallback path bundled with bgscan
//
// Returns the resolved binary path or an error if no executable
// could be found.
func findXrayBinary() (string, error) {
	path, err := exec.LookPath("xray")
	if err == nil {
		return path, nil
	}

	if filemanager.CheckFileExists(fallbackXrayPath) {
		return fallbackXrayPath, nil
	}

	return "", fmt.Errorf("xray binary not found")
}

// IsXrayExists checks whether a usable Xray binary is available.
//
// It attempts to locate the executable using the same logic
// used by the runtime launcher.
func IsXrayExists() bool {
	_, err := findXrayBinary()
	return err == nil
}

// ValidateConfig verifies that a configuration file is valid
// by executing:
//
//	xray -c <config> --test
//
// If the configuration is invalid, the error returned will contain
// the full output produced by Xray to help diagnose the issue.
func ValidateConfig(configPath string) error {

	if !filemanager.CheckFileExists(configPath) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	xrayBin, err := findXrayBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(xrayBin, "-c", configPath, "--test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xray config validation failed:\n%s", string(output))
	}

	return nil
}

// StartXray launches an Xray process using the provided configuration.
//
// The process is started asynchronously and returned as an XrayProcess
// instance so the caller can manage its lifecycle.
//
// The provided context controls the lifetime of the process. If the
// context is canceled, the Xray process will be terminated automatically.
func StartXray(ctx context.Context, configPath string) (*XrayProcess, error) {

	if !filemanager.CheckFileExists(configPath) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	xrayBin, err := findXrayBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, xrayBin, "-c", configPath)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start xray: %w", err)
	}

	return &XrayProcess{cmd: cmd}, nil
}

// StopGracefully attempts to terminate the process using SIGTERM
// and waits for the process to exit.
//
// This allows Xray to clean up resources before shutting down.
func (x *XrayProcess) StopGracefully() error {
	if x.cmd == nil || x.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	if err := x.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	return x.cmd.Wait()
}

// Kill forcefully terminates the process immediately.
//
// This sends a SIGKILL signal and should only be used if
// graceful shutdown fails.
func (x *XrayProcess) Kill() error {
	if x.cmd == nil || x.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	return x.cmd.Process.Kill()
}

// Wait blocks until the Xray process exits.
//
// It returns the exit status of the process.
func (x *XrayProcess) Wait() error {
	if x.cmd == nil {
		return fmt.Errorf("process not initialized")
	}

	return x.cmd.Wait()
}
