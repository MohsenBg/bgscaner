package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Process represents a started external process. It manages background waiting,
// graceful shutdown, forceful termination, and exposes a Wait method similar to
// exec.Cmd.
//
// The process is waited on exactly once, protected by waitOnce, and waitDone is
// closed when Wait() completes, regardless of success or failure.
type Process struct {
	cmd      *exec.Cmd
	waitOnce sync.Once
	waitErr  error
	waitDone chan struct{}
}

// Start resolves the binary path if necessary, sets platform‑specific
// SysProcAttr values (via setSysProcAttr), starts the process, and returns a
// Process wrapper.
//
// The given context controls the lifetime of the process through
// exec.CommandContext. When the context is cancelled, the process receives its
// OS‑specific termination signal automatically.
//
// Example:
//
//	p, err := process.Start(ctx, "/usr/bin/somebinary", "--flag")
//	if err != nil {
//	    return err
//	}
//	defer p.Kill()
//
//	if err := p.Wait(); err != nil {
//	    return err
//	}
func Start(ctx context.Context, binary string, args ...string) (*Process, error) {
	// Ensure binary path is absolute.
	if !filepath.IsAbs(binary) {
		abs, err := filepath.Abs(binary)
		if err != nil {
			return nil, fmt.Errorf("resolve binary path: %w", err)
		}
		binary = abs
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}

	p := &Process{
		cmd:      cmd,
		waitDone: make(chan struct{}),
	}

	go p.backgroundWait()
	return p, nil
}

// backgroundWait waits for the process to exit.
// It runs in a goroutine immediately after Start() and ensures Wait() semantics
// complete exactly once.
func (p *Process) backgroundWait() {
	p.waitOnce.Do(func() {
		if p.cmd != nil && p.cmd.Process != nil {
			p.waitErr = p.cmd.Wait()
		}
		close(p.waitDone)
	})
}

// StopGracefully attempts to shut down the process by sending the OS‑specific
// termination signal via signalTerminate(). If the process does not exit within
// the provided timeout, the process is force‑killed.
//
// Returns the wait result from the process, or Kill()’s error if forced.
func (p *Process) StopGracefully(timeout time.Duration) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	signalTerminate(p.cmd)

	select {
	case <-p.waitDone:
		return p.waitErr
	case <-time.After(timeout):
		return p.Kill()
	}
}

// Kill forcefully terminates the process using killProcess() and waits for it
// to complete. The returned error is the process wait result.
func (p *Process) Kill() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	killProcess(p.cmd)
	<-p.waitDone

	return p.waitErr
}

// Wait blocks until the process exits and returns its exit status.
func (p *Process) Wait() error {
	<-p.waitDone
	return p.waitErr
}

// FindBinaryInPaths searches for a binary within the provided directories.
// If not found, it falls back to exec.LookPath. On Windows, ".exe" is appended
// automatically unless already present.
//
// Returns the absolute path if the binary is located.
func FindBinaryInPaths(binary string, dirs []string) (string, error) {
	if runtime.GOOS == "windows" && !strings.HasSuffix(binary, ".exe") {
		binary += ".exe"
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(dir, binary)

		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			abs, err := filepath.Abs(fullPath)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}

	// Try system PATH.
	if path, err := exec.LookPath(binary); err == nil {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	return "", fmt.Errorf("binary '%s' not found", binary)
}

// EnsureExecutable verifies that a Unix file has executable permissions.
// On Windows, this is a no‑op. If the file is not executable, it is chmod’d to
// mode 0755. The original permission bits (except exec flags) are preserved.
func EnsureExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	mode := info.Mode()
	if mode&0111 != 0 {
		return nil
	}

	// Add execute bits without removing existing permissions.
	return os.Chmod(path, mode|0755)
}
